package dataservice

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"
	"github.com/maahdima/mwp/api/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDB(config config.DBConfig) (db *gorm.DB, err error) {
	conf := new(gorm.Config)
	if os.Getenv("MODE") == "development" {
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      false,
				Colorful:                  true,
			},
		)
		conf.Logger = newLogger
	}

	var dialect gorm.Dialector
	if config.Dialect == "sqlite" {
		dialect = sqlite.Open(
			fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", config.Database),
		)
	} else if config.Dialect == "postgres" {
		var dsn string

		dsn = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
			config.Host, config.Username, config.Password, config.Database, config.Port)

		dialect = postgres.New(postgres.Config{DSN: dsn})
	} else {
		err = errors.New("invalid db dialect")
		return
	}

	db, err = gorm.Open(dialect, conf)
	if err != nil {
		return
	}

	return
}

func AutoMigrate(db *gorm.DB) error {
	migrator := db.Migrator()
	if migrator.HasTable(&model.TelegramChat{}) && !migrator.HasColumn(&model.TelegramChat{}, "peer_uuid") {
		_ = migrator.DropTable(&model.TelegramChat{})
	}

	if migrator.HasTable(&model.Peer{}) {
		renamePeerColumn(migrator, "first_notify", "traffic_first_notify")
		renamePeerColumn(migrator, "second_notify", "traffic_second_notify")
		renamePeerColumn(migrator, "third_notify", "traffic_third_notify")
	}

	err := migrator.AutoMigrate(
		&model.Interface{},
		&model.IPPool{},
		&model.Peer{},
		&model.Traffic{},
		&model.TotalTrafficUsage{},
		&model.Server{},
		&model.Admin{},
		&model.TelegramChat{},
		&model.PeerSession{},
		&model.Plan{},
		&model.Order{},
		&model.Deposit{},
		&model.ShopSetting{},
	)
	if err != nil {
		log.Panic("failed to auto migrate db: ", err)
		return err
	}

	linkOrderPeers(db)
	backfillPlanNames(db)

	return nil
}

// linkOrderPeers copies legacy peer UUID columns onto config_id, then drops them.
func linkOrderPeers(db *gorm.DB) {
	migrator := db.Migrator()
	if !migrator.HasTable(&model.Order{}) || !migrator.HasTable(&model.Peer{}) {
		return
	}

	hasPeerUUID := dbHasColumn(db, &model.Order{}, "peer_uuid")
	hasResultUUID := dbHasColumn(db, &model.Order{}, "result_peer_uuid")
	if !hasPeerUUID && !hasResultUUID {
		return
	}

	type legacyOrder struct {
		ID             uint
		ConfigID       *uint
		PeerUUID       *string `gorm:"column:peer_uuid"`
		ResultPeerUUID *string `gorm:"column:result_peer_uuid"`
	}

	var rows []legacyOrder
	if err := db.Table("vpn_orders").Find(&rows).Error; err != nil {
		log.Printf("failed to read vpn orders for config relation migration: %v", err)
		return
	}

	for _, row := range rows {
		if row.ConfigID != nil && *row.ConfigID != 0 {
			continue
		}
		uuid := ""
		if row.ResultPeerUUID != nil && strings.TrimSpace(*row.ResultPeerUUID) != "" {
			uuid = strings.TrimSpace(*row.ResultPeerUUID)
		} else if row.PeerUUID != nil {
			uuid = strings.TrimSpace(*row.PeerUUID)
		}
		if uuid == "" {
			continue
		}
		var peer model.Peer
		if err := db.Where("uuid = ?", uuid).First(&peer).Error; err != nil {
			continue
		}
		if err := db.Table("vpn_orders").Where("id = ?", row.ID).Update("config_id", peer.ID).Error; err != nil {
			log.Printf("failed to link vpn order %d to peer %d: %v", row.ID, peer.ID, err)
		}
	}

	if hasResultUUID {
		if err := db.Exec("ALTER TABLE vpn_orders DROP COLUMN result_peer_uuid").Error; err != nil {
			log.Printf("failed to drop vpn_orders.result_peer_uuid: %v", err)
		}
	}
	if hasPeerUUID {
		if err := db.Exec("ALTER TABLE vpn_orders DROP COLUMN peer_uuid").Error; err != nil {
			log.Printf("failed to drop vpn_orders.peer_uuid: %v", err)
		}
	}
}

func dbHasColumn(db *gorm.DB, modelValue any, column string) bool {
	columns, err := db.Migrator().ColumnTypes(modelValue)
	if err != nil {
		return false
	}

	for _, item := range columns {
		if strings.EqualFold(item.Name(), column) {
			return true
		}
	}

	return false
}

func backfillPlanNames(db *gorm.DB) {
	if !db.Migrator().HasTable(&model.Plan{}) {
		return
	}

	var plans []model.Plan
	if err := db.Find(&plans).Error; err != nil {
		log.Printf("failed to read vpn plans for name backfill: %v", err)
		return
	}

	for _, plan := range plans {
		updates := map[string]any{}
		if strings.TrimSpace(plan.FaName) == "" {
			faName := strings.TrimSpace(plan.Title)
			if faName == "" {
				faName = utils.PersianPlanTitle(plan.TrafficGB, plan.DurationDays)
			}
			updates["fa_name"] = faName
		}
		if strings.TrimSpace(plan.Name) == "" {
			updates["name"] = fmt.Sprintf("%d GB / %d days", plan.TrafficGB, plan.DurationDays)
		}
		if len(updates) == 0 {
			continue
		}
		if err := db.Model(&plan).Updates(updates).Error; err != nil {
			log.Printf("failed to backfill vpn plan %d names: %v", plan.ID, err)
		}
	}
}

func renamePeerColumn(migrator gorm.Migrator, oldName, newName string) {
	if migrator.HasColumn(&model.Peer{}, oldName) && !migrator.HasColumn(&model.Peer{}, newName) {
		_ = migrator.RenameColumn(&model.Peer{}, oldName, newName)
	}
}
