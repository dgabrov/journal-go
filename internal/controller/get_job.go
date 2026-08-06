package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type GetJobHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewGetJobHandler(config *data.ConfigData, db *sql.DB) *GetJobHandler {
	return &GetJobHandler{
		Config: config,
		DB:     db,
	}
}

func (h *GetJobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	jobs, err := servr.GetUserJobs(ctx, userID)

	writeJsonResponse(w, jobs, err)
}
