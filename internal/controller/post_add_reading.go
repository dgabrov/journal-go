package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostAddReadingHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostAddReadingHandler(config *data.ConfigData, db *sql.DB) *PostAddReadingHandler {
	return &PostAddReadingHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostAddReadingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var addReadingRequest data.JournalUsersRequest

	err := parseJsonRequest(r, &addReadingRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	success, err := h.process(ctx, r, addReadingRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostAddReadingHandler) process(ctx context.Context, r *http.Request, addReadingRequest data.JournalUsersRequest) (*Success, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	err = servr.AddReading(ctx, userID, addReadingRequest.JournalID, addReadingRequest.UserIDs)
	if err != nil {
		return nil, err
	}

	return NewSuccess(), nil
}
