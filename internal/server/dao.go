package server

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
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

func (s Server) retrieveJournalItems(ctx context.Context, tx *sql.Tx, journalID string) ([]data.CompleteJournalItem, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT ji.journal_item_id, ji.journal_id, ji.created_dt, ji.updated_dt, ji.comments,
		       a.attachment_id, a.title, a.width, a.height
		FROM journal_item ji
		LEFT JOIN attachment a ON ji.journal_item_id = a.journal_item_id
		WHERE ji.journal_id = ?
		ORDER BY ji.created_dt DESC, a.attachment_id
	`, journalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemMap := make(map[string]*data.CompleteJournalItem)
	itemOrder := make([]string, 0)

	for rows.Next() {
		var itemID string
		var journalID string
		var dt time.Time
		var lastUpdated time.Time
		var comments string
		var attachmentID *string
		var attachmentTitle *string
		var width *int
		var height *int

		err := rows.Scan(&itemID, &journalID, &dt, &lastUpdated, &comments, &attachmentID, &attachmentTitle, &width, &height)
		if err != nil {
			return nil, err
		}

		if _, exists := itemMap[itemID]; !exists {
			itemMap[itemID] = &data.CompleteJournalItem{
				Id:          itemID,
				JournalID:   journalID,
				Date:        formatDate(dt),
				Comments:    comments,
				LastUpdated: lastUpdated,
				Attachments: make([]data.Attachment, 0),
			}
			itemOrder = append(itemOrder, itemID)
		}

		if attachmentID != nil && attachmentTitle != nil {
			w := 0
			h := 0
			if width != nil {
				w = *width
			}
			if height != nil {
				h = *height
			}
			itemMap[itemID].Attachments = append(itemMap[itemID].Attachments, data.Attachment{
				Id:     *attachmentID,
				Title:  *attachmentTitle,
				Width:  w,
				Height: h,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]data.CompleteJournalItem, len(itemOrder))
	for i, itemID := range itemOrder {
		items[i] = *itemMap[itemID]
	}

	return items, nil
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

func (s Server) validateNoDuplicateIds(ids []string) error {
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			return errors.New("duplicate ids found")
		}
		seen[id] = true
	}
	return nil
}

func (s Server) filterCurrentUser(userID string, targetUserIDs []string) []string {
	filtered := make([]string, 0)
	for _, id := range targetUserIDs {
		if id != userID {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func (s Server) filterDuplicateIds(ids []string) []string {
	seen := make(map[string]bool)
	filtered := make([]string, 0)
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func (s Server) validateJournalOwnership(ctx context.Context, tx *sql.Tx, userID string, journalID string) error {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM user_journal
		WHERE journal_id = ?
		AND user_id = ?
		AND relation_cd = ?
	`, journalID, userID, data.RelationOwner).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("journal not found or user is not the owner")
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

func (s Server) checkJournalOwnership(ctx context.Context, tx *sql.Tx, userID string, journalID string) (string, error) {
	row := tx.QueryRowContext(ctx, "SELECT relation_cd FROM user_journal WHERE user_id = ? AND journal_id = ?", userID, journalID)
	var relationCd string
	err := row.Scan(&relationCd)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("journal not found or user does not have access")
	}
	if err != nil {
		return "", err
	}
	return relationCd, nil
}

func (s Server) getSessionByToken(ctx context.Context, tx *sql.Tx, token string) (userID string, expiredInd string, expireDt *time.Time, err error) {
	row := tx.QueryRowContext(ctx, "SELECT user_id, expired_ind, expire_dt FROM session WHERE token = ?", token)
	err = row.Scan(&userID, &expiredInd, &expireDt)
	return
}

func (s Server) updateSessionExpiry(ctx context.Context, tx *sql.Tx, token string, newExpiry time.Time) error {
	_, err := tx.ExecContext(ctx, "UPDATE session SET expire_dt = ? WHERE token = ?", newExpiry, token)
	return err
}

