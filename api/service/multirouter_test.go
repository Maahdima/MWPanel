package service_test

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/maahdima/mwp/api/adaptor/mikrotik"
	jobs "github.com/maahdima/mwp/api/cmd/jobs"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/http/schema"
	"github.com/maahdima/mwp/api/internal/mockrouteros"
	"github.com/maahdima/mwp/api/internal/testdb"
	"github.com/maahdima/mwp/api/service"
	"gorm.io/gorm"
)

type multiFixture struct {
	db           *gorm.DB
	peers        *service.WgPeer
	interfaces   *service.WgInterface
	sync         *service.SyncService
	calculator   *jobs.Calculator
	adaptor      *mikrotik.Adaptor
	a, b         model.Server
	mockA, mockB *mockrouteros.Router
}

func newMultiFixture(t *testing.T) multiFixture {
	t.Helper()
	db := testdb.Open(t, filepath.Join(t.TempDir(), "multi.db"))
	if err := dataservice.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	add := func(name, hostOverride string) (model.Server, *mockrouteros.Router) {
		mock := mockrouteros.New(name, "u", "p")
		server := httptest.NewServer(mock)
		t.Cleanup(server.Close)
		parsed, _ := url.Parse(server.URL)
		host, port, _ := net.SplitHostPort(parsed.Host)
		number, _ := strconv.Atoi(port)
		if hostOverride != "" {
			host = hostOverride
		}
		record := model.Server{Name: name, IPAddress: host, APIPort: number, Username: "u", Password: "p"}
		if err := db.Create(&record).Error; err != nil {
			t.Fatal(err)
		}
		return record, mock
	}
	a, ma := add("router-a", "")
	b, mb := add("router-b", "localhost")
	adaptor := mikrotik.NewAdaptor(common.NewMwpClient(db))
	files := service.NewConfigGenerator(db)
	qr := service.NewQRCodeGenerator(db)
	peers := service.NewWGPeer(db, adaptor, service.NewScheduler(adaptor), service.NewQueue(adaptor), files, qr)
	interfaces := service.NewWgInterface(db, adaptor, peers)
	syncer := service.NewSyncService(db, adaptor, files, qr, peers, interfaces)
	return multiFixture{db, peers, interfaces, syncer, jobs.NewTrafficCalculator(db, adaptor, nil), adaptor, a, b, ma, mb}
}

func (f multiFixture) createPeer(t *testing.T, serverID uint, name string) model.Peer {
	t.Helper()
	iface, err := f.interfaces.CreateInterface(&schema.CreateInterfaceRequest{ServerID: serverID, Name: "wg0", ListenPort: "51820"})
	if err != nil {
		t.Fatal(err)
	}
	poolService := service.NewIPPool(f.db, f.adaptor)
	if _, err := poolService.CreateIPPool(&schema.CreateIPPoolRequest{Name: "pool", InterfaceID: iface.Id, StartIP: "10.42.0.2/32", EndIP: "10.42.0.10/32"}); err != nil {
		t.Fatal(err)
	}
	credentials, err := f.peers.GetPeerCredentials()
	if err != nil {
		t.Fatal(err)
	}
	traffic, expire := "10", time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	peer, err := f.peers.CreatePeer(&schema.CreatePeerRequest{Name: name, InterfaceId: iface.Id, PrivateKey: credentials.PrivateKey, PublicKey: credentials.PublicKey, AllowedAddress: "10.42.0.2/32", Endpoint: name + ".example.test", TrafficLimit: &traffic, ExpireTime: &expire})
	if err != nil {
		t.Fatal(err)
	}
	var stored model.Peer
	if err := f.db.First(&stored, peer.Id).Error; err != nil {
		t.Fatal(err)
	}
	return stored
}

