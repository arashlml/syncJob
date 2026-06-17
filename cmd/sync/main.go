package main

import (
	"log/slog"
	"os"

	"practice/internals/repository/file"
	"practice/internals/repository/postgres"
	"practice/internals/service"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		}),
	)

	conn, err := postgres.Connect()
	if err != nil {
		logger.Error("database connection failed",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
	defer conn.Close()

	jsonIterator, err := file.NewJSONFileIterator(
		"./users_data.json",
	)
	if err != nil {
		logger.Error("iterator initialization failed",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	repo := postgres.NewUserRepository(
		conn,
		logger,
	)

	syncService := service.NewSyncService(
		jsonIterator,
		repo,
		10,  // workers
		100, // channel size
		"user-sync",
	)

	syncService.Start()
	syncService.Wait()

	logger.Info("sync completed successfully")
}
