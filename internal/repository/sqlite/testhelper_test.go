package sqlite_test

import (
	"database/sql"
	"testing"

	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sqliteRepo.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := sqliteRepo.RunMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
