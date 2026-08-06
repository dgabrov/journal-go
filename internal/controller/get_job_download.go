package controller

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type GetJobDownloadHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewGetJobDownloadHandler(config *data.ConfigData, db *sql.DB) *GetJobDownloadHandler {
	return &GetJobDownloadHandler{
		Config: config,
		DB:     db,
	}
}

func (h *GetJobDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	token, err := getToken(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		http.Error(w, "missing id parameter", http.StatusBadRequest)
		return
	}

	jobExists, err := servr.JobExists(ctx, jobID)
	if err != nil {
		http.Error(w, "error checking job", http.StatusInternalServerError)
		return
	}
	if !jobExists {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	jobBelongsToUser, err := servr.JobBelongsToUser(ctx, jobID, userID)
	if err != nil {
		http.Error(w, "error checking job ownership", http.StatusInternalServerError)
		return
	}
	if !jobBelongsToUser {
		http.Error(w, "job does not belong to user", http.StatusForbidden)
		return
	}

	isCompleted, err := servr.JobIsCompleted(ctx, jobID)
	if err != nil {
		http.Error(w, "error checking job status", http.StatusInternalServerError)
		return
	}
	if !isCompleted {
		http.Error(w, "job is not completed", http.StatusBadRequest)
		return
	}

	zipPath := filepath.Join(h.Config.Files.JobFolder, jobID+".zip")
	file, err := os.Open(zipPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Error("Job zip file not found", "jobID", jobID, "path", zipPath)
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		slog.Error("error opening job file", "jobID", jobID, "error", err)
		http.Error(w, "error reading file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	respHeader := w.Header()
	respHeader.Set("Content-Type", "application/zip")
	respHeader.Set("Content-Disposition", "attachment; filename="+jobID+".zip")
	respHeader.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	respHeader.Set("Pragma", "no-cache")
	respHeader.Set("Expires", "0")

	if _, err := file.WriteTo(w); err != nil {
		slog.Error("error writing file to response", "jobID", jobID, "error", err)
	}
}
