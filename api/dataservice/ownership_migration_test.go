package dataservice

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

func TestOwnershipMigrationKeepsLegacyAccountsAndPeers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "legacy.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	// A pre-role database, with one existing peer and admin. Optional legacy fields
	// omitted here are added by the same upstream AutoMigrate path.
	statements := []string{
		`CREATE TABLE admins (id integer primary key, username varchar(64) not null, password varchar(128) not null, is_active numeric not null default true)`,
		`INSERT INTO admins(id,username,password) VALUES(1,'original-admin','existing-hash')`,
		`CREATE TABLE peers (id integer primary key, uuid varchar(36) not null, peer_id varchar(255) not null, name varchar(255) not null, private_key varchar(255) not null, public_key varchar(255) not null, interface varchar(255) not null, allowed_address varchar(255) not null, endpoint varchar(255) not null, endpoint_port varchar(10) not null)`,
		`INSERT INTO peers VALUES(1,'stable-public-uuid','*1','legacy-peer','secret','public','wg0','10.0.0.2/32','example.test','51820')`,
	}
	for _, sql := range statements {
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := AutoMigrate(db); err != nil {
			t.Fatal(err)
		}
	}
	var account model.Admin
	if err := db.First(&account, 1).Error; err != nil {
		t.Fatal(err)
	}
	if account.Role != model.RoleSuperadmin || account.ResellerID != nil || account.Password != "existing-hash" {
		t.Fatalf("legacy account changed: %+v", account)
	}
	var peer model.Peer
	if err := db.First(&peer, 1).Error; err != nil {
		t.Fatal(err)
	}
	if peer.ResellerID != nil || peer.UUID != "stable-public-uuid" || peer.PrivateKey != "secret" {
		t.Fatal("legacy ownership or configuration changed")
	}
}
