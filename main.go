package main

import (
	"github.com/amanagement24/journal-go/internal/service"
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	err := service.Start()

	if err != nil {
		slog.Error(err.Error())
	}
}
