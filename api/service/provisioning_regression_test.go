package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

func TestProvisioningAcceptsRouterOSISOSchedulerDates(t *testing.T) {
	f, existing, actor, _, purchases := purchaseFixture(t)

	bridge := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := httptest.NewRecorder()

		f.mockA.ServeHTTP(rec, r)

		body := rec.Body.Bytes()

		if r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/system/scheduler") && rec.Code == 200 {
			var value any

			if json.Unmarshal(body, &value) == nil {
				normalize := func(row map[string]any) {
					if raw, ok := row["start-date"].(string); ok {
						if day, err := time.Parse("Jan/02/2006", raw); err == nil {
							row["start-date"] = day.Format("2006-01-02")
						}
					}
				}

				switch v := value.(type) {
				case map[string]any:
					normalize(v)

				case []any:
					for _, x := range v {
						if row, ok := x.(map[string]any); ok {
							normalize(row)
						}
					}
				}

				body, _ = json.Marshal(value)
			}
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(rec.Code)

		_, _ = w.Write(body)
	}))

	defer bridge.Close()

	_, port, _ := net.SplitHostPort(strings.TrimPrefix(bridge.URL, "http://"))

	number, _ := strconv.Atoi(port)

	if err := f.db.Model(&f.a).Update("api_port", number).Error; err != nil {
		t.Fatal(err)
	}

	var iface model.Interface

	f.db.Where("server_id = ?", existing.ServerID).First(&iface)

	op, err := purchases.Create(actor, *actor.ResellerID, "iso-date", service.ResellerPeerCreate{Name: "ISO router", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(service.BillingGiB, 10), DurationDays: "30"})

	if err != nil {
		t.Fatal(err)
	}

	if err = purchases.Reconcile(op.ID); err != nil {
		t.Fatalf("equivalent RouterOS scheduler date prevents provisioning: %v", err)
	}
}

func TestCreateWorkerActivatesPrepaidAtZeroBalance(t *testing.T) {
	f, existing, actor, billing, purchases := purchaseFixture(t)

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := billing.Adjust(admin, *actor.ResellerID, "exact-credit", -9_000_000, "reserve exact plan credit"); err != nil {
		t.Fatal(err)
	}

	var iface model.Interface

	f.db.Where("server_id = ?", existing.ServerID).First(&iface)

	op, err := purchases.Create(actor, *actor.ResellerID, "zero-balance", service.ResellerPeerCreate{Name: "Paid zero wallet", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(10*service.BillingGiB, 10), DurationDays: "30"})

	if err != nil {
		t.Fatal(err)
	}

	purchases.Run()

	var peer model.Peer

	f.db.First(&peer, op.PeerID)

	if peer.ProvisioningBlocked || peer.Disabled || peer.SyncedDisabled == nil || *peer.SyncedDisabled {
		t.Fatalf("worker left eligible paid peer inactive: provisioning=%v disabled=%v", peer.ProvisioningBlocked, peer.Disabled)
	}

	account, err := billing.Account(actor, *actor.ResellerID)

	if err != nil || account.BalanceRial != 0 || account.BillingSuspended {
		t.Fatal("zero balance changed prepaid account eligibility", err)
	}
}

