package controller

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
	"github.com/disintegration/imaging"
)

type RotateRequest struct {
	ID       string `json:"id"`
	Quotient int    `json:"quotient"`
}

type PutRotateHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPutRotateHandler(config *data.ConfigData, db *sql.DB) *PutRotateHandler {
	return &PutRotateHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PutRotateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req RotateRequest
	if err := parseJsonRequest(r, &req); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if req.ID == "" {
		writeJsonResponse(w, nil, errors.New("id is required"))
		return
	}

	if req.Quotient != -1 && req.Quotient != 1 {
		writeJsonResponse(w, nil, errors.New("quotient must be -1 or 1"))
		return
	}

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := servr.ValidateAttachmentsOwnership(ctx, userID, []string{req.ID}); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := h.rotateAttachment(ctx, servr, req.ID, req.Quotient); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	// this time we do it in sync to ensure that the image is ready by the time the server returns
	createThumbnail(req.ID, h.Config.Files.RegularFolder, h.Config.Files.Dimension, h.Config.Files.SmallFolder)

	writeJsonResponse(w, NewSuccess(), nil)
}

func (h *PutRotateHandler) rotateAttachment(ctx context.Context, servr *server.Server, attachmentID string, quotient int) error {
	// Get the attachment's content type from the service layer
	contentType, err := servr.GetAttachmentContentType(ctx, attachmentID)
	if err != nil {
		slog.Error("failed to get attachment content type", "id", attachmentID, "error", err)
		return err
	}

	ext := contentTypeToExtension(contentType)

	// Paths
	sourceFile := attachmentID + ".dat"
	sourcePath := filepath.Join(h.Config.Files.RegularFolder, sourceFile)
	tempFolder := os.TempDir()

	// Create temp file with correct extension for loading
	loadTempPath := filepath.Join(tempFolder, attachmentID+"_load"+ext)
	if err := copyFile(sourcePath, loadTempPath); err != nil {
		slog.Error("failed to copy file for loading", "id", attachmentID, "error", err)
		return err
	}
	defer os.Remove(loadTempPath)

	// Load and rotate
	img, err := imaging.Open(loadTempPath)
	if err != nil {
		slog.Error("failed to open image for rotation", "id", attachmentID, "error", err)
		return err
	}

	angle := float64(quotient) * 90
	rotatedImg := imaging.Rotate(img, -angle, nil) // not sure but positive angle is to the left not to the right so changed sign

	// Save to temp file with correct extension
	rotateTempPath := filepath.Join(tempFolder, attachmentID+"_rotate"+ext)
	if err := imaging.Save(rotatedImg, rotateTempPath); err != nil {
		slog.Error("failed to save rotated image", "id", attachmentID, "error", err)
		os.Remove(rotateTempPath)
		return err
	}
	defer os.Remove(rotateTempPath)

	// Copy rotated image back to .dat file
	if err := copyFile(rotateTempPath, sourcePath); err != nil {
		slog.Error("failed to copy rotated image back", "id", attachmentID, "error", err)
		return err
	}

	// Swap width and height since the image was rotated
	if err := servr.SwapAttachmentDimensions(ctx, attachmentID); err != nil {
		slog.Error("failed to swap attachment dimensions", "id", attachmentID, "error", err)
		return err
	}

	slog.Info("attachment rotated", "id", attachmentID, "angle", angle)
	return nil
}
