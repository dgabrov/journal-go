package server

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amanagement24/journal-go/internal/data"
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

func (s Server) ValidateAttachmentsOwnership(ctx context.Context, userID string, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(attachmentIDs)-1) + "?"
	args := make([]any, len(attachmentIDs))
	for i, id := range attachmentIDs {
		args[i] = id
	}

	query := `
		SELECT COUNT(*) FROM attachment a
		JOIN journal_item ji ON a.journal_item_id = ji.journal_item_id
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE a.attachment_id IN (` + placeholders + `)
		AND uj.user_id = ?
		AND uj.relation_cd = ?
	`
	args = append(args, userID, data.RelationOwner)

	var count int
	err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return err
	}

	if count != len(attachmentIDs) {
		return errors.New("some attachments do not belong to journals owned by the user")
	}

	return tx.Commit()
}

func (s Server) DeleteAttachments(ctx context.Context, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	placeholders := strings.Repeat("?,", len(attachmentIDs)-1) + "?"
	deleteArgs := make([]any, len(attachmentIDs))
	for i, id := range attachmentIDs {
		deleteArgs[i] = id
	}

	deleteQuery := `DELETE FROM attachment WHERE attachment_id IN (` + placeholders + `)`
	_, err = tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) DeleteAttachmentFiles(attachmentIDs []string) error {
	for _, id := range attachmentIDs {
		filename := id + ".dat"

		if err := s.deleteAttachmentFile(s.config.Files.SmallFolder, id, filename); err != nil {
			return err
		}

		if err := s.deleteAttachmentFile(s.config.Files.RegularFolder, id, filename); err != nil {
			return err
		}
	}

	return nil
}

func (s Server) deleteAttachmentFile(folderPath string, id string, filename string) error {
	filePath := filepath.Join(folderPath, filename)
	if err := os.Remove(filePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			slog.Error("failed to delete attachment file", "id", id, "path", filePath, "error", err)
			return err
		}
	}
	return nil
}

func (s Server) UpdateAttachmentTitles(ctx context.Context, titleHolders []data.TitleHolder) error {
	if len(titleHolders) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, th := range titleHolders {
		_, err := tx.ExecContext(ctx, "UPDATE attachment SET title = ? WHERE attachment_id = ?", th.Title, th.AttachmentID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s Server) UpdateAttachmentTitle(ctx context.Context, attachmentID string, title string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE attachment SET title = ? WHERE attachment_id = ?", title, attachmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) UpdateAttachmentContentType(ctx context.Context, attachmentID string, contentType string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "UPDATE attachment SET content_type = ? WHERE attachment_id = ?", contentType, attachmentID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