func TestPeerLifecycleAndSyncAreRouterSpecific(t *testing.T) {
	f := newMultiFixture(t)
	a := f.createPeer(t, f.a.ID, "a")
	b := f.createPeer(t, f.b.ID, "b")
	if a.PeerID != b.PeerID || a.Interface != b.Interface || a.AllowedAddress != b.AllowedAddress {
		t.Fatal("test requires colliding local IDs, interface names and addresses")
	}
	if a.ServerID != f.a.ID || b.ServerID != f.b.ID {
		t.Fatal("peer router not persisted")
	}
	if err := f.peers.TogglePeerStatus(a.ID); err != nil {
		t.Fatal(err)
	}
	adaptorB, err := f.adaptor.ForServer(f.b.ID)
	if err != nil {
		t.Fatal(err)
	}
	observed, err := adaptorB.FetchWgPeer(context.Background(), b.PeerID)
	if err != nil || observed.Disabled != "false" {
		t.Fatalf("wrong router toggled: %+v %v", observed, err)
	}
	rows, err := f.peers.GetPeers()
	if err != nil || len(*rows) != 2 {
		t.Fatalf("list merged router-local IDs: %+v %v", rows, err)
	}
	addr, err := f.peers.GetNewPeerAllowedAddress(2)
	if err != nil || addr.AllowedAddress != "10.42.0.3/32" {
		t.Fatalf("cross-router pool collision: %+v %v", addr, err)
	}
	selected, err := f.sync.ForServer(f.a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := selected.SyncPeers(); err != nil {
		t.Fatal(err)
	}
	var kept model.Peer
	if err := f.db.First(&kept, b.ID).Error; err != nil {
		t.Fatal("sync removed another router's peer")
	}
	var synced model.Peer
	if err := f.db.First(&synced, a.ID).Error; err != nil {
		t.Fatal(err)
	}
	if synced.UUID != a.UUID || synced.Endpoint != a.Endpoint {
		t.Fatal("sync changed public UUID or custom endpoint")
	}
	adaptorA, err := f.adaptor.ForServer(f.a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := adaptorA.DeleteWgPeer(context.Background(), a.PeerID); err != nil {
		t.Fatal(err)
	}
	if err := selected.SyncPeers(); err != nil {
		t.Fatal(err)
	}
	if err := f.db.First(&kept, b.ID).Error; err != nil {
		t.Fatal("stale cleanup removed peer on another router")
	}
	if _, err := adaptorB.FetchWgPeer(context.Background(), b.PeerID); err != nil {
		t.Fatal("stale cleanup touched destination router")
	}
}

func TestTrafficResetExpiryAndSessionsUseActualRouters(t *testing.T) {
	f := newMultiFixture(t)
	a := f.createPeer(t, f.a.ID, "a")
	b := f.createPeer(t, f.b.ID, "b")
	aa, _ := f.adaptor.ForServer(f.a.ID)
	bb, _ := f.adaptor.ForServer(f.b.ID)
	handshake := "5s"
	for _, tc := range []struct {
		adaptor *mikrotik.Adaptor
		peer    model.Peer
		tx, rx  string
	}{{aa, a, "101", "21"}, {bb, b, "307", "43"}} {
		if _, err := tc.adaptor.UpdateWgPeer(context.Background(), tc.peer.PeerID, mikrotik.WireGuardPeer{TransferTx: tc.tx, TransferRx: tc.rx, LastHandshake: &handshake}); err != nil {
			t.Fatal(err)
		}
	}
	f.calculator.CalculatePeerTraffic()
	var storedA, storedB model.Peer
	f.db.First(&storedA, a.ID)
	f.db.First(&storedB, b.ID)
	if storedA.DownloadUsage != 101 || storedB.DownloadUsage != 307 {
		t.Fatalf("wrong traffic counters: %d %d", storedA.DownloadUsage, storedB.DownloadUsage)
	}
	if err := f.calculator.ResetPeerUsage(b.ID); err != nil {
		t.Fatal(err)
	}
	f.db.First(&storedA, a.ID)
	f.db.First(&storedB, b.ID)
	if storedA.DownloadUsage != 101 || storedB.DownloadUsage != 0 || storedB.LastTx != 307 {
		t.Fatal("reset used wrong router counters")
	}
	f.calculator.TrackPeerSessions()
	var sessions int64
	f.db.Model(&model.PeerSession{}).Where("disconnected_at IS NULL").Count(&sessions)
	if sessions != 2 {
		t.Fatalf("session IDs collided: %d", sessions)
	}
	f.mockB.FailNext(http.MethodGet, common.WGPeerPath, 1)
	f.calculator.TrackPeerSessions()
	f.db.Model(&model.PeerSession{}).Where("peer_id = ? AND disconnected_at IS NULL", b.ID).Count(&sessions)
	if sessions != 1 {
		t.Fatal("router outage invented a disconnect")
	}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if err := f.db.Model(&model.Peer{}).Where("id = ?", a.ID).Update("expire_time", yesterday).Error; err != nil {
		t.Fatal(err)
	}
	f.calculator.ExpireOverduePeers()
	observedA, err := aa.FetchWgPeer(context.Background(), a.PeerID)
	if err != nil {
		t.Fatal(err)
	}
	observedB, err := bb.FetchWgPeer(context.Background(), b.PeerID)
	if err != nil {
		t.Fatal(err)
	}
	if observedA.Disabled != "true" || observedB.Disabled != "false" {
		t.Fatal("expiry disabled peer on another router")
	}
}

func TestMultipleRouterWritesRequireExplicitSelection(t *testing.T) {
	f := newMultiFixture(t)
	if _, err := f.interfaces.CreateInterface(&schema.CreateInterfaceRequest{Name: "ambiguous", ListenPort: "51820"}); err != service.ErrServerSelectionRequired {
		t.Fatalf("ambiguous create: %v", err)
	}
	if err := f.sync.SyncInterfaces(); err != service.ErrServerSelectionRequired {
		t.Fatalf("ambiguous sync: %v", err)
	}
	if _, err := f.peers.ShopInterface(0); err != service.ErrServerSelectionRequired {
		t.Fatalf("ambiguous shop interface: %v", err)
	}
}

func TestDailyInterfaceTrafficAndBulkResetRemainPartitioned(t *testing.T) {
	f := newMultiFixture(t)
	a := f.createPeer(t, f.a.ID, "a")
	b := f.createPeer(t, f.b.ID, "b")
	f.calculator.CalculateDailyTraffic()
	for _, tc := range []struct {
		server model.Server
		tx, rx string
	}{{f.a, "100", "20"}, {f.b, "300", "40"}} {
		client, err := common.NewServerHTTPClient(tc.server)
		if err != nil {
			t.Fatal(err)
		}
		if err := client.Patch(context.Background(), common.InterfacePath+"/*1", map[string]string{"tx-byte": tc.tx, "rx-byte": tc.rx}, nil); err != nil {
			t.Fatal(err)
		}
	}
	f.calculator.CalculateDailyTraffic()
	var records []model.Traffic
	if err := f.db.Order("interface_id ASC").Find(&records).Error; err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 || records[0].DownloadUsage != 100 || records[1].DownloadUsage != 300 {
		t.Fatalf("daily counters merged: %+v", records)
	}
	if err := f.db.Model(&model.Peer{}).Where("id IN ?", []uint{a.ID, b.ID}).Update("download_usage", 123).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.db.Model(&model.Server{}).Where("id = ?", f.a.ID).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := f.calculator.ResetPeerUsages(); err == nil {
		t.Fatal("bulk reset ignored an unavailable router")
	}
	var check model.Peer
	if err := f.db.First(&check, b.ID).Error; err != nil {
		t.Fatal(err)
	}
	if check.DownloadUsage != 123 {
		t.Fatal("bulk reset partially changed a healthy router's peer")
	}
}
