package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
)

func TestResellerBandwidthIsFreeScopedVerifiedAndRetryable(t *testing.T) {
	f, peer, actor, billing, purchases := purchaseFixture(t)

	other := f.createPeer(t, f.b.ID, "other-router")

	before, err := billing.State(actor, *actor.ResellerID)

	if err != nil {
		t.Fatal(err)
	}

	down, up := "24M", "2.5M"

	req := service.PeerMutationRequest{Action: "set_bandwidth", DownloadBandwidth: &down, UploadBandwidth: &up}

	if _, err = purchases.Mutate(actor, *actor.ResellerID, other.ID, "foreign-speed", req); err == nil {
		t.Fatal("cross-tenant mutation allowed")
	}

	op, err := purchases.Mutate(actor, *actor.ResellerID, peer.ID, "speed", req)

	if err != nil {
		t.Fatal(err)
	}

	f.mockA.FailAfterNext(http.MethodPatch, common.QueuePath+"/"+*peer.QueueID, 1)

	if err = purchases.Reconcile(op.ID); err == nil {
		t.Fatal("lost response was not reported")
	}

	var pending model.Peer

	f.db.First(&pending, peer.ID)

	if pending.DownloadBandwidth != nil || pending.PurchaseID == nil {
		t.Fatal("unverified speed was finalized")
	}

	f.db.Model(&model.PeerPurchase{}).Where("id = ?", op.ID).Update("next_attempt_at", 0)

	if err = purchases.Reconcile(op.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&peer, peer.ID)

	if peer.DownloadBandwidth == nil || *peer.DownloadBandwidth != down || peer.UploadBandwidth == nil || *peer.UploadBandwidth != up {
		t.Fatal("speed was not saved")
	}

	router, _ := f.adaptor.ForServer(peer.ServerID)

	queue, err := router.FetchSimpleQueue(context.Background(), *peer.QueueID)

	if err != nil || queue.MaxLimit == nil || *queue.MaxLimit != "24M/2.5M" {
		t.Fatal("queue speed mismatch", err)
	}

	otherRouter, _ := f.adaptor.ForServer(other.ServerID)

	untouched, err := otherRouter.FetchSimpleQueue(context.Background(), *other.QueueID)

	if err != nil || untouched.MaxLimit == nil || *untouched.MaxLimit != "0/0" {
		t.Fatal("other router modified", err)
	}

	if replay, err := purchases.Mutate(actor, *actor.ResellerID, peer.ID, "speed", req); err != nil || replay.ID != op.ID {
		t.Fatal("retry duplicated operation", err)
	}

	req.DownloadBandwidth = nil

	if _, err := purchases.Mutate(actor, *actor.ResellerID, peer.ID, "speed", req); !errors.Is(err, service.ErrIdempotencyConflict) {
		t.Fatal("changed payload reused key", err)
	}

	after, err := billing.State(actor, *actor.ResellerID)

	if err != nil || before != after {
		t.Fatal("free speed change altered billing", err)
	}

	unlimited, err := purchases.Mutate(actor, *actor.ResellerID, peer.ID, "unlimited-speed", service.PeerMutationRequest{Action: "set_bandwidth"})

	if err != nil {
		t.Fatal(err)
	}

	if err = purchases.Reconcile(unlimited.ID); err != nil {
		t.Fatal(err)
	}

	queue, err = router.FetchSimpleQueue(context.Background(), *peer.QueueID)

	if err != nil || queue.MaxLimit == nil || *queue.MaxLimit != "0/0" {
		t.Fatal("unlimited reset failed", err)
	}
}

func TestResellerBandwidthValidationAndCreation(t *testing.T) {
	f, peer, actor, _, purchases := purchaseFixture(t)

	for _, value := range []string{"", "0M", "-1M", "1/2", "10M;delete", "1e3M", "1001G"} {
		_, err := purchases.Mutate(actor, *actor.ResellerID, peer.ID, "invalid-"+value, service.PeerMutationRequest{Action: "set_bandwidth", DownloadBandwidth: &value})

		if !errors.Is(err, service.ErrInvalidBillingQuantity) {
			t.Fatalf("accepted %q: %v", value, err)
		}
	}

	var iface model.Interface

	f.db.Where("server_id = ?", peer.ServerID).First(&iface)

	speed := "16M"

	op, err := purchases.Create(actor, *actor.ResellerID, "limited-create", service.ResellerPeerCreate{Name: "limited", InterfaceID: iface.ID, QuotaBytes: "1073741824", DurationDays: "30", DownloadBandwidth: &speed})

	if err != nil {
		t.Fatal(err)
	}

	if err = purchases.Reconcile(op.ID); err != nil {
		t.Fatal(err)
	}

	var created model.Peer

	f.db.First(&created, op.PeerID)

	router, _ := f.adaptor.ForServer(created.ServerID)

	queue, err := router.FetchSimpleQueue(context.Background(), *created.QueueID)

	if err != nil || queue.MaxLimit == nil || *queue.MaxLimit != "16M/0" || op.AmountRial != "100000" {
		t.Fatal("creation speed/price mismatch", err)
	}
}
