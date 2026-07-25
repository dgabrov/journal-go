package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostUpdateTitlesHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostUpdateTitlesHandler(config *data.ConfigData, db *sql.DB) *PostUpdateTitlesHandler {
	return &PostUpdateTitlesHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostUpdateTitlesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	var payload data.Titles
	if err := parseJsonRequest(r, &payload); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	if err := h.process(ctx, userID, payload.Titles, servr); err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, NewSuccess(), nil)
}

func (h *PostUpdateTitlesHandler) process(ctx context.Context, userID string, titleHolders []data.TitleHolder, servr *server.Server) error {
	if err := servr.ValidateAttachmentOwnership(ctx, userID, titleHolders); err != nil {
		return err
	}

	if err := servr.UpdateAttachmentTitles(ctx, titleHolders); err != nil {
		return err
	}

	return nil
}
