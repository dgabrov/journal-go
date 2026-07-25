package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/google/uuid"
)

func (s Server) getUserByProvidedId(ctx context.Context, tx *sql.Tx, id string) (string, error) {
	var userID string
	err := tx.QueryRowContext(ctx, "SELECT user_id FROM user WHERE provided_id = ?", id).Scan(&userID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s Server) createSessionForUser(ctx context.Context, tx *sql.Tx, userID string, token string, expiry time.Time) error {
	sessionID := uuid.Must(uuid.NewV7()).String()

	_, err := tx.ExecContext(ctx,
		"INSERT INTO session (session_id, user_id, expired_ind, expire_dt, token) VALUES (?, ?, ?, ?, ?)",
		sessionID, userID, "N", expiry, token,
	)

	return err
}

func (s Server) getLoginResponse(ctx context.Context, tx *sql.Tx, userID string) (*data.LoginResponse, error) {
	var user data.User
	err := tx.QueryRowContext(ctx, "SELECT user_id, login, name FROM user WHERE user_id = ?", userID).
		Scan(&user.Id, &user.Login, &user.FullName)
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT j.journal_id, j.title, j.created_dt, j.created_dt,
		       ou.user_id, ou.login, ou.name,
		       CASE WHEN r.relation_cd = ? THEN 1 ELSE 0 END as is_owner
		FROM user_journal uj
		JOIN journal j ON uj.journal_id = j.journal_id
		JOIN user ou ON (SELECT user_id FROM user_journal WHERE journal_id = j.journal_id AND relation_cd = ?) = ou.user_id
		JOIN relation r ON uj.relation_cd = r.relation_cd
		WHERE uj.user_id = ?
	`, data.RelationOwner, data.RelationOwner, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	journals := make([]data.CompleteJournal, 0)
	for rows.Next() {
		var cj data.CompleteJournal
		var isOwner int

		err := rows.Scan(
			&cj.Journal.Id, &cj.Journal.Title, &cj.Journal.Created, &cj.Journal.LastUpdated,
			&cj.User.Id, &cj.User.Login, &cj.User.FullName,
			&isOwner,
		)
		if err != nil {
			return nil, err
		}

		cj.Owner = isOwner == 1
		journals = append(journals, cj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &data.LoginResponse{
		User:     user,
		Journals: journals,
	}, nil
}

func (s Server) buildLikePattern(search string) string {
	pattern := strings.ReplaceAll(search, "*", "%")
	pattern = strings.ReplaceAll(pattern, "?", "_")

	if !strings.HasPrefix(pattern, "%") {
		pattern = "%" + pattern
	}

	if !strings.HasSuffix(pattern, "%") {
		pattern = pattern + "%"
	}

	return pattern
}

func (s Server) searchUsers(ctx context.Context, tx *sql.Tx, userID string, search string) ([]data.User, error) {
	pattern := s.buildLikePattern(search)

	rows, err := tx.QueryContext(ctx, `
		SELECT user_id, login, name, provided_id
		FROM user
		WHERE login LIKE ?
		AND user_id != ?
		ORDER BY login
	`, pattern, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users = make([]data.User, 0)
	for rows.Next() {
		var user data.User
		err := rows.Scan(&user.Id, &user.Login, &user.FullName, &user.ProvidedId)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s Server) markSessionExpired(ctx context.Context, tx *sql.Tx, token string) error {
	_, err := tx.ExecContext(ctx, "UPDATE session SET expired_ind = 'Y' WHERE token = ?", token)
	return err
}

func (s Server) validateItemsOwnership(ctx context.Context, tx *sql.Tx, userID string, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(itemIDs)-1) + "?"
	args := make([]any, len(itemIDs))
	for i, id := range itemIDs {
		args[i] = id
	}

	query := `
		SELECT COUNT(*) FROM journal_item ji
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE ji.journal_item_id IN (` + placeholders + `)
		AND uj.user_id = ?
		AND uj.relation_cd = ?
	`
	args = append(args, userID, data.RelationOwner)

	var count int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return err
	}

	if count != len(itemIDs) {
		return errors.New("some items do not belong to journals owned by the user")
	}

	return nil
}

func (s Server) ValidateJournalItemOwnership(ctx context.Context, userID string, journalItemID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		SELECT COUNT(*) FROM journal_item ji
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE ji.journal_item_id = ?
		AND uj.user_id = ?
		AND uj.relation_cd = ?
	`

	var count int
	err = tx.QueryRowContext(ctx, query, journalItemID, userID, data.RelationOwner).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("user does not own this journal item")
	}

	return tx.Commit()
}

func (s Server) deleteItems(ctx context.Context, tx *sql.Tx, itemIDs []string) error {
	if len(itemIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(itemIDs)-1) + "?"
	deleteArgs := make([]any, len(itemIDs))
	for i, id := range itemIDs {
		deleteArgs[i] = id
	}

	deleteQuery := `DELETE FROM journal_item WHERE journal_item_id IN (` + placeholders + `)`
	_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	return err
}

func (s Server) validateJournalsOwnership(ctx context.Context, tx *sql.Tx, userID string, journalIDs []string) error {
	if len(journalIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(journalIDs)-1) + "?"
	args := make([]any, len(journalIDs))
	for i, id := range journalIDs {
		args[i] = id
	}

	query := `
		SELECT COUNT(*) FROM user_journal
		WHERE journal_id IN (` + placeholders + `)
		AND user_id = ?
		AND relation_cd = ?
	`
	args = append(args, userID, data.RelationOwner)

	var count int
	err := tx.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return err
	}

	if count != len(journalIDs) {
		return errors.New("some journals do not belong to the user or user is not the owner")
	}

	return nil
}

