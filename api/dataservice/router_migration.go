package dataservice

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/maahdima/mwp/api/dataservice/model"
	"gorm.io/gorm"
)

// MigrateRouterOwnership binds pre-multi-router records only when their source is
// unambiguous. With multiple legacy servers, the operator must explicitly select
// LEGACY_ROUTER_SERVER_ID. UUIDs, keys, counters and ownership are never changed.
func MigrateRouterOwnership(db *gorm.DB, legacyServerID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var unresolved int64

		for _, table := range []string{"peers", "interfaces"} {
			var count int64

			if err := tx.Table(table).Where("server_id = ?", 0).Count(&count).Error; err != nil {
				return err
			}

			unresolved += count
		}

		if unresolved > 0 {
			var servers []model.Server

			if err := tx.Order("id ASC").Find(&servers).Error; err != nil {
				return err
			}

			selected := uint(0)

			if strings.TrimSpace(legacyServerID) != "" {
				id, err := strconv.ParseUint(strings.TrimSpace(legacyServerID), 10, 32)

				if err != nil || id == 0 {
					return fmt.Errorf("invalid LEGACY_ROUTER_SERVER_ID")
				}

				for _, server := range servers {
					if server.ID == uint(id) {
						selected = server.ID
					}
				}

				if selected == 0 {
					return fmt.Errorf("LEGACY_ROUTER_SERVER_ID does not identify an existing server")
				}
			} else if len(servers) == 1 {
				selected = servers[0].ID
			} else if len(servers) > 1 {
				return fmt.Errorf("legacy router ownership is ambiguous: set LEGACY_ROUTER_SERVER_ID to the verified source router before restarting")
			}

			if selected != 0 {
				for _, table := range []string{"peers", "interfaces"} {
					if err := tx.Table(table).Where("server_id = ?", 0).UpdateColumn("server_id", selected).Error; err != nil {
						return err
					}
				}
			}
		}

		// Pools are bound by their existing, globally unique interface database ID.
		var pools []model.IPPool

		if err := tx.Where("server_id = ?", 0).Find(&pools).Error; err != nil {
			return err
		}

		for _, pool := range pools {
			var iface model.Interface

			if err := tx.First(&iface, pool.InterfaceID).Error; err != nil {
				return fmt.Errorf("pool %d has no source interface: %w", pool.ID, err)
			}

			if err := tx.Model(&model.IPPool{}).Where("id = ?", pool.ID).UpdateColumn("server_id", iface.ServerID).Error; err != nil {
				return err
			}
		}

		// AutoMigrate creates new composite indexes first. Remove only the superseded
		// global indexes, within this transaction and after verifying their replacement.
		replacements := []struct {
			table    any
			old, new string
		}{
			{&model.Peer{}, "idx_peers_peer_id", "uidx_peer_router_id"},
			{&model.Peer{}, "idx_peers_allowed_address", "uidx_peer_router_address"},
			{&model.Interface{}, "idx_interfaces_interface_id", "uidx_interface_router_id"},
			{&model.Interface{}, "idx_interfaces_name", "uidx_interface_router_name"},
			{&model.IPPool{}, "idx_ip_pools_name", "uidx_pool_router_name"},
			{&model.IPPool{}, "idx_ip_pools_start_ip", "uidx_pool_router_start"},
			{&model.IPPool{}, "idx_ip_pools_end_ip", "uidx_pool_router_end"},
		}

		for _, index := range replacements {
			if !tx.Migrator().HasIndex(index.table, index.new) {
				return fmt.Errorf("router migration missing replacement index %s", index.new)
			}

			if tx.Migrator().HasIndex(index.table, index.old) {
				if err := tx.Migrator().DropIndex(index.table, index.old); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
