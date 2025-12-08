package persistence

import (
	"database/sql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func Migrate(db *sql.DB) error {
	sql := `CREATE TABLE IF NOT EXISTS sync_state(
			    id TEXT PRIMARY KEY,
			    value INTEGER NOT NULL
			);
			
			INSERT INTO sync_state(id, value) VALUES ('Record', 0) ON CONFLICT(id) DO NOTHING;
			
			CREATE TABLE IF NOT EXISTS records (
			    id          TEXT PRIMARY KEY,
			    created_at  DATETIME NOT NULL,
			    modified_at DATETIME NOT NULL,
			    type        INTEGER NOT NULL,
			    big_data  	BOOLEAN NOT NULL DEFAULT 0,
			    data        BLOB NULL,
			    dek         BLOB NULL,
			    deleted     BOOLEAN NOT NULL DEFAULT 0,
			    version     INT NOT NULL,
			    corrupted 	BOOLEAN NOT NULL DEFAULT 0
			);
			
			CREATE TABLE IF NOT EXISTS conflicts(
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
			    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			    record_id TEXT,
			    local BLOB,
			    remote BLOB
			);
			
			CREATE TABLE IF NOT EXISTS users(
			    id TEXT PRIMARY KEY,
			    username TEXT NOT NULL,
			    password BLOB NOT NULL,
			    salt    BLOB NOT NULL
			);
			
			CREATE TABLE IF NOT EXISTS servers(
			    id INTEGER PRIMARY KEY AUTOINCREMENT,
			    address TEXT NOT NULL,
			    login TEXT NOT NULL,
			    password BLOB NOT NULL,
			    active BOOLEAN
			)`
	_, err := db.Exec(sql)
	if err != nil {
		return err
	}
	return nil
}
