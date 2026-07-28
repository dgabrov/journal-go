package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostAddJournalItemHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostAddJournalItemHandler(config *data.ConfigData, db *sql.DB) *PostAddJournalItemHandler {
	return &PostAddJournalItemHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostAddJournalItemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (h *PostAddJournalItemHandler) process(ctx context.Context, r *http.Request, item data.JournalItem) error {
	token, err := getToken(r)
	if err != nil {
		return err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return err
	}

	return servr.AddJournalItem(ctx, userID, item)
}
