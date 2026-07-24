package controller

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
	"github.com/google/uuid"
)

type PostUploadFilesHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostUploadFilesHandler(config *data.ConfigData, db *sql.DB) *PostUploadFilesHandler {
	return &PostUploadFilesHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostUploadFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	journalItemID := r.FormValue("journalItemId")
	if journalItemID == "" {
		writeJsonResponse(w, nil, errors.New("journalItemId is required"))
		return
	}

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	ids, err := h.process(ctx, r, journalItemID, token)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, ids, nil)
}

func (h *PostUploadFilesHandler) process(ctx context.Context, r *http.Request, journalItemID string, token string) (*data.IdsHolder, error) {
	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if err := servr.ValidateJournalItemOwnership(ctx, userID, journalItemID); err != nil {
		return nil, err
	}

	attachmentIDs := make([]string, 0)

	for _, fileHeaders := range r.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			file, err := fileHeader.Open()
			if err != nil {
				return nil, err
			}
			defer file.Close()

			isValid := isValidImageFile(file)

			if !isValid {
				continue
			}

			attachmentID, err := h.processValidFile(ctx, file, journalItemID, servr)
			if err != nil {
				return nil, err
			}

			attachmentIDs = append(attachmentIDs, attachmentID)
		}
	}

	return &data.IdsHolder{Ids: attachmentIDs}, nil
}

func (h *PostUploadFilesHandler) processValidFile(ctx context.Context, file io.ReadSeeker, journalItemID string, servr *server.Server) (string, error) {
	guid := uuid.Must(uuid.NewV7()).String()
	filename := guid + ".dat"

	if err := os.MkdirAll(h.Config.Files.TempFolder, 0755); err != nil {
		return "", err
	}

	tempPath := filepath.Join(h.Config.Files.TempFolder, filename)
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	if err := os.MkdirAll(h.Config.Files.RegularFolder, 0755); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	finalPath := filepath.Join(h.Config.Files.RegularFolder, filename)
	if err := os.Rename(tempPath, finalPath); err != nil {
		os.Remove(tempPath)
		return "", err
	}

	attachmentID, err := servr.CreateAttachment(ctx, journalItemID, filename)
	if err != nil {
		os.Remove(finalPath)
		return "", err
	}

	return attachmentID, nil
}

func isValidImageFile(file io.ReadSeeker) bool {
	if _, err := file.Seek(0, 0); err != nil {
		return false
	}

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return false
	}

	if _, err := file.Seek(0, 0); err != nil {
		return false
	}

	isJPEG := len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	return isJPEG || isPNG
}
