package job

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/amanagement24/journal-go/internal/data"
	"github.com/amanagement24/journal-go/internal/server"
)

func StartJobWithContext(ctx context.Context, config *data.ConfigData, db *sql.DB, jobEventChan chan int) {
	timeoutTicker := time.NewTicker(5 * time.Minute)
	defer timeoutTicker.Stop()

	srv := server.New(db, config)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Job processor shutting down")
			return

		case <-jobEventChan:
			drainJobEventChan(jobEventChan)
			processPendingJobs(ctx, srv, db, config)

		case <-timeoutTicker.C:
			drainJobEventChan(jobEventChan)
			processPendingJobs(ctx, srv, db, config)
		}
	}
}

func processPendingJobs(ctx context.Context, srv *server.Server, db *sql.DB, config *data.ConfigData) {
	s := server.New(db, config)
	pendingJobs, err := s.GetPendingJobs(ctx)
	if err != nil {
		slog.Error("Failed to get pending jobs", "error", err)
		return
	}

	for _, jobID := range pendingJobs {
		if err := srv.ProcessJob(ctx, jobID); err != nil {
			slog.Error("Failed to process job", "jobID", jobID, "error", err)
		} else {
			slog.Info("Job processed successfully", "jobID", jobID)
		}
	}
}

func drainJobEventChan(ch chan int) {
	for {
		select {
		case <-ch:
		default:
			return
		}
	}
}
