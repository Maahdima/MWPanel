package dataservice

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestLegacyRouterMigrationRequiresVerifiedSource(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	a := model.Server{Name: "a", IPAddress: "192.0.2.1"}
	b := model.Server{Name: "b", IPAddress: "192.0.2.2"}
	if err := db.Create(&a).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&b).Error; err != nil {
		t.Fatal(err)
	}
	peer := model.Peer{UUID: "stable", PeerID: "*1", Name: "legacy", AllowedAddress: "10.0.0.2", PrivateKey: "keep-key", DownloadUsage: 123}
	if err := db.Create(&peer).Error; err != nil {
		t.Fatal(err)
	}
	if err := MigrateRouterOwnership(db, ""); err == nil {
		t.Fatal("ambiguous router was guessed")
	}
	var check model.Peer
	if err := db.First(&check, peer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if check.ServerID != 0 {
		t.Fatal("ambiguous migration changed data")
	}
	if err := MigrateRouterOwnership(db, "999999"); err == nil {
		t.Fatal("missing source accepted")
	}
	for i := 0; i < 2; i++ {
		if err := MigrateRouterOwnership(db, strconv.Itoa(int(b.ID))); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.First(&check, peer.ID).Error; err != nil {
		t.Fatal(err)
	}
	if check.ServerID != b.ID || check.UUID != "stable" || check.PrivateKey != "keep-key" || check.DownloadUsage != 123 || check.ResellerID != nil {
		t.Fatal("migration lost legacy state")
	}
}

func TestRouterMigrationReplacesLegacyUniqueIndexes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "indexes.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	for _, query := range []string{
		"CREATE UNIQUE INDEX idx_peers_peer_id ON peers(peer_id)",
		"CREATE UNIQUE INDEX idx_peers_allowed_address ON peers(allowed_address)",
		"CREATE UNIQUE INDEX idx_interfaces_interface_id ON interfaces(interface_id)",
		"CREATE UNIQUE INDEX idx_interfaces_name ON interfaces(name)",
		"CREATE UNIQUE INDEX idx_ip_pools_name ON ip_pools(name)",
		"CREATE UNIQUE INDEX idx_ip_pools_start_ip ON ip_pools(start_ip)",
		"CREATE UNIQUE INDEX idx_ip_pools_end_ip ON ip_pools(end_ip)",
	} {
		if err := db.Exec(query).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := MigrateRouterOwnership(db, ""); err != nil {
		t.Fatal(err)
	}
	for _, serverID := range []uint{1, 2} {
		peer := model.Peer{ServerID: serverID, UUID: "peer-" + strconv.Itoa(int(serverID)), PeerID: "*1", Name: "same", AllowedAddress: "10.0.0.2"}
		if err := db.Create(&peer).Error; err != nil {
			t.Fatalf("legacy global peer index survived: %v", err)
		}
		iface := model.Interface{ServerID: serverID, InterfaceID: "*1", Name: "wg0"}
		if err := db.Create(&iface).Error; err != nil {
			t.Fatalf("legacy global interface index survived: %v", err)
		}
		pool := model.IPPool{ServerID: serverID, InterfaceID: iface.ID, Name: "pool", StartIP: "10.0.0.2/32", EndIP: "10.0.0.10/32"}
		if err := db.Create(&pool).Error; err != nil {
			t.Fatalf("legacy global pool index survived: %v", err)
		}
	}
	duplicate := model.Peer{ServerID: 1, UUID: "different", PeerID: "*1", Name: "duplicate", AllowedAddress: "10.0.0.3"}
	if err := db.Create(&duplicate).Error; err == nil {
		t.Fatal("router-local uniqueness was lost")
	}
}
