package server

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/google/uuid"
)

type Server struct {
	db     *sql.DB
	config *data.ConfigData
}

type ErrTokenNotFound struct{}

func (e *ErrTokenNotFound) Error() string {
	return "token not found"
}

type ErrTokenExpired struct{}

func (e *ErrTokenExpired) Error() string {
	return "token expired"
}

func (s Server) GetUserByProvidedId(ctx context.Context, providedID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	userID, err := s.getUserByProvidedId(ctx, tx, providedID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return userID, nil
}

func (s Server) CreateSessionForUser(ctx context.Context, userID string, token string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	expiry := time.Now().Add(time.Duration(s.config.TokenTimeToLive) * time.Second)
	if err := s.createSessionForUser(ctx, tx, userID, token, expiry); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func New(db *sql.DB, config *data.ConfigData) *Server {
	return &Server{
		db:     db,
		config: config,
	}
}

func (s Server) GetLoginResponse(ctx context.Context, id string) (*data.LoginResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	response, err := s.getLoginResponse(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return response, nil
}

func (s Server) GetUserIdFromToken(ctx context.Context, token string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var userID string
	var expiredInd string
	var expireDt *time.Time

	row := tx.QueryRowContext(ctx, "SELECT user_id, expired_ind, expire_dt FROM session WHERE token = ?", token)
	err = row.Scan(&userID, &expiredInd, &expireDt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", &ErrTokenNotFound{}
	}
	if err != nil {
		return "", err
	}

	// Check if token is marked as expired
	if expiredInd == "Y" {
		return "", &ErrTokenExpired{}
	}

	// Check if token has passed expiry date
	if expireDt == nil || time.Now().After(*expireDt) {
		return "", &ErrTokenExpired{}
	}

	// Update expiry_dt to now + tokenTimeToLive
	newExpiry := time.Now().Add(time.Duration(s.config.TokenTimeToLive) * time.Second)
	_, err = tx.ExecContext(ctx, "UPDATE session SET expire_dt = ? WHERE token = ?", newExpiry, token)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return userID, nil
}

func (s Server) UpdateJournalItem(ctx context.Context, userID string, item data.JournalItem) error {
	if item.Id == "" || item.JournalID == "" {
		return errors.New("journal item id and journal id are required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if user owns the journal
	row := tx.QueryRowContext(ctx, "SELECT relation_cd FROM user_journal WHERE user_id = ? AND journal_id = ?", userID, item.JournalID)
	var relationCd string
	err = row.Scan(&relationCd)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("journal not found or user does not have access")
	}
	if err != nil {
		return err
	}

	if relationCd != data.RelationOwner {
		return errors.New("user is not the owner of this journal")
	}

	// Update the journal item with comments and current timestamp
	now := time.Now()
	_, err = tx.ExecContext(ctx, "UPDATE journal_item SET comments = ?, updated_dt = ? WHERE journal_item_id = ?", item.Comments, now, item.Id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) AddJournalItem(ctx context.Context, userID string, item data.JournalItem) error {
	if item.JournalID == "" {
		return errors.New("journal id is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if user owns the journal
	row := tx.QueryRowContext(ctx, "SELECT relation_cd FROM user_journal WHERE user_id = ? AND journal_id = ?", userID, item.JournalID)
	var relationCd string
	err = row.Scan(&relationCd)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("journal not found or user does not have access")
	}
	if err != nil {
		return err
	}

	if relationCd != data.RelationOwner {
		return errors.New("user is not the owner of this journal")
	}

	// Generate ID for the new journal item
	id := item.Id
	createdDate, err := parseDate(item.Date)
	if err != nil {
		return err
	}
	updatedDate := time.Now()

	// Insert the journal item
	_, err = tx.ExecContext(ctx, "INSERT INTO journal_item (journal_item_id, journal_id, created_dt, updated_dt, comments) VALUES (?, ?, ?, ?, ?)", id, item.JournalID, createdDate, updatedDate, item.Comments)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func parseDate(dtStr string) (time.Time, error) {
	formats := []string{
		"Jan 2, 2006",
		"Jan2, 2006",
		"Jan 2",
		"Jan2",
		"2006-01-02",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dtStr)
		if err == nil {
			if format == "Jan 2" {
				currentYear := time.Now().Year()
				return time.Date(currentYear, t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
			}
			return t, nil
		}
	}

	return time.Time{}, errors.New("cannot parse date in any supported format")
}

func (s Server) SearchUsers(ctx context.Context, userID string, search string) ([]data.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	users, err := s.searchUsers(ctx, tx, userID, search)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s Server) LogoutUser(ctx context.Context, token string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = s.markSessionExpired(ctx, tx, token)
	if err != nil {
		return err
	}

	return tx.Commit()
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

func (s Server) RemoveItems(ctx context.Context, userID string, itemIDs []string) error {
	if err := s.validateNoDuplicateIds(itemIDs); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.validateItemsOwnership(ctx, tx, userID, itemIDs); err != nil {
		return err
	}

	if err := s.deleteItems(ctx, tx, itemIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) RemoveJournals(ctx context.Context, userID string, journalIDs []string) error {
	if err := s.validateNoDuplicateIds(journalIDs); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.validateJournalsOwnership(ctx, tx, userID, journalIDs); err != nil {
		return err
	}

	if err := s.deleteUserUserJournals(ctx, tx, journalIDs); err != nil {
		return err
	}

	if err := s.deleteJournals(ctx, tx, journalIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) GetJournalItems(ctx context.Context, userID string, journalID string) ([]data.CompleteJournalItem, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.validateJournalAccess(ctx, tx, userID, journalID); err != nil {
		return nil, err
	}

	items, err := s.retrieveJournalItems(ctx, tx, journalID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s Server) RemoveReading(ctx context.Context, userID string, journalID string, targetUserIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.validateJournalOwnership(ctx, tx, userID, journalID); err != nil {
		return err
	}

	filteredUserIDs := s.filterCurrentUser(userID, targetUserIDs)
	filteredUserIDs = s.filterDuplicateIds(filteredUserIDs)

	if err := s.removeReadingPrivileges(ctx, tx, journalID, filteredUserIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) GetReadingUsers(ctx context.Context, userID string, journalID string) ([]data.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if err := s.validateJournalOwnership(ctx, tx, userID, journalID); err != nil {
		return nil, err
	}

	users, err := s.getReadingUsers(ctx, tx, journalID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s Server) AddReading(ctx context.Context, userID string, journalID string, targetUserIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.validateJournalOwnership(ctx, tx, userID, journalID); err != nil {
		return err
	}

	filteredUserIDs := s.filterCurrentUser(userID, targetUserIDs)
	filteredUserIDs = s.filterDuplicateIds(filteredUserIDs)

	if err := s.addReadingPrivileges(ctx, tx, journalID, filteredUserIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) AddJournal(ctx context.Context, userID string, journalData data.JournalUpdateData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	journalID := journalData.ID
	now := time.Now()

	if err := s.createJournal(ctx, tx, journalID, journalData.Title, now); err != nil {
		return err
	}

	if err := s.createJournalOwnership(ctx, tx, journalID, userID, now); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) UpdateJournal(ctx context.Context, userID string, journalData data.JournalUpdateData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.validateJournalOwnership(ctx, tx, userID, journalData.ID); err != nil {
		return err
	}

	if err := s.updateJournalTitle(ctx, tx, journalData.ID, journalData.Title); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) CreateUserByProvidedID(ctx context.Context, id string, login string, name string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	userID := uuid.Must(uuid.NewV7()).String()
	now := time.Now()

	_, err = tx.ExecContext(ctx, "INSERT INTO user (user_id, provided_id, login, name, created_dt) VALUES (?, ?, ?, ?, ?)", userID, id, login, name, now)
	if err != nil {
		return "", err
	}

	return userID, tx.Commit()
}
