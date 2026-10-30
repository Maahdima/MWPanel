package service_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/service"
)

func extraPeer(t *testing.T, f multiFixture, interfaceID uint, name, address string) model.Peer {
	t.Helper()

	credentials, err := f.peers.GetPeerCredentials()

	if err != nil {
		t.Fatal(err)
	}

	quota, expiry := "10", time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	created, err := f.peers.CreatePeer(&schema.CreatePeerRequest{Name: name, InterfaceId: interfaceID, PrivateKey: credentials.PrivateKey, PublicKey: credentials.PublicKey, AllowedAddress: address, Endpoint: "vpn.example.test", TrafficLimit: &quota, ExpireTime: &expiry})

	if err != nil {
		t.Fatal(err)
	}

	var peer model.Peer

	if err := f.db.First(&peer, created.Id).Error; err != nil {
		t.Fatal(err)
	}

	return peer
}

func TestBillingRecoveryPreservesMixedPeerRestrictions(t *testing.T) {
	f := newMultiFixture(t)

	normal := f.createPeer(t, f.a.ID, "normal")

	var iface model.Interface

	f.db.Where("server_id = ?", f.a.ID).First(&iface)

	manual := extraPeer(t, f, iface.ID, "manual", "10.42.0.3/32")

	expired := extraPeer(t, f, iface.ID, "expired", "10.42.0.4/32")

	quota := extraPeer(t, f, iface.ID, "quota", "10.42.0.5/32")

	if err := f.peers.TogglePeerStatus(manual.ID); err != nil {
		t.Fatal(err)
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	if err := f.db.Model(&model.Peer{}).Where("id = ?", expired.ID).Update("expire_time", yesterday).Error; err != nil {
		t.Fatal(err)
	}

	if err := f.db.Model(&model.Peer{}).Where("id = ?", quota.ID).Update("download_usage", *quota.TrafficLimit).Error; err != nil {
		t.Fatal(err)
	}

	tenant := model.Reseller{Name: "billing tenant"}

	if err := f.db.Create(&tenant).Error; err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	clock := now.Add(-36 * 24 * time.Hour)

	billing := service.NewBilling(f.db).WithClock(func() time.Time { return clock })

	admin := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}

	if err := billing.CreateAccount(admin, tenant.ID, model.BillingPostpaid); err != nil {
		t.Fatal(err)
	}

	if _, err := billing.ReserveCharge(admin, tenant.ID, "charge", nil, service.BillingCreate, service.BillingGiB, 30); err != nil {
		t.Fatal(err)
	}

	if err := billing.CommitCharge(admin, tenant.ID, "charge", nil); err != nil {
		t.Fatal(err)
	}

	ids := []uint{normal.ID, manual.ID, expired.ID, quota.ID}

	if err := f.db.Model(&model.Peer{}).Where("id IN ?", ids).Update("reseller_id", tenant.ID).Error; err != nil {
		t.Fatal(err)
	}

	clock = now

	if _, err := billing.ReconcileAccount(tenant.ID); err != nil {
		t.Fatal(err)
	}

	reconciler := service.NewPeerReconciler(f.db, f.adaptor, billing)

	router, _ := f.adaptor.ForServer(f.a.ID)

	for _, id := range ids {
		if err := reconciler.ReconcilePeer(id); err != nil {
			t.Fatal(err)
		}
	}

	for _, peer := range []model.Peer{normal, manual, expired, quota} {
		observed, err := router.FetchWgPeer(context.Background(), peer.PeerID)

		if err != nil || observed.Disabled != "true" {
			t.Fatal("billing suspension failed")
		}
	}

	if err := billing.Adjust(admin, tenant.ID, "settlement", 100000, "approved settlement adjustment"); err != nil {
		t.Fatal(err)
	}

	if _, err := billing.ReconcileAccount(tenant.ID); err != nil {
		t.Fatal(err)
	}

	for _, id := range ids {
		if err := reconciler.ReconcilePeer(id); err != nil {
			t.Fatal(err)
		}
	}

	for _, peer := range []model.Peer{normal, manual, expired, quota} {
		observed, err := router.FetchWgPeer(context.Background(), peer.PeerID)

		if err != nil {
			t.Fatal(err)
		}

		want := "true"

		if peer.ID == normal.ID {
			want = "false"
		}

		if observed.Disabled != want {
			t.Fatalf("recovery changed unrelated restriction for %s: %s", peer.Name, observed.Disabled)
		}
	}

	var stillExpired model.Peer

	f.db.First(&stillExpired, expired.ID)

	if stillExpired.ExpireTime == nil || *stillExpired.ExpireTime != yesterday {
		t.Fatal("suspension compensated expiry days")
	}

	if err := f.peers.TogglePeerStatus(quota.ID); !errors.Is(err, service.ErrPeerRestricted) {
		t.Fatalf("manual endpoint bypassed quota restriction: %v", err)
	}
}

