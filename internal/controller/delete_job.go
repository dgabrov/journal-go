package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type DeleteJobHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewDeleteJobHandler(config *data.ConfigData, db *sql.DB) *DeleteJobHandler {
	return &DeleteJobHandler{
		Config: config,
		DB:     db,
	}
}

func (h *DeleteJobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	jobIDs := r.URL.Query()["id"]
	if len(jobIDs) == 0 {
		writeJsonResponse(w, nil, NewErrorResponse("no job ids provided"))
		return
	}

	distinctJobIDs := consolidateJobIDs(jobIDs)

	err = h.process(ctx, servr, userID, distinctJobIDs)

	success := NewSuccess()
	writeJsonResponse(w, success, err)
}

func (h *DeleteJobHandler) process(ctx context.Context, servr *server.Server, userID string, jobIDs []string) error {
	if err := servr.ValidateJobsExist(ctx, jobIDs); err != nil {
		return err
	}

	if err := servr.ValidateJobsBelongToUser(ctx, jobIDs, userID); err != nil {
		return err
	}

	return servr.DeleteJobs(ctx, jobIDs, h.Config)
}

func consolidateJobIDs(jobIDs []string) []string {
	seen := make(map[string]bool)
	distinct := make([]string, 0)

	for _, id := range jobIDs {
		if !seen[id] {
			seen[id] = true
			distinct = append(distinct, id)
		}
	}

	return distinct
}

func NewErrorResponse(message string) error {
	return NewError(message)
}

type deleteError struct {
	msg string
}

func NewError(msg string) error {
	return &deleteError{msg: msg}
}

func (e *deleteError) Error() string {
	return e.msg
}
