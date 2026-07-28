package controller

import (
	"context"
	"database/sql"
	"errors"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
	"github.com/google/uuid"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
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

	go h.processThumbnails(attachmentIDs, h.Config.Files.RegularFolder, h.Config.Files.Dimension, h.Config.Files.SmallFolder)

	return &data.IdsHolder{Ids: attachmentIDs}, nil
}

func (h *PostUploadFilesHandler) processValidFile(ctx context.Context, file io.ReadSeeker, journalItemID string, servr *server.Server) (string, error) {
	attachmentID := uuid.Must(uuid.NewV7()).String()
	filename := attachmentID + ".dat"

	contentType, err := getImageContentType(file)
	if err != nil {
		return "", err
	}

	finalPath := filepath.Join(h.Config.Files.RegularFolder, filename)
	finalFile, err := os.Create(finalPath)
	if err != nil {
		return "", err
	}
	defer finalFile.Close()

	if _, err := io.Copy(finalFile, file); err != nil {
		os.Remove(finalPath)
		return "", err
	}

	width, height, err := getImageDimensions(finalPath)
	if err != nil {
		os.Remove(finalPath)
		return "", err
	}

	// by default, when created, the title is empty string
	err = servr.CreateAttachment(ctx, journalItemID, attachmentID, contentType, width, height)
	if err != nil {
		os.Remove(finalPath)
		return "", err
	}

	return attachmentID, nil
}

func (h *PostUploadFilesHandler) processThumbnails(attachmentIDs []string, regularFolder string, dimension int, smallFolder string) {
	for _, id := range attachmentIDs {
		createThumbnail(id, regularFolder, dimension, smallFolder)
	}
}

func getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return config.Width, config.Height, nil
}
