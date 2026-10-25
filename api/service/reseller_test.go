package service_test

import (
	"encoding/json"
	"errors"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/service"
	"strings"
	"testing"
)

func resellerSettings(name string, ifaceID uint) service.ResellerSettings {
	return service.ResellerSettings{Name: name, Mode: model.BillingPrepaid, InterfaceIDs: []uint{ifaceID}, DefaultInterfaceID: &ifaceID}
}

func TestResellerOwnershipSettingsAndFinancialHistory(t *testing.T) {
	f := newMultiFixture(t)
	pa := f.createPeer(t, f.a.ID, "a")
	pb := f.createPeer(t, f.b.ID, "b")
	actor := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}
	manager := service.NewResellers(f.db, f.peers)
	var ia, ib model.Interface
	f.db.Where("server_id = ?", f.a.ID).First(&ia)
	f.db.Where("server_id = ?", f.b.ID).First(&ib)
	settings := resellerSettings("One", ia.ID)
	one, err := manager.Create(actor, service.CreateReseller{ResellerSettings: settings, Username: "one", Password: "initial-password-one"})
	if err != nil {
		t.Fatal(err)
	}
	two, err := manager.Create(actor, service.CreateReseller{ResellerSettings: resellerSettings("Two", ia.ID), Username: "two", Password: "initial-password-two"})
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.TransferPeer(actor, pa.ID, &one.ID, "assign"); err != nil {
		t.Fatal(err)
	}
	if err := manager.TransferPeer(actor, pb.ID, &one.ID, "wrong resource"); !errors.Is(err, service.ErrResourceForbidden) {
		t.Fatalf("resource bypass: %v", err)
	}
	peers, err := manager.PeerList(actor, one.ID)
	if err != nil || len(*peers) != 1 || (*peers)[0].UUID != pa.UUID {
		t.Fatalf("scope: %+v %v", peers, err)
	}
	billing := service.NewBilling(f.db)
	if err := billing.Adjust(actor, one.ID, "credit", 1234567, "initial credit"); err != nil {
		t.Fatal(err)
	}
	settings.Mode = model.BillingPostpaid
	updated, err := manager.Update(actor, one.ID, settings)
	if err != nil || updated.BalanceRial != "1234567" {
		t.Fatalf("balance lost: %+v %v", updated, err)
	}
	settings.Mode = model.BillingPrepaid
	settings.InterfaceIDs = []uint{ia.ID, ib.ID}
	settings.DefaultInterfaceID = &ib.ID
	if _, err := manager.Update(actor, one.ID, settings); !errors.Is(err, service.ErrRouterMigrationRequired) {
		t.Fatalf("unsafe migration: %v", err)
	}
	unchanged, err := manager.Get(actor, one.ID)
	if err != nil || unchanged.Mode != model.BillingPostpaid || *unchanged.DefaultInterfaceID != ia.ID {
		t.Fatal("failed update was not atomic")
	}
	if _, err := billing.ReserveCharge(actor, one.ID, "extension", &pa.ID, service.BillingExtendExpiry, service.BillingGiB, 30); err != nil {
		t.Fatal(err)
	}
	if err := manager.TransferPeer(actor, pa.ID, &two.ID, "pending transfer"); !errors.Is(err, service.ErrPendingPeerOperation) {
		t.Fatalf("pending operation bypass: %v", err)
	}
	if err := billing.ReleaseCharge(actor, one.ID, "extension", "no router operation performed"); err != nil {
		t.Fatal(err)
	}
	if err := manager.TransferPeer(actor, pa.ID, &two.ID, "approved transfer"); err != nil {
		t.Fatal(err)
	}
	var events []model.PeerOwnershipEvent
	f.db.Where("peer_id = ?", pa.ID).Find(&events)
	found := false
	for _, event := range events {
		if event.PreviousResellerID != nil && *event.PreviousResellerID == one.ID && event.ResellerID != nil && *event.ResellerID == two.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("previous owner lost in audit")
	}
	account, err := billing.Account(actor, one.ID)
	if err != nil || account.BalanceRial != 1234567 {
		t.Fatal("transfer changed money")
	}
	encoded, _ := json.Marshal(one)
	for _, secret := range []string{"password", "private_key", "totp", "initial-password"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatal("credential leak")
		}
	}
	var before, after int64
	f.db.Model(&model.Reseller{}).Count(&before)
	if _, err := manager.Create(actor, service.CreateReseller{ResellerSettings: resellerSettings("duplicate", ia.ID), Username: "one", Password: "initial-password-duplicate"}); err == nil {
		t.Fatal("duplicate accepted")
	}
	f.db.Model(&model.Reseller{}).Count(&after)
	if before != after {
		t.Fatal("orphan tenant after failed creation")
	}
	member := service.Principal{AccountID: 9, Role: model.RoleReseller, ResellerID: &one.ID}
	if _, err := manager.List(member); !errors.Is(err, service.ErrRoleForbidden) {
		t.Fatal("reseller listed accounts")
	}
	if err := manager.TransferPeer(member, pa.ID, &one.ID, "forged"); !errors.Is(err, service.ErrRoleForbidden) {
		t.Fatal("reseller transferred ownership")
	}
}

func TestResellerLimitsRejectTransfers(t *testing.T) {
	f := newMultiFixture(t)
	peer := f.createPeer(t, f.a.ID, "a")
	actor := service.Principal{AccountID: 1, Role: model.RoleSuperadmin}
	manager := service.NewResellers(f.db, f.peers)
	var iface model.Interface
	f.db.Where("server_id = ?", f.a.ID).First(&iface)
	zero, small := int64(0), int64(1)
	for index, kind := range []string{"active", "peer-quota", "total-quota", "duration"} {
		settings := resellerSettings(kind, iface.ID)
		switch index {
		case 0:
			settings.MaxActivePeers = &zero
		case 1:
			settings.MaxPeerTrafficBytes = &small
		case 2:
			settings.MaxTotalTrafficBytes = &small
		case 3:
			settings.MaxDurationDays = &small
		}
		created, err := manager.Create(actor, service.CreateReseller{ResellerSettings: settings, Username: kind, Password: "test-password-long"})
		if err != nil {
			t.Fatal(err)
		}
		if err := manager.TransferPeer(actor, peer.ID, &created.ID, "over limit"); !errors.Is(err, service.ErrResellerLimit) {
			t.Fatalf("%s: %v", kind, err)
		}
	}
	var check model.Peer
	f.db.First(&check, peer.ID)
	if check.ResellerID != nil {
		t.Fatal("failed transfer changed ownership")
	}
}
