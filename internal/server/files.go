package server

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s Server) CreateAttachment(ctx context.Context, journalItemID string, filename string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	attachmentID := uuid.Must(uuid.NewV7()).String()
	now := time.Now()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO attachment (attachment_id, journal_item_id, title, created_dt, updated_dt) VALUES (?, ?, ?, ?, ?)",
		attachmentID, journalItemID, filename, now, now,
	)
	if err != nil {
		return "", err
	}

	return attachmentID, tx.Commit()
}
