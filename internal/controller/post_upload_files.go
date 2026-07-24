package controller

import (
	"context"
	"database/sql"
	"errors"
	"image"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
	"github.com/disintegration/imaging"
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

	go h.processThumbnails(attachmentIDs)

	return &data.IdsHolder{Ids: attachmentIDs}, nil
}

func (h *PostUploadFilesHandler) processValidFile(ctx context.Context, file io.ReadSeeker, journalItemID string, servr *server.Server) (string, error) {
	guid := uuid.Must(uuid.NewV7()).String()
	filename := guid + ".dat"

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

func (h *PostUploadFilesHandler) processThumbnails(attachmentIDs []string) {
	for _, id := range attachmentIDs {
		h.createThumbnail(id)
	}
}

func (h *PostUploadFilesHandler) createThumbnail(id string) {
	filename := id + ".dat"
	sourcePath := filepath.Join(h.Config.Files.RegularFolder, filename)

	img, err := imaging.Open(sourcePath)
	if err != nil {
		slog.Error("failed to open image for thumbnail", "id", id, "error", err)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maxDimension := h.Config.Files.Dimension

	if width <= maxDimension && height <= maxDimension {
		return
	}

	var scaledImg image.Image
	if width > height {
		scaledImg = imaging.Resize(img, maxDimension, 0, imaging.Lanczos)
	} else {
		scaledImg = imaging.Resize(img, 0, maxDimension, imaging.Lanczos)
	}

	destPath := filepath.Join(h.Config.Files.SmallFolder, filename)
	if err := imaging.Save(scaledImg, destPath); err != nil {
		slog.Error("failed to save thumbnail", "id", id, "error", err)
		return
	}
}
