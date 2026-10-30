package service_test

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestPaidCreatePricingAndOwnedRouterProvisioning(t *testing.T) {
	for _, days := range []int{15, 30, 45, 60} {
		t.Run(strconv.Itoa(days), func(t *testing.T) {
			f, existing, actor, billing, purchases := purchaseFixture(t)

			var iface model.Interface

			f.db.Where("server_id = ?", existing.ServerID).First(&iface)

			req := service.ResellerPeerCreate{Name: "new paid peer", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(10*service.BillingGiB, 10), DurationDays: strconv.Itoa(days)}

			op, err := purchases.Create(actor, *actor.ResellerID, "new-peer", req)

			if err != nil {
				t.Fatal(err)
			}

			want := int64(1_000_000 * max(days, 30) / 30)

			if op.AmountRial != strconv.FormatInt(want, 10) {
				t.Fatal("wrong creation price", op)
			}

			var peer model.Peer

			f.db.First(&peer, op.PeerID)

			stableUUID := peer.UUID

			if !peer.ProvisioningBlocked || !peer.Disabled || peer.ResellerID == nil || *peer.ResellerID != *actor.ResellerID {
				t.Fatal("unowned or enabled pending peer")
			}

			if err = purchases.Reconcile(op.ID); err != nil {
				t.Fatal(err)
			}

			f.db.First(&peer, op.PeerID)

			if peer.ProvisioningBlocked || peer.PurchaseID != nil || peer.UUID != stableUUID || peer.ConfiguredDurationDays != int64(days) {
				t.Fatal("configuration not finalized")
			}

			router, _ := f.adaptor.ForServer(existing.ServerID)

			observed, err := router.FetchWgPeer(context.Background(), peer.PeerID)

			if err != nil || observed.Disabled != "false" {
				t.Fatal("paid peer was not activated through restriction reconciliation")
			}

			if err = service.NewPeerReconciler(f.db, f.adaptor, billing).ReconcilePeer(peer.ID); err != nil {
				t.Fatal(err)
			}

			observed, err = router.FetchWgPeer(context.Background(), peer.PeerID)

			if err != nil || observed.Disabled != "false" {
				t.Fatal("paid peer failed activation")
			}

			replay, err := purchases.Create(actor, *actor.ResellerID, "new-peer", req)

			if err != nil || replay.ID != op.ID || replay.PeerID != peer.ID {
				t.Fatal("create replay duplicated peer", err)
			}

			account, _ := billing.Account(actor, *actor.ResellerID)

			if account.BalanceRial != 10_000_000-want || account.ReservedRial != 0 {
				t.Fatal("wrong wallet balance")
			}

			other, _ := f.adaptor.ForServer(f.b.ID)

			list, err := other.FetchWgPeers(context.Background())

			if err != nil || len(list) != 0 {
				t.Fatal("wrong router modified")
			}
		})
	}
}

func TestPaidCreateAmbiguousResponseResumesWithoutDuplicateObjects(t *testing.T) {
	for _, path := range []string{"/interface/wireguard/peers", "/queue/simple", "/system/scheduler"} {
		t.Run(path, func(t *testing.T) {
			f, existing, actor, billing, purchases := purchaseFixture(t)

			var iface model.Interface

			f.db.Where("server_id = ?", existing.ServerID).First(&iface)

			req := service.ResellerPeerCreate{Name: "resumable", InterfaceID: iface.ID, QuotaBytes: strconv.FormatInt(service.BillingGiB, 10), DurationDays: "30"}

			op, err := purchases.Create(actor, *actor.ResellerID, "ambiguous", req)

			if err != nil {
				t.Fatal(err)
			}

			f.mockA.FailAfterNext(http.MethodPut, path, 1)

			if err = purchases.Reconcile(op.ID); err == nil {
				t.Fatal("expected ambiguous response")
			}

			account, _ := billing.Account(actor, *actor.ResellerID)

			if account.BalanceRial != 10_000_000 || account.ReservedRial != 100000 {
				t.Fatal("uncertain creation finalized charge")
			}

			f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)

			service.NewPeerPurchases(billing, f.peers).Run()

			var peer model.Peer

			f.db.First(&peer, op.PeerID)

			if peer.PurchaseID != nil {
				t.Fatal("did not resume")
			}

			router, _ := f.adaptor.ForServer(existing.ServerID)

			peers, err := router.FetchWgPeers(context.Background())

			if err != nil || len(peers) != 2 {
				t.Fatal("duplicate router peer", len(peers), err)
			}

			queues, err := router.FetchSimpleQueues(context.Background())

			if err != nil || len(queues) != 2 {
				t.Fatal("duplicate queue", len(queues), err)
			}

			schedulers, err := router.FetchSchedulers(context.Background())

			if err != nil || len(schedulers) != 2 {
				t.Fatal("duplicate scheduler", len(schedulers), err)
			}

			account, _ = billing.Account(actor, *actor.ResellerID)

			if account.BalanceRial != 9_900_000 || account.ReservedRial != 0 {
				t.Fatal("resumed operation billed twice")
			}
		})
	}
}

func TestPaidCreateRejectsUnlimitedUnauthorizedAndInsufficientCredit(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	var iface model.Interface

	f.db.Where("server_id = ?", peer.ServerID).First(&iface)

	req := service.ResellerPeerCreate{Name: "new", InterfaceID: iface.ID, QuotaBytes: "1073741824", DurationDays: "30"}

	invalid := req

	invalid.QuotaBytes = "0"

	if _, err := purchases.Create(actor, *actor.ResellerID, "unlimited", invalid); !errors.Is(err, service.ErrInvalidBillingQuantity) {
		t.Fatal(err)
	}

	invalid = req

	invalid.DurationDays = "0"

	if _, err := purchases.Create(actor, *actor.ResellerID, "unlimited-days", invalid); !errors.Is(err, service.ErrInvalidBillingQuantity) {
		t.Fatal(err)
	}

	invalid = req

	invalid.InterfaceID = iface.ID + 999

	if _, err := purchases.Create(actor, *actor.ResellerID, "forbidden-interface", invalid); !errors.Is(err, service.ErrResourceForbidden) {
		t.Fatal(err)
	}

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := billing.Adjust(admin, *actor.ResellerID, "remove-credit", -10_000_000, "test"); err != nil {
		t.Fatal(err)
	}

	if _, err := purchases.Create(actor, *actor.ResellerID, "insufficient", req); !errors.Is(err, service.ErrInsufficientCredit) {
		t.Fatal(err)
	}

	var count int64

	f.db.Model(&model.Peer{}).Where("reseller_id = ?", *actor.ResellerID).Count(&count)

	if count != 1 {
		t.Fatal("failed creation leaked a peer")
	}
}
