package service_test

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func purchaseFixture(t *testing.T) (multiFixture, model.Peer, service.Principal, *service.Billing, *service.PeerPurchases) {
	t.Helper()

	f := newMultiFixture(t)

	peer := f.createPeer(t, f.a.ID, "paid")

	tenant := model.Reseller{Name: "purchaser", IsActive: true}

	if err := f.db.Create(&tenant).Error; err != nil {
		t.Fatal(err)
	}

	var iface model.Interface

	f.db.Where("server_id = ?", f.a.ID).First(&iface)

	if err := f.db.Create(&model.ResellerResource{ResellerID: tenant.ID, InterfaceID: iface.ID}).Error; err != nil {
		t.Fatal(err)
	}

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	billing := service.NewBilling(f.db)

	if err := billing.CreateAccount(admin, tenant.ID, model.BillingPrepaid); err != nil {
		t.Fatal(err)
	}

	if err := billing.Adjust(admin, tenant.ID, "initial-credit", 10_000_000, "test credit"); err != nil {
		t.Fatal(err)
	}

	if err := f.db.Model(&peer).Updates(map[string]any{"reseller_id": tenant.ID, "configured_duration_days": 30}).Error; err != nil {
		t.Fatal(err)
	}

	f.db.First(&peer, peer.ID)

	actor := service.Principal{AccountID: 2, Role: model.RoleReseller, ResellerID: &tenant.ID}

	return f, peer, actor, billing, service.NewPeerPurchases(billing, f.peers)
}

func TestPeerPurchasesExactRulesAndReplay(t *testing.T) {
	for _, test := range []struct {
		name  string
		req   service.PeerPurchaseRequest
		want  string
		quota int64
		reset bool
		days  int
	}{
		{"increase", service.PeerPurchaseRequest{Operation: service.BillingAddTraffic, Bytes: strconv.FormatInt(5*service.BillingGiB, 10)}, "500000", 15 * service.BillingGiB, false, 0},
		{"extend15", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"}, "1000000", 10 * service.BillingGiB, false, 15},
		{"extend60", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "60"}, "1000000", 10 * service.BillingGiB, false, 60},
		{"reset", service.PeerPurchaseRequest{Operation: service.BillingResetUsage}, "400000", 10 * service.BillingGiB, true, 0},
		{"full", service.PeerPurchaseRequest{Operation: service.BillingFullReset}, "1000000", 10 * service.BillingGiB, true, -1},
	} {
		t.Run(test.name, func(t *testing.T) {
			f, peer, actor, billing, purchases := purchaseFixture(t)

			if err := f.db.Model(&peer).Update("download_usage", 4*service.BillingGiB).Error; err != nil {
				t.Fatal(err)
			}

			expiry := *peer.ExpireTime

			op, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "operation", test.req)

			if err != nil {
				t.Fatal(err)
			}

			if op.AmountRial != test.want || op.State != "pending" {
				t.Fatalf("unexpected snapshot %+v", op)
			}

			a, _ := billing.Account(actor, *actor.ResellerID)

			if a.BalanceRial != 10_000_000 {
				t.Fatal("pending operation finalized a charge")
			}

			if err = purchases.Reconcile(op.ID); err != nil {
				t.Fatal(err)
			}

			replay, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "operation", test.req)

			if err != nil || replay.ID != op.ID || replay.State != "applied" {
				t.Fatalf("replay %+v %v", replay, err)
			}

			if err = purchases.Reconcile(op.ID); err != nil {
				t.Fatal(err)
			}

			f.db.First(&peer, peer.ID)

			if peer.PurchaseID != nil || *peer.TrafficLimit != test.quota {
				t.Fatal("purchase not applied")
			}

			if test.reset && peer.DownloadUsage != 0 {
				t.Fatal("usage not reset")
			}

			if test.days == 0 && *peer.ExpireTime != expiry {
				t.Fatal("reset/increase changed expiry")
			}

			if test.days > 0 {
				end, _ := time.Parse("2006-01-02", expiry)

				if *peer.ExpireTime != end.AddDate(0, 0, test.days).Format("2006-01-02") {
					t.Fatal("extension did not add days")
				}
			}

			var count int64

			f.db.Model(&model.LedgerEntry{}).Where("kind = ?", "charge").Count(&count)

			if count != 1 {
				t.Fatal("duplicate charge")
			}

			a, _ = billing.Account(actor, *actor.ResellerID)

			want, _ := strconv.ParseInt(test.want, 10, 64)

			if a.BalanceRial != 10_000_000-want || a.ReservedRial != 0 {
				t.Fatal("incorrect posted balance")
			}
		})
	}
}

func TestPeerPurchaseOutageRestartsWithoutFinalCharge(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	op, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "outage", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"})

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodGet, "/interface/wireguard/peers/"+peer.PeerID, 1)

	if err = purchases.Reconcile(op.ID); err == nil {
		t.Fatal("expected outage")
	}

	a, _ := billing.Account(actor, *actor.ResellerID)

	if a.BalanceRial != 10_000_000 || a.ReservedRial != 1_000_000 {
		t.Fatal("outage finalized or lost hold")
	}

	if err = f.peers.DeletePeer(peer.ID); !errors.Is(err, service.ErrPendingPeerOperation) {
		t.Fatalf("delete bypass: %v", err)
	}

	f.calculator.ResetPeerUsages()

	f.db.First(&peer, peer.ID)

	if peer.PurchaseID == nil {
		t.Fatal("reset lost operation")
	}

	f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)

	restart := service.NewPeerPurchases(billing, f.peers)

	restart.Run()

	f.db.First(&peer, peer.ID)

	if peer.PurchaseID != nil {
		t.Fatal("restart did not finish")
	}

	router, _ := f.adaptor.ForServer(peer.ServerID)

	schedule, e := router.FetchScheduler(context.Background(), *peer.SchedulerID)

	if e != nil || schedule.StartDate == nil {
		t.Fatal("missing verified scheduler")
	}
}

