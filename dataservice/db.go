package dataservice

import (
	"context"
	"database/sql"
	_ "embed"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"mikrotik-wg-go/dataservice/db"
)

//go:embed schema.sql
var ddl string

func ConnectDB() (*db.Queries, error) {

	sqlDB, err := sql.Open("sqlite3", "./mwp.db")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := sqlDB.ExecContext(context.Background(), ddl); err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	return db.New(sqlDB), nil
}
