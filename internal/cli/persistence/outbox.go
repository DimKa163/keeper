package persistence

import (
	"context"
	"database/sql"
	"github.com/DimKa163/keeper/internal/cli/core"
)

const (
	insertOutboxStmt   = `INSERT INTO outbox (created_at, message, type) VALUES (?, ?, ?)`
	getAllMessagesStmt = `SELECT id, created_at, message, type FROM outbox`

	deleteOutboxStmt = `DELETE FROM outbox WHERE id = ?`
)

func TxInsertOutbox(ctx context.Context, db *sql.Tx, model *core.Outbox) error {
	if _, err := db.ExecContext(ctx, insertOutboxStmt, model.CreatedAt, model.Message, model.Type); err != nil {
		return err
	}
	return nil
}

func TxGetAllMessages(ctx context.Context, db *sql.Tx) ([]*core.Outbox, error) {
	rows, err := db.QueryContext(ctx, getAllMessagesStmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	outboxes := make([]*core.Outbox, 0)
	for rows.Next() {
		var id int64
		var created_at sql.NullTime
		var message []byte
		var tp core.OperationType
		if err := rows.Scan(&id, &created_at, &message, &tp); err != nil {
			return nil, err
		}
		var outbox core.Outbox
		outbox.ID = id
		if created_at.Valid {
			outbox.CreatedAt = created_at.Time
		}
		outbox.Message = message
		outbox.Type = tp
		outboxes = append(outboxes, &outbox)
	}
	return outboxes, nil
}

func TxDeleteOutbox(ctx context.Context, db *sql.Tx, id int64) error {
	if _, err := db.ExecContext(ctx, deleteOutboxStmt, id); err != nil {
		return err
	}
	return nil
}
