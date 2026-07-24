package server

import (
	"context"
	"time"
)

func (s Server) CreateAttachment(ctx context.Context, journalItemID string, attachmentID string, filename string, contentType string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	_, err = tx.ExecContext(ctx,
		"INSERT INTO attachment (attachment_id, journal_item_id, title, content_type, created_dt, updated_dt) VALUES (?, ?, ?, ?, ?, ?)",
		attachmentID, journalItemID, filename, contentType, now, now,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
