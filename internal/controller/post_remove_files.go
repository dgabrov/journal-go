package controller

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostRemoveFilesHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostRemoveFilesHandler(config *data.ConfigData, db *sql.DB) *PostRemoveFilesHandler {
	return &PostRemoveFilesHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostRemoveFilesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	var payload data.IdsHolder
	if err := parseJsonRequest(r, &payload); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := h.process(ctx, userID, payload.Ids, servr); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, NewSuccess(), nil)
}

func (h *PostRemoveFilesHandler) process(ctx context.Context, userID string, attachmentIDs []string, servr *server.Server) error {
	if err := servr.ValidateAttachmentsOwnership(ctx, userID, attachmentIDs); err != nil {
		return err
	}

	if err := servr.DeleteAttachmentFiles(attachmentIDs); err != nil {
		return err
	}

	if err := servr.DeleteAttachments(ctx, attachmentIDs); err != nil {
		slog.Error("failed to delete attachments from database", "error", err)
		return err
	}

	return nil
}
