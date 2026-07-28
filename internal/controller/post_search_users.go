package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostSearchUsersHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostSearchUsersHandler(config *data.ConfigData, db *sql.DB) *PostSearchUsersHandler {
	return &PostSearchUsersHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostSearchUsersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var searchData data.SearchData

	err := parseJsonRequest(r, &searchData)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	users, err := h.process(ctx, r, searchData)

	writeJsonResponse(w, users, err)
}

func (h *PostSearchUsersHandler) process(ctx context.Context, r *http.Request, searchData data.SearchData) ([]data.User, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return nil, err
	}

	return servr.SearchUsers(ctx, userID, searchData.Search)
}