func (s Server) updateJournalItemComments(ctx context.Context, tx *sql.Tx, journalItemID string, comments string, now time.Time) error {
	_, err := tx.ExecContext(ctx, "UPDATE journal_item SET comments = ?, updated_dt = ? WHERE journal_item_id = ?", comments, now, journalItemID)
	return err
}

func (s Server) insertJournalItem(ctx context.Context, tx *sql.Tx, id string, journalID string, createdDate time.Time, updatedDate time.Time, comments string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO journal_item (journal_item_id, journal_id, created_dt, updated_dt, comments) VALUES (?, ?, ?, ?, ?)", id, journalID, createdDate, updatedDate, comments)
	return err
}

func (s Server) insertUser(ctx context.Context, tx *sql.Tx, userID string, id string, login string, name string, now time.Time) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO user (user_id, provided_id, login, name, created_dt) VALUES (?, ?, ?, ?, ?)", userID, id, login, name, now)
	return err
}

func (s Server) validateAttachmentAccessQuery(ctx context.Context, tx *sql.Tx, attachmentID string, userID string) (int, error) {
	var count int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM attachment a
		JOIN journal_item ji ON a.journal_item_id = ji.journal_item_id
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE a.attachment_id = ?
		AND uj.user_id = ?
	`, attachmentID, userID).Scan(&count)
	return count, err
}

func (s Server) getAttachmentContentTypeQuery(ctx context.Context, tx *sql.Tx, attachmentID string) (string, error) {
	var contentType sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT content_type FROM attachment WHERE attachment_id = ?", attachmentID).Scan(&contentType)
	if err != nil {
		return "image/png", nil
	}

	if !contentType.Valid || contentType.String == "" {
		return "image/png", nil
	}

	return contentType.String, nil
}

func (s Server) validateJournalItemOwnershipQuery(ctx context.Context, tx *sql.Tx, userID string, journalItemID string) (int, error) {
	query := `
		SELECT COUNT(*) FROM journal_item ji
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE ji.journal_item_id = ?
		AND uj.user_id = ?
		AND uj.relation_cd = ?
	`

	var count int
	err := tx.QueryRowContext(ctx, query, journalItemID, userID, data.RelationOwner).Scan(&count)
	return count, err
}

func (s Server) insertAttachment(ctx context.Context, tx *sql.Tx, attachmentID string, journalItemID string, contentType string, width int, height int, now time.Time) error {
	_, err := tx.ExecContext(ctx,
		"INSERT INTO attachment (attachment_id, journal_item_id, title, content_type, width, height, created_dt, updated_dt) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		attachmentID, journalItemID, "", contentType, width, height, now, now,
	)
	return err
}

func (s Server) validateAttachmentsOwnershipQuery(ctx context.Context, tx *sql.Tx, attachmentIDs []string, userID string) (int, error) {
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
	err := tx.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func (s Server) deleteAttachmentsQuery(ctx context.Context, tx *sql.Tx, attachmentIDs []string) error {
	placeholders := strings.Repeat("?,", len(attachmentIDs)-1) + "?"
	deleteArgs := make([]any, len(attachmentIDs))
	for i, id := range attachmentIDs {
		deleteArgs[i] = id
	}

	deleteQuery := `DELETE FROM attachment WHERE attachment_id IN (` + placeholders + `)`
	_, err := tx.ExecContext(ctx, deleteQuery, deleteArgs...)
	return err
}

func (s Server) updateAttachmentTitlesQuery(ctx context.Context, tx *sql.Tx, titleHolders []data.TitleHolder) error {
	for _, th := range titleHolders {
		_, err := tx.ExecContext(ctx, "UPDATE attachment SET title = ? WHERE attachment_id = ?", th.Title, th.AttachmentID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s Server) updateAttachmentTitleQuery(ctx context.Context, tx *sql.Tx, attachmentID string, title string) error {
	_, err := tx.ExecContext(ctx, "UPDATE attachment SET title = ? WHERE attachment_id = ?", title, attachmentID)
	return err
}

func (s Server) updateAttachmentContentTypeQuery(ctx context.Context, tx *sql.Tx, attachmentID string, contentType string) error {
	_, err := tx.ExecContext(ctx, "UPDATE attachment SET content_type = ? WHERE attachment_id = ?", contentType, attachmentID)
	return err
}

func (s Server) swapAttachmentDimensions(ctx context.Context, tx *sql.Tx, attachmentID string) error {
	var width, height int
	err := tx.QueryRowContext(ctx, "SELECT width, height FROM attachment WHERE attachment_id = ?", attachmentID).Scan(&width, &height)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "UPDATE attachment SET width = ?, height = ? WHERE attachment_id = ?", height, width, attachmentID)
	return err
}

func (s Server) validateJournalItemExistenceAndAccess(ctx context.Context, tx *sql.Tx, userID string, journalItemID string) error {
	query := `
		SELECT COUNT(*) FROM journal_item ji
		JOIN user_journal uj ON ji.journal_id = uj.journal_id
		WHERE ji.journal_item_id = ?
		AND uj.user_id = ?
	`

	var count int
	err := tx.QueryRowContext(ctx, query, journalItemID, userID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("journal item not found or user does not have access")
	}

	return nil
}

func (s Server) validateAttachmentExistence(ctx context.Context, tx *sql.Tx, journalItemID string) error {
	query := `
		SELECT COUNT(*) FROM attachment
		WHERE journal_item_id = ?
	`

	var count int
	err := tx.QueryRowContext(ctx, query, journalItemID).Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("no attachments found for this journal item")
	}

	return nil
}

func (s Server) getJournalTitleAndItemCreatedDt(ctx context.Context, tx *sql.Tx, journalItemID string) (string, time.Time, error) {
	query := `
		SELECT j.title, ji.created_dt
		FROM journal_item ji
		JOIN journal j ON ji.journal_id = j.journal_id
		WHERE ji.journal_item_id = ?
	`

	var title string
	var createdDt time.Time

	err := tx.QueryRowContext(ctx, query, journalItemID).Scan(&title, &createdDt)
	if err != nil {
		return "", time.Time{}, err
	}

	return title, createdDt, nil
}

func (s Server) createJobEntry(ctx context.Context, tx *sql.Tx, journalItemID string, journalTitle string, itemCreatedDt time.Time, now time.Time) error {
	jobID := uuid.Must(uuid.NewV7()).String()
	formattedDt := itemCreatedDt.Format("2006-01-02")
	jobName := journalTitle + "_" + formattedDt

	_, err := tx.ExecContext(ctx,
		"INSERT INTO job (job_id, name, journal_item_id, status, create_dt) VALUES (?, ?, ?, ?, ?)",
		jobID, jobName, journalItemID, "pending", now,
	)

	return err
}

type JobRecord struct {
	ID            string
	Name          string
	JournalItemID string
	Status        string
	CreateDt      time.Time
}

type AttachmentRecord struct {
	ID          string
	ContentType string
}

func (s Server) getJobRecord(ctx context.Context, tx *sql.Tx, jobID string) (*JobRecord, error) {
	query := `
		SELECT job_id, name, journal_item_id, status, create_dt
		FROM job
		WHERE job_id = ?
	`

	var record JobRecord
	err := tx.QueryRowContext(ctx, query, jobID).Scan(
		&record.ID, &record.Name, &record.JournalItemID, &record.Status, &record.CreateDt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("job not found")
	}

	if err != nil {
		return nil, err
	}

	return &record, nil
}

func (s Server) getAttachmentsForJournalItem(ctx context.Context, tx *sql.Tx, journalItemID string) ([]AttachmentRecord, error) {
	query := `
		SELECT attachment_id, content_type
		FROM attachment
		WHERE journal_item_id = ?
	`

	rows, err := tx.QueryContext(ctx, query, journalItemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []AttachmentRecord
	for rows.Next() {
		var attachment AttachmentRecord
		if err := rows.Scan(&attachment.ID, &attachment.ContentType); err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

func (s Server) updateJobStatus(ctx context.Context, tx *sql.Tx, jobID string, status string) error {
	_, err := tx.ExecContext(ctx,
		"UPDATE job SET status = ? WHERE job_id = ?",
		status, jobID,
	)
	return err
}
