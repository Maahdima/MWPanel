package mikrotik

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/maahdima/mwp/api/common"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/internal/mockrouteros"
	"github.com/maahdima/mwp/api/utils/httphelper"
	"gorm.io/gorm"
)

func routingDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "routing.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Server{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	t.Cleanup(func() { sqlDB.Close() })
	return db
}

func mockServerRecord(t *testing.T, db *gorm.DB, name string) (*mockrouteros.Router, model.Server) {
	t.Helper()
	mock := mockrouteros.New(name, "testuser", "testpassword")
	server := httptest.NewServer(mock)
	t.Cleanup(server.Close)
	u, _ := url.Parse(server.URL)
	host, port, _ := net.SplitHostPort(u.Host)
	n, _ := strconv.Atoi(port)
	record := model.Server{Name: name, IPAddress: host, APIPort: n, Username: "testuser", Password: "testpassword", IsActive: true}
	// Distinct loopback host spellings respect the legacy globally unique address.
	var existing int64
	if err := db.Model(&model.Server{}).Count(&existing).Error; err != nil {
		t.Fatal(err)
	}
	if existing > 0 {
		record.IPAddress = "localhost"
	}
	if err := db.Create(&record).Error; err != nil {
		t.Fatal(err)
	}
	return mock, record
}

func TestBoundRoutersRemainIsolatedWithCollidingIDs(t *testing.T) {
	db := routingDB(t)
	mockA, a := mockServerRecord(t, db, "router-a")
	_, b := mockServerRecord(t, db, "router-b")
	base := NewAdaptor(common.NewMwpClient(db))
	aa, err := base.ForServer(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	bb, err := base.ForServer(b.ID)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pa, err := aa.CreateWgPeer(ctx, WireGuardPeer{Name: "tenant-a", PublicKey: "a"})
	if err != nil {
		t.Fatal(err)
	}
	pb, err := bb.CreateWgPeer(ctx, WireGuardPeer{Name: "tenant-b", PublicKey: "b"})
	if err != nil {
		t.Fatal(err)
	}
	if pa.ID != pb.ID {
		t.Fatalf("fixture must collide: %q != %q", pa.ID, pb.ID)
	}
	var wg sync.WaitGroup
	for _, tc := range []struct {
		router *Adaptor
		name   string
	}{{aa, "tenant-a"}, {bb, "tenant-b"}} {
		wg.Add(1)
		go func(router *Adaptor, name string) {
			defer wg.Done()
			for i := 0; i < 20; i++ {
				p, err := router.FetchWgPeer(ctx, "*1")
				if err != nil || p.Name != name {
					t.Errorf("router isolation: %v %+v, want %s", err, p, name)
				}
			}
		}(tc.router, tc.name)
	}
	wg.Wait()
	mockA.FailNext(http.MethodPatch, common.WGPeerPath+"/*1", 1)
	if _, err := aa.UpdateWgPeer(ctx, pa.ID, WireGuardPeer{Disabled: "true"}); err == nil {
		t.Fatal("injected failure was ignored")
	}
	if _, err := aa.UpdateWgPeer(ctx, pa.ID, WireGuardPeer{Disabled: "true"}); err != nil {
		t.Fatal(err)
	}
	unchanged, err := bb.FetchWgPeer(ctx, pb.ID)
	if err != nil || unchanged.Disabled != "false" {
		t.Fatalf("router b was changed: %+v %v", unchanged, err)
	}
	if err := aa.DeleteWgPeer(ctx, pa.ID); err != nil {
		t.Fatal(err)
	}
	if err := aa.DeleteWgPeer(ctx, pa.ID); err != nil {
		t.Fatalf("delete retry: %v", err)
	}
	if _, err := bb.FetchWgPeer(ctx, pb.ID); err != nil {
		t.Fatalf("deleted peer on wrong router: %v", err)
	}
}

func TestRouterSelectionDoesNotFallback(t *testing.T) {
	db := routingDB(t)
	_, a := mockServerRecord(t, db, "first")
	_, _ = mockServerRecord(t, db, "second")
	registry := common.NewMwpClient(db)
	base := NewAdaptor(registry)
	for i := 0; i < 20; i++ {
		got, err := base.FetchDeviceIdentity(context.Background())
		if err != nil || got.Name != "first" {
			t.Fatalf("unstable default: %+v %v", got, err)
		}
	}
	if _, err := base.ForServer(0); err == nil {
		t.Fatal("zero server accepted")
	}
	if _, err := base.ForServer(999999); err == nil {
		t.Fatal("missing server fell back")
	}
	if err := db.Model(&model.Server{}).Where("id = ?", a.ID).Update("is_active", false).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := base.ForServer(a.ID); err == nil {
		t.Fatal("inactive server fell back")
	}
}

func TestUnconfiguredAdaptorReturnsError(t *testing.T) {
	adaptor := NewAdaptor(common.NewMwpClient(routingDB(t)))
	if _, err := adaptor.FetchWgPeers(context.Background()); err == nil {
		t.Fatal("expected error without configured router")
	}
}

func TestMockRequiresCredentials(t *testing.T) {
	server := httptest.NewServer(mockrouteros.New("router", "u", "p"))
	defer server.Close()
	client, err := httphelper.NewClient(httphelper.Config{BaseURL: server.URL + "/rest"})
	if err != nil {
		t.Fatal(err)
	}
	if err := client.Get(context.Background(), common.DeviceIdentityPath, nil); err == nil {
		t.Fatal("mock accepted missing credentials")
	}
}
