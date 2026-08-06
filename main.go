package main

import (
	"log/slog"
	"os"

	"github.com/amanagement24/journal-go/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	err := service.Start()

	if err != nil {
		slog.Error(err.Error())
	}
}
