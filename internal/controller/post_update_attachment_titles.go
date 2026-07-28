package controller

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

type PostUpdateAttachmentTitlesHandler struct {
	Config *data.ConfigData
	DB     *sql.DB
}

func NewPostUpdateAttachmentTitlesHandler(config *data.ConfigData, db *sql.DB) *PostUpdateAttachmentTitlesHandler {
	return &PostUpdateAttachmentTitlesHandler{
		Config: config,
		DB:     db,
	}
}

func (h *PostUpdateAttachmentTitlesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var attachments []data.Attachment

	err := parseJsonRequest(r, &attachments)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	success, err := h.process(ctx, r, attachments)
	if err != nil {
		writeJsonResponse(w, nil, err)
		return
	}

	writeJsonResponse(w, success, nil)
}

func (h *PostUpdateAttachmentTitlesHandler) process(ctx context.Context, r *http.Request, attachments []data.Attachment) (*Success, error) {
	token, err := getToken(r)
	if err != nil {
		return nil, err
	}

	servr := server.New(h.DB, h.Config)
	userID, err := servr.GetUserIdFromToken(ctx, token, true)
	if err != nil {
		return nil, err
	}

	attachmentIDs := make([]string, len(attachments))
	for i, a := range attachments {
		attachmentIDs[i] = a.Id
	}

	if err := servr.ValidateAttachmentsOwnership(ctx, userID, attachmentIDs); err != nil {
		return nil, err
	}

	titleHolders := make([]data.TitleHolder, len(attachments))
	for i, a := range attachments {
		titleHolders[i] = data.TitleHolder{
			AttachmentID: a.Id,
			Title:        a.Title,
		}
	}

	err = servr.UpdateAttachmentTitles(ctx, titleHolders)
	if err != nil {
		return nil, err
	}

	return NewSuccess(), nil
}
