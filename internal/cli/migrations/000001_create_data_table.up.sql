CREATE TABLE IF NOT EXISTS sync_state(
    id TEXT PRIMARY KEY,
    value DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS records (
    id          TEXT PRIMARY KEY,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type        INTEGER NOT NULL,
    data        BLOB NOT NULL,
    data_nonce  BLOB NOT NULL,
    dek         BLOB NOT NULL,
    dek_nonce   BLOB NOT NULL,
    version     INT NOT NULL
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