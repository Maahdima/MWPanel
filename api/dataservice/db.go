package dataservice

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/maahdima/mwp/api/config"
	"github.com/maahdima/mwp/api/dataservice/model"

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
	err := db.Migrator().AutoMigrate(&model.Admin{}, &model.Server{}, &model.Interface{}, &model.IPPool{}, &model.Peer{}, &model.Traffic{})
	if err != nil {
		log.Panic("failed to auto migrate db: ", err)
		return err
	}
	return nil
}
