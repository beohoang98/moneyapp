package services_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	rawDB, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$12$cbb.WlWr0vj.nkJkv7dckOYvNPEjGzCTqyHxz.7rEORF2jurs1jJG')`,
		`CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type TEXT NOT NULL CHECK(type IN ('expense', 'income')),
			is_default BOOLEAN NOT NULL DEFAULT 0,
			color TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE UNIQUE INDEX idx_categories_name_type ON categories(name, type)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Food', 'expense', 1)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Transport', 'expense', 1)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Uncategorized', 'expense', 1)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Salary', 'income', 1)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Freelance', 'income', 1)`,
		`INSERT INTO categories (name, type, is_default) VALUES ('Uncategorized', 'income', 1)`,
		`CREATE TABLE expenses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount INTEGER NOT NULL CHECK(amount > 0),
			date DATE NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id),
			description TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE incomes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			amount INTEGER NOT NULL CHECK(amount > 0),
			date DATE NOT NULL,
			category_id INTEGER NOT NULL REFERENCES categories(id),
			description TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE invoices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			vendor_name TEXT NOT NULL,
			amount INTEGER NOT NULL CHECK(amount > 0),
			issue_date DATE NOT NULL,
			due_date DATE NOT NULL,
			status TEXT NOT NULL DEFAULT 'unpaid' CHECK(status IN ('unpaid', 'paid', 'overdue')),
			description TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, m := range migrations {
		if _, err := rawDB.Exec(m); err != nil {
			t.Fatalf("apply migration: %v\nSQL: %s", err, m)
		}
	}

	db, err := gorm.Open(sqlite.Dialector{Conn: rawDB}, &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	t.Cleanup(func() { rawDB.Close() })

	return db
}
