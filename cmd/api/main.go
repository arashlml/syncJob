package main

import (
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"

	"practice/internals/handler"
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

	userRepo := postgres.NewUserRepository(conn, logger)

	userService := service.NewUserService(userRepo)

	userHandler := handler.New(
		userService,
		logger,
	)

	e := echo.New()

	userHandler.RegisterRoutes(e)

	logger.Info("http server started",
		slog.String("address", ":8080"),
	)

	if err := e.Start(":8080"); err != nil {
		logger.Error("server stopped",
			slog.Any("error", err),
		)
	}
}
