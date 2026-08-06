package service

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start job processor in a goroutine
	go job.StartJobWithContext(ctx, config, db, jobEventChan)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server
	router := SetupRouter(config, db, jobEventChan)
	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}

	// Start server in a goroutine
	serverErrChan := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server", "address", config.ServerAddress, "context", config.Context)
		serverErrChan <- server.ListenAndServe()
	}()

	// Wait for signal or server error
	select {
	case sig := <-sigChan:
		slog.Info("Received signal, initiating graceful shutdown", "signal", sig)
		cancel()
		server.Shutdown(context.Background())
		return nil
	case err := <-serverErrChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server error", "error", err)
			return err
		}
		return nil
	}
}
