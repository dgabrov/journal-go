package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
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

	success, err := h.process(ctx, r, journalItemID, token)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostUploadFilesHandler) process(ctx context.Context, r *http.Request, journalItemID string, token string) (*Success, error) {
	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if err := servr.ValidateJournalItemOwnership(ctx, userID, journalItemID); err != nil {
		return nil, err
	}

	// TODO: Implement file processing logic
	// - Extract files from multipart request
	// - Validate file types and sizes
	// - Store files in appropriate folders (based on Files config)
	// - Create file records in database if needed

	return NewSuccess(), nil
}
