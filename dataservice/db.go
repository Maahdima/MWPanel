package dataservice

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"mikrotik-wg-go/config"
	"mikrotik-wg-go/dataservice/model"
	"os"
	"time"
)

func ConnectDB(config config.DBConfig) (db *gorm.DB, err error) {
	conf := new(gorm.Config)
	if os.Getenv("MODE") == "development" {
		newLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				// TODO : use config values
				SlowThreshold:             time.Second,
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: false,
				ParameterizedQueries:      false,
				Colorful:                  false,
			},
		)
		conf.Logger = newLogger
	}

	var dialect gorm.Dialector
	if config.Dialect == "sqlite" {
		dialect = sqlite.New(sqlite.Config{
			DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
				config.Username, config.Password, config.Host, config.Port, config.Database)})

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
	err := db.Migrator().AutoMigrate(&model.Admin{}, &model.Interface{}, &model.Peer{}, &model.Server{})
	if err != nil {
		log.Panic("failed to auto migrate db: ", err)
		return err
	}
	return nil
}
