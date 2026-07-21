package controller

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostLogoutHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostLogoutHandler(config *data.ConfigData, db *sql.DB) *PostLogoutHandler {
	return &PostLogoutHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostLogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		slog.Error("logout: token not found", "error", err)
		writeJsonResponse(w, NewSuccess(), nil)
		return
	}

	servr := server.New(h.DB, h.Config)
	err = servr.LogoutUser(ctx, token)
	if err != nil {
		slog.Error("logout: failed to logout user", "error", err)
		writeJsonResponse(w, NewSuccess(), nil)
		return
	}

	slog.Info("logout: user logged out successfully")
	writeJsonResponse(w, NewSuccess(), nil)
}
