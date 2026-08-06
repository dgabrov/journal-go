package service

import (
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/controller"
	"github.com/amanagement24/journal-go/internal/data"
)

func SetupRouter(config *data.ConfigData, db *sql.DB, jobEventChan chan int) *http.ServeMux {
	mux := http.NewServeMux()
	context := config.Context

	mux.Handle(fullUrlWithMethod("GET", context, "/"), controller.NewRootHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/login"), controller.NewLoginHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/editJournalItem"), controller.NewPostEditJournalItemHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/addJournalItem"), controller.NewPostAddJournalItemHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/searchUsers"), controller.NewPostSearchUsersHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/logout"), controller.NewPostLogoutHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/removeItems"), controller.NewPostRemoveItemsHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/removeJournals"), controller.NewPostRemoveJournalsHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/journalItems"), controller.NewPostJournalItemsHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/removeReading"), controller.NewPostRemoveReadingHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/readingUsers"), controller.NewPostReadingUsersHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/addReadingUsers"), controller.NewPostAddReadingHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/addJournal"), controller.NewPostAddJournalHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/updateJournal"), controller.NewPostUpdateJournalHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/uploadFiles"), controller.NewPostUploadFilesHandler(config, db))
	mux.Handle(fullUrlWithMethod("GET", context, "/getFile"), controller.NewGetFileHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/removeFiles"), controller.NewPostRemoveFilesHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/updateTitles"), controller.NewPostUpdateTitlesHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/updateAttachment"), controller.NewPostUpdateAttachmentHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/updateAttachmentTitles"), controller.NewPostUpdateAttachmentTitlesHandler(config, db))
	mux.Handle(fullUrlWithMethod("PUT", context, "/rotate"), controller.NewPutRotateHandler(config, db))
	mux.Handle(fullUrlWithMethod("POST", context, "/job"), controller.NewPostJobHandler(config, db, jobEventChan))
	mux.Handle(fullUrlWithMethod("GET", context, "/job"), controller.NewGetJobHandler(config, db))
	mux.Handle(fullUrlWithMethod("GET", context, "/job/download"), controller.NewGetJobDownloadHandler(config, db))
	mux.Handle(fullUrlWithMethod("DELETE", context, "/job"), controller.NewDeleteJobHandler(config, db))

	return mux
}

func fullUrlWithMethod(method string, context string, url string) string {
	return method + " " + context + url
}
