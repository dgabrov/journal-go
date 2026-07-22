package service

import (
	"database/sql"
	"net/http"

	"github.com/amanagement24/journal-go/internal/controller"
	"github.com/amanagement24/journal-go/internal/data"
)

func SetupRouter(config *data.ConfigData, db *sql.DB) *http.ServeMux {
	mux := http.NewServeMux()
	context := config.Context

	mux.Handle(fullUrl(true, context, "/"), controller.NewRootHandler(config, db))
	mux.Handle(fullUrl(false, context, "/login"), controller.NewLoginHandler(config, db))
	mux.Handle(fullUrl(false, context, "/editJournalItem"), controller.NewPostEditJournalItemHandler(config, db))
	mux.Handle(fullUrl(false, context, "/addJournalItem"), controller.NewPostAddJournalItemHandler(config, db))
	mux.Handle(fullUrl(false, context, "/searchUsers"), controller.NewPostSearchUsersHandler(config, db))
	mux.Handle(fullUrl(false, context, "/logout"), controller.NewPostLogoutHandler(config, db))
	mux.Handle(fullUrl(false, context, "/removeItems"), controller.NewPostRemoveItemsHandler(config, db))
	mux.Handle(fullUrl(false, context, "/removeJournals"), controller.NewPostRemoveJournalsHandler(config, db))
	mux.Handle(fullUrl(false, context, "/journalItems"), controller.NewPostJournalItemsHandler(config, db))
	mux.Handle(fullUrl(false, context, "/removeReading"), controller.NewPostRemoveReadingHandler(config, db))
	mux.Handle(fullUrl(false, context, "/readingUsers"), controller.NewPostReadingUsersHandler(config, db))
	mux.Handle(fullUrl(false, context, "/addReadingUsers"), controller.NewPostAddReadingHandler(config, db))
	mux.Handle(fullUrl(false, context, "/addJournal"), controller.NewPostAddJournalHandler(config, db))
	mux.Handle(fullUrl(false, context, "/updateJournal"), controller.NewPostUpdateJournalHandler(config, db))

	return mux
}

func fullUrl(get bool, context string, url string) string {
	prefix := "POST "
	if get {
		prefix = "GET "
	}

	return prefix + context + url
}