func (s Server) deleteUserUserJournals(ctx context.Context, tx *sql.Tx, journalIDs []string) error {
	if len(journalIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(journalIDs)-1) + "?"
	deleteArgs := make([]any, len(journalIDs))
	for i, id := range journalIDs {
		deleteArgs[i] = id
	}

	deleteQuery := `DELETE FROM user_journal WHERE journal_id IN (` + placeholders + `)`
	_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	return err
}

func (s Server) deleteJournals(ctx context.Context, tx *sql.Tx, journalIDs []string) error {
	if len(journalIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(journalIDs)-1) + "?"
	deleteArgs := make([]any, len(journalIDs))
	for i, id := range journalIDs {
		deleteArgs[i] = id
	}

	deleteQuery := `DELETE FROM journal WHERE journal_id IN (` + placeholders + `)`
	_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	return err
}

func (s Server) validateJournalAccess(ctx context.Context, tx *sql.Tx, userID string, journalID string) error {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_journal
		WHERE user_id = ? AND journal_id = ?
	`, userID, journalID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("journal not found or user does not have access")
	}

	return nil
}

func (s Server) retrieveJournalItems(ctx context.Context, tx *sql.Tx, journalID string) ([]data.JournalItem, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT journal_item_id, journal_id, created_dt, updated_dt, comments
		FROM journal_item
		WHERE journal_id = ?
		ORDER BY created_dt DESC
	`, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items = make([]data.JournalItem, 0)
	for rows.Next() {
		var item data.JournalItem
		var dt time.Time
		err := rows.Scan(&item.Id, &item.JournalID, &dt, &item.LastUpdated, &item.Comments)
		if err != nil {
			return nil, err
		}

		item.Date = formatDate(dt)
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func formatDate(dt time.Time) string {
	return dt.Format("Jan 2, 2006")
}

func (s Server) removeReadingPrivileges(ctx context.Context, tx *sql.Tx, journalID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(userIDs)-1) + "?"
	args := make([]any, len(userIDs))
	for i, id := range userIDs {
		args[i] = id
	}

	query := `DELETE FROM user_journal WHERE journal_id = ? AND relation_cd != ? AND user_id IN (` + placeholders + `)`
	args = append([]any{journalID, data.RelationOwner}, args...)

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (s Server) getReadingUsers(ctx context.Context, tx *sql.Tx, journalID string) ([]data.User, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT u.user_id, u.login, u.name, u.provided_id
		FROM user_journal uj
		JOIN user u ON uj.user_id = u.user_id
		WHERE uj.journal_id = ?
		AND uj.relation_cd != ?
		ORDER BY u.login
	`, journalID, data.RelationOwner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []data.User = make([]data.User, 0)
	for rows.Next() {
		var user data.User
		err := rows.Scan(&user.Id, &user.Login, &user.FullName, &user.ProvidedId)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s Server) addReadingPrivileges(ctx context.Context, tx *sql.Tx, journalID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	existingUsers, err := s.getExistingReadingUsers(ctx, tx, journalID)
	if err != nil {
		return err
	}

	existingSet := make(map[string]bool)
	for _, id := range existingUsers {
		existingSet[id] = true
	}

	for _, userID := range userIDs {
		if !existingSet[userID] {
			userJournalID := uuid.Must(uuid.NewV7()).String()
			now := time.Now()
			_, err := tx.ExecContext(ctx,
				"INSERT INTO user_journal (user_journal_id, relation_cd, user_id, journal_id, created_dt) VALUES (?, ?, ?, ?, ?)",
				userJournalID, data.RelationRead, userID, journalID, now,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s Server) getExistingReadingUsers(ctx context.Context, tx *sql.Tx, journalID string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id
		FROM user_journal
		WHERE journal_id = ?
		AND relation_cd != ?
	`, journalID, data.RelationOwner)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userIDs, nil
}

func (s Server) createJournal(ctx context.Context, tx *sql.Tx, journalID string, title string, createdDt time.Time) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO journal (journal_id, title, created_dt) VALUES (?, ?, ?)",
		journalID, title, createdDt,
	)
	return err
}

func (s Server) createJournalOwnership(ctx context.Context, tx *sql.Tx, journalID string, userID string, createdDt time.Time) error {
	userJournalID := uuid.Must(uuid.NewV7()).String()
	_, err := tx.ExecContext(ctx,
		"INSERT INTO user_journal (user_journal_id, relation_cd, user_id, journal_id, created_dt) VALUES (?, ?, ?, ?, ?)",
		userJournalID, data.RelationOwner, userID, journalID, createdDt,
	)
	return err
}

func (s Server) updateJournalTitle(ctx context.Context, tx *sql.Tx, journalID string, title string) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE journal SET title = ? WHERE journal_id = ?",
		title, journalID,
	)
	return err
}

func (s Server) ValidateAttachmentAccess(ctx context.Context, userID string, attachmentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM attachment a
		JOIN journal_item ji ON a.journal_item_id = ji.journal_item_id
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE a.attachment_id = ?
		AND uj.user_id = ?
	`, attachmentID, userID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("attachment not found or user does not have access")
	}

	return tx.Commit()
}

func (s Server) ValidateAttachmentOwnership(ctx context.Context, userID string, attachmentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM attachment a
		JOIN journal_item ji ON a.journal_item_id = ji.journal_item_id
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE a.attachment_id = ?
		AND uj.user_id = ?
		AND uj.relation_cd = ?
	`, attachmentID, userID, data.RelationOwner).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("attachment not found or user is not the owner")
	}

	return tx.Commit()
}
