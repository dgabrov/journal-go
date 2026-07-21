package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostEditJournalItemHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostEditJournalItemHandler(config *data.ConfigData, db *sql.DB) *PostEditJournalItemHandler {
	return &PostEditJournalItemHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostEditJournalItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var journalItem data.JournalItem

	err := parseJsonRequest(r, &journalItem)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	err = h.process(ctx, r, journalItem)

	success := NewSuccess()

	writeJsonResponse(w, success, err)
}

func (h *PostEditJournalItemHandler) process(ctx context.Context, r *http.Request, item data.JournalItem) error {
	token, err := getToken(r)
	if err != nil {
		return err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return err
	}

	return servr.UpdateJournalItem(ctx, userID, item)
}
