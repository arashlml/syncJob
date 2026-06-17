package handler

import (
	"errors"
	"log/slog"
	"practice/internals/entity"
	"practice/internals/params"
	"practice/internals/params/errs"

	"github.com/labstack/echo/v4"
)

type handler struct {
	service Service
	logger  *slog.Logger
}

func New(service Service, logger *slog.Logger) *handler {
	return &handler{
		service: service,
		logger:  logger,
	}
}

type Service interface {
	GetUserByID(userID params.UserID) (*entity.User, error)
}

func (h *handler) getByID(c echo.Context) error {
	id := c.Param("id")

	logger := h.logger.With(
		slog.String("handler", "UserHandler"),
		slog.String("method", "getByID"),
		slog.String("user_id", id),
	)

	logger.Info("request received")

	user, err := h.service.GetUserByID(
		params.UserID{ID: id},
	)

	if err != nil {

		switch {

		case errors.Is(err, errs.ErrUserNotFound):
			logger.Warn("user not found")

			return c.JSON(404, map[string]string{
				"message": "User not found",
			})

		default:
			logger.Error(
				"unexpected error",
				slog.Any("error", err),
			)

			return c.JSON(500, map[string]string{
				"message": "Internal Server Error",
			})
		}
	}

	logger.Info("request completed successfully")

	return c.JSON(200, user)
}

func (h *handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/users/:id", h.getByID)
}
