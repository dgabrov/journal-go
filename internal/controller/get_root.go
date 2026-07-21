package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
)

type RootHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewRootHandler(config *data.ConfigData, db *sql.DB) *RootHandler {
	return &RootHandler{
		Config: config,
		DB:     db,
	}
}

func (h *RootHandler) process(ctx context.Context) (map[string]string, error) {
	payload := make(map[string]string)
	payload["application"] = "Journal"
	payload["message"] = "all good"

	return payload, nil
}

func (h *RootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	payload, err := h.process(ctx)

	writeJsonResponse(w, payload, err)
}