func TestPeerPurchaseIsolationConcurrentWalletAndChangedReplay(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	other := actor

	id := *actor.ResellerID + 1

	other.ResellerID = &id

	if _, err := purchases.Begin(other, *actor.ResellerID, peer.ID, "idor", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"}); !errors.Is(err, service.ErrRoleForbidden) {
		t.Fatal(err)
	}

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := billing.Adjust(admin, *actor.ResellerID, "reduce-credit", -9_000_000, "test balance"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup

	results := make(chan error, 8)

	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "same", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"})

			results <- err
		}()
	}

	wg.Wait()

	close(results)

	success := 0

	for err := range results {
		if err == nil {
			success++
		} else if !errors.Is(err, service.ErrPendingPeerOperation) {
			t.Fatal(err)
		}
	}

	if success == 0 {
		t.Fatal("no successful reservation")
	}

	var count int64

	f.db.Model(&model.PeerPurchase{}).Count(&count)

	if count != 1 {
		t.Fatal("duplicate operation")
	}

	if _, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "same", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "60"}); !errors.Is(err, service.ErrIdempotencyConflict) {
		t.Fatal("changed replay accepted", err)
	}

	a, _ := billing.Account(actor, *actor.ResellerID)

	if a.ReservedRial != 1_000_000 || a.BalanceRial != 1_000_000 {
		t.Fatal("wallet overspent")
	}
}

func TestPeerPurchaseAmbiguousSchedulerPatchRechecksOnRestart(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	op, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "retry", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "30"})

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodPatch, "/system/scheduler/"+*peer.SchedulerID, 1)

	if err = purchases.Reconcile(op.ID); err == nil {
		t.Fatal("expected patch failure")
	}

	a, _ := billing.Account(actor, *actor.ResellerID)

	if a.BalanceRial != 10_000_000 {
		t.Fatal("unsuccessful operation billed")
	}

	f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)

	if err = purchases.Reconcile(op.ID); err != nil {
		t.Fatal(err)
	}
}

func TestPeerPurchaseRouterSuccessDatabaseFailureRemainsRecoverable(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	op, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "db-failure", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"})

	if err != nil {
		t.Fatal(err)
	}

	const callback = "test_reject_posting"

	if err = f.db.Callback().Create().Before("gorm:create").Register(callback, func(tx *gorm.DB) {
		if tx.Statement.Table == "ledger_entries" {
			tx.AddError(errors.New("injected posting failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}

	if err = purchases.Reconcile(op.ID); err == nil {
		t.Fatal("expected posting failure")
	}

	f.db.Callback().Create().Remove(callback)

	a, _ := billing.Account(actor, *actor.ResellerID)

	if a.BalanceRial != 10_000_000 || a.ReservedRial != 1_000_000 {
		t.Fatal("failed posting changed balance")
	}

	f.db.First(&peer, peer.ID)

	if peer.PurchaseID == nil {
		t.Fatal("operation lost")
	}

	router, _ := f.adaptor.ForServer(peer.ServerID)

	schedule, err := router.FetchScheduler(context.Background(), *peer.SchedulerID)

	if err != nil {
		t.Fatal(err)
	}

	original, _ := time.Parse("2006-01-02", *peer.ExpireTime)

	if schedule.StartDate == nil || *schedule.StartDate != strings.ToLower(original.AddDate(0, 0, 15).Format("Jan/02/2006")) {
		t.Fatalf("router mutation did not precede injected database error: %+v", schedule)
	}

	if err = billing.UpdateInstructions(service.Principal{AccountID: 1, Role: model.RoleSuperadmin}, service.PaymentInstructions{RateRial: "200000"}); err != nil {
		t.Fatal(err)
	}

	f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)

	service.NewPeerPurchases(billing, f.peers).Run()

	a, _ = billing.Account(actor, *actor.ResellerID)

	if a.BalanceRial != 9_000_000 || a.ReservedRial != 0 {
		t.Fatal("restart recalculated historical rate or duplicated charge")
	}
}

func TestPendingPurchaseDoesNotPreventAdministrativeSuspension(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	_, err := purchases.Begin(actor, *actor.ResellerID, peer.ID, "pending-block", service.PeerPurchaseRequest{Operation: service.BillingExtendExpiry, Days: "15"})

	if err != nil {
		t.Fatal(err)
	}

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err = service.NewResellers(f.db, f.peers).SetSuspension(admin, *actor.ResellerID, true, "test suspension"); err != nil {
		t.Fatal(err)
	}

	reconciler := service.NewPeerReconciler(f.db, f.adaptor, billing)

	if err = reconciler.ReconcilePeer(peer.ID); err != nil {
		t.Fatal(err)
	}

	router, _ := f.adaptor.ForServer(peer.ServerID)

	state, err := router.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || state.Disabled != "true" {
		t.Fatal("pending purchase bypassed suspension")
	}
}
