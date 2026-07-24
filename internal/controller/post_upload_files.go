package controller

import (
	"context"
	"database/sql"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
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

	err = servr.CreateAttachment(ctx, journalItemID, attachmentID, filename, contentType)
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
	slog.Info("starting processing thumbnail", slog.String("id", id))
	sourcePath := filepath.Join(h.Config.Files.RegularFolder, filename)

	ext, err := getImageExtension(sourcePath)
	if err != nil {
		slog.Error("failed to detect image extension", "id", id, "error", err)
		return
	}

	img, err := imaging.Open(sourcePath)
	if err != nil {
		slog.Error("failed to open image for thumbnail", "id", id, "error", err)
		return
	}

	slog.Info("opened source path for thumbnail processing")

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	maxDimension := h.Config.Files.Dimension

	destPath := filepath.Join(h.Config.Files.SmallFolder, filename)
	_ = os.Remove(destPath)

	if width <= maxDimension && height <= maxDimension {
		sourceFile, err := os.Open(sourcePath)
		if err != nil {
			slog.Error("failed to open source file for copying", "id", id, "error", err)
			return
		}
		defer sourceFile.Close()

		destFile, err := os.Create(destPath)
		if err != nil {
			slog.Error("failed to create destination file", "id", id, "error", err)
			return
		}
		defer destFile.Close()

		if _, err := io.Copy(destFile, sourceFile); err != nil {
			slog.Error("failed to copy file to small folder", "id", id, "error", err)
			return
		}

		slog.Info("thumbnail copied (no resizing needed)", "id", id)
		return
	}

	var scaledImg image.Image
	if width > height {
		scaledImg = imaging.Resize(img, maxDimension, 0, imaging.Lanczos)
	} else {
		scaledImg = imaging.Resize(img, 0, maxDimension, imaging.Lanczos)
	}

	slog.Info("thumbnail resizing complete")

	tempPath := filepath.Join(h.Config.Files.SmallFolder, id+ext)
	if err := imaging.Save(scaledImg, tempPath); err != nil {
		slog.Error("failed to save thumbnail", "id", id, "error", err)
		return
	}

	if err := os.Rename(tempPath, destPath); err != nil {
		slog.Error("failed to rename thumbnail to .dat", "id", id, "error", err)
		os.Remove(tempPath)
		return
	}

	slog.Info("thumbnail created", "id", id)
}

func getImageExtension(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	isJPEG := len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	if isJPEG {
		return ".jpg", nil
	}
	if isPNG {
		return ".png", nil
	}

	return "", errors.New("unknown image type")
}

func getImageContentType(file io.ReadSeeker) (string, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	header := make([]byte, 12)
	if _, err := file.Read(header); err != nil {
		return "", err
	}

	if _, err := file.Seek(0, 0); err != nil {
		return "", err
	}

	isJPEG := len(header) >= 2 && header[0] == 0xFF && header[1] == 0xD8
	isPNG := len(header) >= 8 && header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47

	if isJPEG {
		return "image/jpeg", nil
	}
	if isPNG {
		return "image/png", nil
	}

	return "", errors.New("unknown image type")
}
