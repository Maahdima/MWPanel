package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestFreePeerMutationsPreserveWalletAndShareIdentity(t *testing.T) {
	for _, action := range []string{"rename", "set_address", "regenerate_keys", "set_manual", "set_share", "reduce_quota", "delete"} {
		t.Run(action, func(t *testing.T) {
			f, peer, actor, billing, operations := purchaseFixture(t)
			enabled := true
			req := service.PeerMutationRequest{Action: action}
			switch action {
			case "rename":
				req.Value = "renamed"
			case "set_address":
				req.Value = "10.42.0.3/32"
			case "set_manual":
				enabled = false
				req.Enabled = &enabled
			case "set_share":
				req.Enabled = &enabled
			case "reduce_quota":
				req.Value = "1073741824"
			}
			oldUUID, oldPublic := peer.UUID, peer.PublicKey
			op, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "free", req)
			if err != nil {
				t.Fatal(err)
			}
			if op.AmountRial != "0" {
				t.Fatal("free operation charged")
			}
			if err = operations.Reconcile(op.ID); err != nil {
				t.Fatal(err)
			}
			account, _ := billing.Account(actor, *actor.ResellerID)
			if account.BalanceRial != 10_000_000 || account.ReservedRial != 0 {
				t.Fatal("free operation changed credit")
			}
			var charges int64
			f.db.Model(&model.BillingCharge{}).Count(&charges)
			if charges != 0 {
				t.Fatal("free operation created a financial charge")
			}
			replay, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "free", req)
			if err != nil || replay.ID != op.ID || replay.State != "applied" {
				t.Fatal("replay failed", err)
			}
			router, _ := f.adaptor.ForServer(peer.ServerID)
			if action == "delete" {
				var count int64
				f.db.Model(&model.Peer{}).Where("id = ?", peer.ID).Count(&count)
				if count != 0 {
					t.Fatal("peer not deleted")
				}
				list, e := router.FetchWgPeers(context.Background())
				if e != nil || len(list) != 0 {
					t.Fatal("router peer not removed")
				}
				return
			}
			f.db.First(&peer, peer.ID)
			if peer.UUID != oldUUID || peer.PurchaseID != nil {
				t.Fatal("share identity or operation lock changed")
			}
			switch action {
			case "rename":
				if peer.Name != "renamed" {
					t.Fatal("name not changed")
				}
			case "set_address":
				if peer.AllowedAddress != req.Value {
					t.Fatal("address not changed")
				}
			case "regenerate_keys":
				if peer.PublicKey == oldPublic {
					t.Fatal("keys not regenerated")
				}
			case "set_manual":
				if !peer.ManualDisabled {
					t.Fatal("manual intent not recorded")
				}
			case "set_share":
				if !peer.IsShared {
					t.Fatal("share not enabled")
				}
			case "reduce_quota":
				if *peer.TrafficLimit != 1<<30 {
					t.Fatal("quota reduction failed")
				}
			}
		})
	}
}

func TestFreeMutationCannotBypassLimitsOrBillingRestrictions(t *testing.T) {
	f, peer, actor, billing, operations := purchaseFixture(t)
	if _, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "increase", service.PeerMutationRequest{Action: "reduce_quota", Value: "999999999999"}); !errors.Is(err, service.ErrBilledOperationRequired) {
		t.Fatal("free quota increase accepted", err)
	}
	if _, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "address", service.PeerMutationRequest{Action: "set_address", Value: "192.0.2.1/32"}); !errors.Is(err, service.ErrResourceForbidden) {
		t.Fatal("unauthorized address", err)
	}
	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}
	if err := service.NewResellers(f.db, f.peers).SetSuspension(admin, *actor.ResellerID, true, "test"); err != nil {
		t.Fatal(err)
	}
	enabled := true
	if _, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "enable", service.PeerMutationRequest{Action: "set_manual", Enabled: &enabled}); !errors.Is(err, service.ErrPeerRestricted) {
		t.Fatal("suspension bypass", err)
	}
	// Account restrictions still permit a free rename and keep router disabled.
	op, err := operations.Mutate(admin, *actor.ResellerID, peer.ID, "rename", service.PeerMutationRequest{Action: "rename", Value: "renamed while suspended"})
	if err != nil {
		t.Fatal(err)
	}
	if err = operations.Reconcile(op.ID); err != nil {
		t.Fatal(err)
	}
	if err = service.NewPeerReconciler(f.db, f.adaptor, billing).ReconcilePeer(peer.ID); err != nil {
		t.Fatal(err)
	}
	router, _ := f.adaptor.ForServer(peer.ServerID)
	remote, err := router.FetchWgPeer(context.Background(), peer.PeerID)
	if err != nil || remote.Disabled != "true" {
		t.Fatal("free operation lifted suspension")
	}
}

func TestFreeMutationAmbiguousResponsePreservesGeneratedKeys(t *testing.T) {
	f, peer, actor, _, operations := purchaseFixture(t)
	op, err := operations.Mutate(actor, *actor.ResellerID, peer.ID, "keys", service.PeerMutationRequest{Action: "regenerate_keys"})
	if err != nil {
		t.Fatal(err)
	}
	f.mockA.FailAfterNext(http.MethodPatch, "/interface/wireguard/peers/"+peer.PeerID, 1)
	if err = operations.Reconcile(op.ID); err == nil {
		t.Fatal("expected ambiguous response")
	}
	router, _ := f.adaptor.ForServer(peer.ServerID)
	changed, err := router.FetchWgPeer(context.Background(), peer.PeerID)
	if err != nil || changed.PublicKey == peer.PublicKey {
		t.Fatal("injection did not mutate router")
	}
	f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)
	operations.Run()
	f.db.First(&peer, peer.ID)
	if peer.PurchaseID != nil || peer.PublicKey != changed.PublicKey {
		t.Fatal("retry generated a different key or failed to resume")
	}
}
