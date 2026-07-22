package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostRemoveReadingHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostRemoveReadingHandler(config *data.ConfigData, db *sql.DB) *PostRemoveReadingHandler {
	return &PostRemoveReadingHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostRemoveReadingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var removeReadingRequest data.JournalUsersRequest

	err := parseJsonRequest(r, &removeReadingRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	success, err := h.process(ctx, r, removeReadingRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostRemoveReadingHandler) process(ctx context.Context, r *http.Request, removeReadingRequest data.JournalUsersRequest) (*Success, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	err = servr.RemoveReading(ctx, userID, removeReadingRequest.JournalID, removeReadingRequest.UserIDs)
	if err != nil {
		return nil, err
	}

	return NewSuccess(), nil
}