func TestPeerSynchronizationRecoversAcrossFailureAndRestart(t *testing.T) {
	f := newMultiFixture(t)

	a := f.createPeer(t, f.a.ID, "a")

	b := f.createPeer(t, f.b.ID, "b")

	if err := f.db.Model(&model.Peer{}).Where("id IN ?", []uint{a.ID, b.ID}).Update("manual_disabled", true).Error; err != nil {
		t.Fatal(err)
	}

	f.mockA.FailNext(http.MethodPatch, common.WGPeerPath+"/"+a.PeerID, 1)

	first := service.NewPeerReconciler(f.db, f.adaptor, service.NewBilling(f.db))

	if err := first.ReconcilePeer(a.ID); err == nil {
		t.Fatal("router failure not recorded")
	}

	if err := first.ReconcilePeer(b.ID); err != nil {
		t.Fatalf("another router was blocked: %v", err)
	}

	var pending model.Peer

	f.db.First(&pending, a.ID)

	if pending.SyncError == "" || pending.SyncNextAttemptAt == 0 || pending.SyncAttempts != 1 {
		t.Fatal("retry state was not durable")
	}

	if err := f.db.Model(&model.Peer{}).Where("id = ?", a.ID).Update("sync_next_attempt_at", 0).Error; err != nil {
		t.Fatal(err)
	}

	restarted := service.NewPeerReconciler(f.db, f.adaptor, service.NewBilling(f.db))

	if err := restarted.ReconcilePeer(a.ID); err != nil {
		t.Fatal(err)
	}

	f.db.First(&pending, a.ID)

	if pending.SyncError != "" || pending.SyncedDisabled == nil || !*pending.SyncedDisabled || pending.AppliedRevision != pending.RestrictionRevision {
		t.Fatal("restarted worker did not converge")
	}
}

func TestReconciliationRefusesReusedRouterObjectIdentity(t *testing.T) {
	f := newMultiFixture(t)

	peer := f.createPeer(t, f.a.ID, "a")

	router, _ := f.adaptor.ForServer(f.a.ID)

	if _, err := router.UpdateWgPeer(context.Background(), peer.PeerID, mikrotik.WireGuardPeer{PublicKey: "another-customer-key"}); err != nil {
		t.Fatal(err)
	}

	if err := f.db.Model(&model.Peer{}).Where("id = ?", peer.ID).Update("manual_disabled", true).Error; err != nil {
		t.Fatal(err)
	}

	reconciler := service.NewPeerReconciler(f.db, f.adaptor, service.NewBilling(f.db))

	if err := reconciler.ReconcilePeer(peer.ID); !errors.Is(err, service.ErrRouterPeerIdentity) {
		t.Fatalf("identity mismatch ignored: %v", err)
	}

	observed, err := router.FetchWgPeer(context.Background(), peer.PeerID)

	if err != nil || observed.Disabled != "false" {
		t.Fatal("worker modified a different customer's router object")
	}
}
