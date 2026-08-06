package service

import (
	"log/slog"
	"net/http"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/job"

	_ "github.com/go-sql-driver/mysql"
)

func Start() error {
	jobEventChan := make(chan int, 1000)

	// Load configuration
	config, err := data.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		return err
	}

	// Connect to database
	db, err := data.ConnectDB(config.DB)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return err
	}
	defer db.Close()

	job.StartJob(config, db, jobEventChan)

	// Start HTTP server
	router := SetupRouter(config, db, jobEventChan)
	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}

	slog.Info("Starting HTTP server", "address", config.ServerAddress, "context", config.Context)
	if err := server.ListenAndServe(); err != nil {
		slog.Error("HTTP server failed", "error", err)
		return err
	}

	return nil
}
