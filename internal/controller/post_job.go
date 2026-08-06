package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostJobHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostJobHandler(config *data.ConfigData, db *sql.DB) *PostJobHandler {
	return &PostJobHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostJobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var holder data.JournalItemHolder

	err := parseJsonRequest(r, &holder)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	err = h.process(ctx, r, holder)

	success := NewSuccess()

	writeJsonResponse(w, success, err)
}

func (h *PostJobHandler) process(ctx context.Context, r *http.Request, holder data.JournalItemHolder) error {
	token, err := getToken(r)
	if err != nil {
		return err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return err
	}

	return servr.CreateJob(ctx, userID, holder.JournalItemID)
}
