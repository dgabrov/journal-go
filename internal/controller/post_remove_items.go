package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostRemoveItemsHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostRemoveItemsHandler(config *data.ConfigData, db *sql.DB) *PostRemoveItemsHandler {
	return &PostRemoveItemsHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostRemoveItemsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var idsHolder data.IdsHolder

	err := parseJsonRequest(r, &idsHolder)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	success, err := h.process(ctx, r, idsHolder)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostRemoveItemsHandler) process(ctx context.Context, r *http.Request, idsHolder data.IdsHolder) (*Success, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return nil, err
	}

	err = servr.RemoveItems(ctx, userID, idsHolder.Ids)
	if err != nil {
		return nil, err
	}

	return NewSuccess(), nil
}
