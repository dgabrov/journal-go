package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostReadingUsersHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostReadingUsersHandler(config *data.ConfigData, db *sql.DB) *PostReadingUsersHandler {
	return &PostReadingUsersHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostReadingUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var readingUsersRequest data.StringHolder

	err := parseJsonRequest(r, &readingUsersRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	users, err := h.process(ctx, r, readingUsersRequest)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, users, nil)
}

func (h *PostReadingUsersHandler) process(ctx context.Context, r *http.Request, readingUsersRequest data.StringHolder) ([]data.User, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		return nil, err
	}

	users, err := servr.GetReadingUsers(ctx, userID, readingUsersRequest.Val)
	if err != nil {
		return nil, err
	}

	return users, nil
}
