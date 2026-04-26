package database

import (
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open initialises the SQLite database and returns a *gorm.DB handle.
//
// Migration approach: embedded .sql files are applied first via database/sql
// (keeping the existing migrations table as the version tracker). GORM then
// opens the same file in ORM mode. This avoids two competing migration runners
// and prevents silent schema drift — the SQL files remain the single source of
// truth for DDL.
func Open(path string, migrationsFS embed.FS) (*gorm.DB, error) {
	dsn := path + "?_journal_mode=WAL&_foreign_keys=on"

	rawDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := rawDB.Ping(); err != nil {
		rawDB.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	if err := RunMigrations(rawDB, migrationsFS); err != nil {
		rawDB.Close()
		return nil, fmt.Errorf("migrate db: %w", err)
	}
	rawDB.Close()

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Warn),
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
	})
	if err != nil {
		return nil, fmt.Errorf("open gorm: %w", err)
	}

	return db, nil
}
