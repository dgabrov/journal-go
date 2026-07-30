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

func New(db *sql.DB, config *data.ConfigData) *Server {
	return &Server{
		db:     db,
		config: config,
	}
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

func (s Server) GetUserIdFromToken(ctx context.Context, token string, advanceExpiry bool) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	userID, expiredInd, expireDt, err := s.getSessionByToken(ctx, tx, token)
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
	if advanceExpiry {
		newExpiry := time.Now().Add(time.Duration(s.config.TokenTimeToLive) * time.Second)
		if err := s.updateSessionExpiry(ctx, tx, token, newExpiry); err != nil {
			return "", err
		}
	}

	return userID, tx.Commit()
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
	relationCd, err := s.checkJournalOwnership(ctx, tx, userID, item.JournalID)
	if err != nil {
		return err
	}

	if relationCd != data.RelationOwner {
		return errors.New("user is not the owner of this journal")
	}

	// Update the journal item with comments and current timestamp
	now := time.Now()
	if err := s.updateJournalItemComments(ctx, tx, item.Id, item.Comments, now); err != nil {
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
	relationCd, err := s.checkJournalOwnership(ctx, tx, userID, item.JournalID)
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
	if err := s.insertJournalItem(ctx, tx, id, item.JournalID, createdDate, updatedDate, item.Comments); err != nil {
		return err
	}

	return tx.Commit()
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

	if err := s.insertUser(ctx, tx, userID, id, login, name, now); err != nil {
		return "", err
	}

	return userID, tx.Commit()
}

func (s Server) ValidateAttachmentAccess(ctx context.Context, userID string, attachmentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	count, err := s.validateAttachmentAccessQuery(ctx, tx, attachmentID, userID)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("attachment not found or user does not have access")
	}

	return tx.Commit()
}

func (s Server) GetAttachmentContentType(ctx context.Context, attachmentID string) (string, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	contentType, err := s.getAttachmentContentTypeQuery(ctx, tx, attachmentID)
	if err != nil {
		return "image/png", nil
	}

	if err := tx.Commit(); err != nil {
		return "image/png", nil
	}

	return contentType, nil
}

func (s Server) ValidateJournalItemOwnership(ctx context.Context, userID string, journalItemID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	count, err := s.validateJournalItemOwnershipQuery(ctx, tx, userID, journalItemID)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("user does not own this journal item")
	}

	return tx.Commit()
}

func (s Server) CreateAttachment(ctx context.Context, journalItemID string, attachmentID string, contentType string, width int, height int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	if err := s.insertAttachment(ctx, tx, attachmentID, journalItemID, contentType, width, height, now); err != nil {
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

	count, err := s.validateAttachmentsOwnershipQuery(ctx, tx, attachmentIDs, userID)
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

	if err := s.deleteAttachmentsQuery(ctx, tx, attachmentIDs); err != nil {
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

func (s Server) UpdateAttachmentTitles(ctx context.Context, titleHolders []data.TitleHolder) error {
	if len(titleHolders) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.updateAttachmentTitlesQuery(ctx, tx, titleHolders); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) UpdateAttachmentTitle(ctx context.Context, attachmentID string, title string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.updateAttachmentTitleQuery(ctx, tx, attachmentID, title); err != nil {
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

	if err := s.updateAttachmentContentTypeQuery(ctx, tx, attachmentID, contentType); err != nil {
		return err
	}

	return tx.Commit()
}

func (s Server) SwapAttachmentDimensions(ctx context.Context, attachmentID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.swapAttachmentDimensions(ctx, tx, attachmentID); err != nil {
		return err
	}

	return tx.Commit()
}
