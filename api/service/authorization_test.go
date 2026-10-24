package service

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

func TestPeerOwnershipScopesFailClosed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "owners.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err := db.AutoMigrate(&model.Peer{}); err != nil {
		t.Fatal(err)
	}
	one, two := uint(1), uint(2)
	peers := []model.Peer{{UUID: "legacy", PeerID: "*1", Name: "legacy", AllowedAddress: "10.0.0.1"}, {UUID: "one", PeerID: "*2", Name: "own", AllowedAddress: "10.0.0.2", ResellerID: &one}, {UUID: "two", PeerID: "*3", Name: "other", AllowedAddress: "10.0.0.3", ResellerID: &two}}
	for i := range peers {
		if err := db.Create(&peers[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	reseller := Principal{AccountID: 10, Role: model.RoleReseller, ResellerID: &one}
	for _, tc := range []struct {
		actor Principal
		want  int
	}{{reseller, 1}, {Principal{AccountID: 11, Role: model.RoleSuperadmin}, 3}, {Principal{}, 0}, {Principal{AccountID: 12, Role: model.RoleReseller}, 0}, {Principal{AccountID: 13, Role: "forged", ResellerID: &one}, 0}} {
		var got []model.Peer
		if err := tc.actor.PeerScope(db).Find(&got).Error; err != nil {
			t.Fatal(err)
		}
		if len(got) != tc.want {
			t.Errorf("%+v: got %d, want %d", tc.actor, len(got), tc.want)
		}
	}
	for _, id := range []uint{peers[0].ID, peers[2].ID, 999} {
		if _, err := reseller.FindOwnedPeer(db, id); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatal(fmt.Sprintf("cross-tenant ID %d: %v", id, err))
		}
	}
	if got, err := reseller.FindOwnedPeer(db, peers[1].ID); err != nil || got.UUID != "one" {
		t.Fatalf("own peer unavailable: %+v %v", got, err)
	}
}
