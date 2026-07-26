package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostJournalItemsHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostJournalItemsHandler(config *data.ConfigData, db *sql.DB) *PostJournalItemsHandler {
	return &PostJournalItemsHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostJournalItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var stringHolder data.StringHolder

	err := parseJsonRequest(r, &stringHolder)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	items, err := h.process(ctx, r, stringHolder)

	writeJsonResponse(w, items, err)
}

func (h *PostJournalItemsHandler) process(ctx context.Context, r *http.Request, stringHolder data.StringHolder) ([]data.CompleteJournalItem, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	return servr.GetJournalItems(ctx, userID, stringHolder.Val)
}