func TestProvisioningFailureRetryAndPrepaidDeposit(t *testing.T) {
	f, existing, actor, billing, purchases := purchaseFixture(t)

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := billing.Adjust(admin, *actor.ResellerID, "empty-wallet", -10_000_000, "zero available credit"); err != nil {
		t.Fatal(err)
	}

	var iface model.Interface

	f.db.Where("server_id = ?", existing.ServerID).First(&iface)

	req := service.ResellerPeerCreate{Name: "Retry after deposit", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(service.BillingGiB, 10), DurationDays: "30"}

	if _, err := purchases.Create(actor, *actor.ResellerID, "before-deposit", req); !errors.Is(err, service.ErrInsufficientCredit) {
		t.Fatal("insufficient credit accepted", err)
	}

	payments := service.NewPayments(billing, t.TempDir())

	var imageData bytes.Buffer

	if err := png.Encode(&imageData, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatal(err)
	}

	receipt, err := payments.Submit(actor, *actor.ResellerID, service.ReceiptInput{IdempotencyKey: "deposit", AmountRial: "100000", PaidAt: time.Now().UTC().Format(time.RFC3339), TransactionReference: "activation-test-bank"}, bytes.NewReader(imageData.Bytes()))

	if err != nil {
		t.Fatal(err)
	}

	account, _ := billing.Account(actor, *actor.ResellerID)

	if account.BalanceRial != 0 {
		t.Fatal("unapproved deposit changed wallet")
	}

	if err = payments.Approve(admin, receipt.ID); err != nil {
		t.Fatal(err)
	}

	op, err := purchases.Create(actor, *actor.ResellerID, "after-deposit", req)

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodGet, "/interface/wireguard", 1)

	// Fail at the actual interface identity lookup, before finalizing the charge.
	f.mockA.FailNext(http.MethodGet, "/interface/wireguard/"+iface.InterfaceID, 1)

	purchases.Run()

	var pending model.PeerPurchase

	f.db.First(&pending, "id = ?", op.ID)

	if pending.State != "pending" || pending.Attempts != 1 || !strings.Contains(pending.LastError, "interface:") {
		t.Fatalf("failure not persisted: %+v", pending)
	}

	account, _ = billing.Account(actor, *actor.ResellerID)

	if account.BalanceRial != 100000 || account.ReservedRial != 100000 {
		t.Fatal("failed provisioning finalized charge")
	}

	manager := service.NewResellers(f.db, f.peers)

	if err = manager.RetryPeer(actor, *actor.ResellerID, op.PeerID); !errors.Is(err, service.ErrRoleForbidden) {
		t.Fatal("reseller could trigger admin retry", err)
	}

	f.db.Model(&pending).Update("lease_until", time.Now().Unix()+120)

	if err = manager.RetryPeer(admin, *actor.ResellerID, op.PeerID); !errors.Is(err, service.ErrPendingPeerOperation) {
		t.Fatal("retry stole active worker lease", err)
	}

	f.db.Model(&pending).Update("lease_until", time.Now().Unix()-1)

	if err = manager.RetryPeer(admin, *actor.ResellerID, op.PeerID); err != nil {
		t.Fatal(err)
	}

	// A fresh worker models restart recovery; backoff was released, not the hold.
	service.NewPeerPurchases(billing, f.peers).Run()

	var peer model.Peer

	f.db.First(&peer, op.PeerID)

	if peer.ProvisioningBlocked || peer.Disabled {
		t.Fatal("retry did not activate eligible peer")
	}

	account, _ = billing.Account(actor, *actor.ResellerID)

	if account.BalanceRial != 0 || account.ReservedRial != 0 || account.BillingSuspended {
		t.Fatal("recovery changed prepaid financial invariants")
	}

	if err = payments.Approve(admin, receipt.ID); err != nil {
		t.Fatal(err)
	}

	if err = purchases.Reconcile(op.ID); err != nil {
		t.Fatal(err)
	}

	var charges int64

	f.db.Model(&model.LedgerEntry{}).Where("reseller_id = ? AND kind = ?", *actor.ResellerID, "charge").Count(&charges)

	if charges != 1 {
		t.Fatal("retry duplicated ledger charge")
	}
}

func TestProvisioningPreservesIndependentRestrictions(t *testing.T) {
	for _, restriction := range []string{"admin", "manual", "expired", "quota"} {
		t.Run(restriction, func(t *testing.T) {
			f, existing, actor, _, purchases := purchaseFixture(t)

			var iface model.Interface

			f.db.Where("server_id = ?", existing.ServerID).First(&iface)

			op, err := purchases.Create(actor, *actor.ResellerID, "restricted", service.ResellerPeerCreate{Name: "Restricted setup", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(service.BillingGiB, 10), DurationDays: "30"})

			if err != nil {
				t.Fatal(err)
			}

			updates := map[string]any{}

			switch restriction {
			case "admin":
				f.db.Model(&model.Reseller{}).Where("id = ?", *actor.ResellerID).Update("admin_suspended", true)

			case "manual":
				updates["manual_disabled"] = true

			case "expired":
				updates["expire_time"] = "2020-01-01"

			case "quota":
				updates["download_usage"] = service.BillingGiB
			}

			if len(updates) > 0 {
				if err = f.db.Model(&model.Peer{}).Where("id = ?", op.PeerID).Updates(updates).Error; err != nil {
					t.Fatal(err)
				}
			}

			purchases.Run()

			var peer model.Peer

			f.db.First(&peer, op.PeerID)

			if peer.ProvisioningBlocked || !peer.Disabled || peer.SyncedDisabled == nil || !*peer.SyncedDisabled {
				t.Fatal("provisioning bypassed independent restriction")
			}
		})
	}
}

func TestNewPrepaidResellerStartsActiveWithZeroCredit(t *testing.T) {
	f := newMultiFixture(t)

	peer := f.createPeer(t, f.a.ID, "admin-owned")

	var iface model.Interface

	f.db.Where("server_id = ?", peer.ServerID).First(&iface)

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	manager := service.NewResellers(f.db, f.peers)

	reseller, err := manager.Create(admin, schema.CreateReseller{ResellerSettings: resellerSettings("New prepaid", iface.ID), Username: "new-prepaid", Password: "valid-test-password"})

	if err != nil {
		t.Fatal(err)
	}

	if !reseller.IsActive || reseller.BillingSuspended || reseller.AdminSuspended || reseller.CreditRial != "0" {
		t.Fatal("empty prepaid wallet prevented account activation")
	}

	state, err := service.NewBilling(f.db).State(admin, reseller.ID)

	if err != nil || state.Status != "active" || state.BillingSuspended {
		t.Fatal("new prepaid account has an incorrect financial restriction", err)
	}
}
