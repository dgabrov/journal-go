package controller

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostUpdateAttachmentHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostUpdateAttachmentHandler(config *data.ConfigData, db *sql.DB) *PostUpdateAttachmentHandler {
	return &PostUpdateAttachmentHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostUpdateAttachmentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	attachmentID := r.FormValue("attachmentID")
	if attachmentID == "" {
		writeJsonResponse(w, nil, errors.New("attachmentID is required"))
		return
	}

	title := r.FormValue("title")

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := servr.ValidateAttachmentsOwnership(ctx, userID, []string{attachmentID}); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := h.process(ctx, attachmentID, title, r, servr); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, NewSuccess(), nil)
}

func (h *PostUpdateAttachmentHandler) process(ctx context.Context, attachmentID string, title string, r *http.Request, servr *server.Server) error {
	fileHeaders, hasFile := r.MultipartForm.File["file"]

	if hasFile && len(fileHeaders) > 0 {
		file, err := fileHeaders[0].Open()
		if err != nil {
			return err
		}
		defer file.Close()

		if !isValidImageFile(file) {
			return errors.New("invalid image file")
		}

		if err := h.updateAttachmentFile(ctx, file, attachmentID, servr); err != nil {
			return err
		}
	}

	if title != "" {
		if err := servr.UpdateAttachmentTitle(ctx, attachmentID, title); err != nil {
			slog.Error("failed to update attachment title", "id", attachmentID, "error", err)
			return err
		}
	}

	return nil
}

func (h *PostUpdateAttachmentHandler) updateAttachmentFile(ctx context.Context, file io.ReadSeeker, attachmentID string, servr *server.Server) error {
	contentType, err := saveAttachmentFile(file, attachmentID, h.Config.Files.RegularFolder)
	if err != nil {
		return err
	}

	// Create new thumbnail (old files are cleaned up by createThumbnail)
	createThumbnail(attachmentID, h.Config.Files.RegularFolder, h.Config.Files.Dimension, h.Config.Files.SmallFolder)

	// Update content type in database
	if err := servr.UpdateAttachmentContentType(ctx, attachmentID, contentType); err != nil {
		slog.Error("failed to update attachment content type", "id", attachmentID, "error", err)
		return err
	}

	return nil
}
