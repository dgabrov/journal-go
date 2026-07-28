package controller

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostAddJournalHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostAddJournalHandler(config *data.ConfigData, db *sql.DB) *PostAddJournalHandler {
	return &PostAddJournalHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostAddJournalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var journalData data.JournalUpdateData

	err := parseJsonRequest(r, &journalData)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	success, err := h.process(ctx, r, journalData)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostAddJournalHandler) process(ctx context.Context, r *http.Request, journalData data.JournalUpdateData) (*Success, error) {
	if journalData.Title == "" {
		return nil, errors.New("title is required")
	}

	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return nil, err
	}

	err = servr.AddJournal(ctx, userID, journalData)
	if err != nil {
		return nil, err
	}

	return NewSuccess(), nil
}
