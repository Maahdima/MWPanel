// Package testdb supplies isolated databases to integration tests only.
package testdb

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var schemaMu sync.Mutex
var schemas = map[string]bool{}

// Open defaults to SQLite. MWP_TEST_POSTGRES_DSN must point to a disposable
// PostgreSQL test database using keyword DSN syntax. Each temporary path gets a
// private schema; reconnecting to that path reuses its schema and preserves the
// concurrency tests' independent connection pools. Credentials are never logged.
func Open(t *testing.T, path string) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MWP_TEST_POSTGRES_DSN")
	config := &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)}
	var dialect gorm.Dialector
	if dsn == "" {
		dialect = sqlite.Open("file:" + path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	} else {
		if strings.Contains(dsn, "://") {
			t.Fatal("test PostgreSQL DSN must use keyword syntax")
		}
		digest := sha256.Sum256([]byte(path))
		schema := "mwp_test_" + hex.EncodeToString(digest[:16])
		schemaMu.Lock()
		if !schemas[schema] {
			base, err := gorm.Open(postgres.Open(dsn), config)
			if err != nil {
				schemaMu.Unlock()
				t.Fatal("could not connect to disposable PostgreSQL")
			}
			if err = base.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
				schemaMu.Unlock()
				t.Fatal(err)
			}
			sqlBase, _ := base.DB()
			sqlBase.Close()
			schemas[schema] = true
			t.Cleanup(func() {
				admin, err := gorm.Open(postgres.Open(dsn), config)
				if err != nil {
					t.Error("could not clean test schema")
					return
				}
				if err = admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
					t.Error(err)
				}
				pool, _ := admin.DB()
				pool.Close()
			})
		}
		schemaMu.Unlock()
		dialect = postgres.Open(dsn + " search_path=" + schema)
	}
	db, err := gorm.Open(dialect, config)
	if err != nil {
		t.Fatal("could not open isolated test database")
	}
	pool, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	pool.SetMaxOpenConns(8)
	t.Cleanup(func() { pool.Close() })
	return db
}
