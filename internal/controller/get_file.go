package controller

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

//go:embed placeholder.png
var placeholderImage []byte

type GetFileHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewGetFileHandler(config *data.ConfigData, db *sql.DB) *GetFileHandler {
	return &GetFileHandler{
		Config: config,
		DB:     db,
	}
}

func (h *GetFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		slog.Error("getToken error: " + err.Error())
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, false)
	if err != nil {
		slog.Error("getUserIdFromToken error: " + err.Error())
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	if err := servr.ValidateAttachmentAccess(ctx, userID, id); err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	small := r.URL.Query().Get("small")

	contentType, err := servr.GetAttachmentContentType(ctx, id)
	if err != nil {
		contentType = "image/png"
	}

	if small != "" {
		h.serveSmallFile(w, id, contentType)
	} else {
		h.serveRegularFile(w, id, contentType)
	}
}

func (h *GetFileHandler) setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func (h *GetFileHandler) serveSmallFile(w http.ResponseWriter, id string, contentType string) {
	filename := id + ".dat"
	filepath := filepath.Join(h.Config.Files.SmallFolder, filename)

	file, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			h.setNoCacheHeaders(w)
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(http.StatusOK)
			w.Write(placeholderImage)
			return
		}
		slog.Error("error opening small file", "id", id, "error", err)
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	h.setNoCacheHeaders(w)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if _, err := file.WriteTo(w); err != nil {
		slog.Error("error writing file to response", "id", id, "error", err)
	}
}

func (h *GetFileHandler) serveRegularFile(w http.ResponseWriter, id string, contentType string) {
	filename := id + ".dat"
	filepath := filepath.Join(h.Config.Files.RegularFolder, filename)

	file, err := os.Open(filepath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		slog.Error("error opening regular file", "id", id, "error", err)
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	h.setNoCacheHeaders(w)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if _, err := file.WriteTo(w); err != nil {
		slog.Error("error writing file to response", "id", id, "error", err)
	}
}
