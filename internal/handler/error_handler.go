package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/yourusername/user-management-api/internal/models"
)

func NewErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		statusCode := fiber.StatusInternalServerError
		message := "internal server error"

		var fiberError *fiber.Error
		if errors.As(err, &fiberError) {
			statusCode = fiberError.Code
			message = fiberError.Message
		}

		logFields := []zap.Field{
			zap.Int("status", statusCode),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("request_id", requestIDFromCtx(c)),
			zap.Error(err),
		}

		if statusCode >= fiber.StatusInternalServerError {
			logger.Error("request failed", logFields...)
		} else {
			logger.Warn("request rejected", logFields...)
		}

		return c.Status(statusCode).JSON(models.ErrorResponse{
			Error: message,
		})
	}
}

func requestIDFromCtx(c *fiber.Ctx) string {
	value := c.Locals("request_id")
	requestID, _ := value.(string)
	return requestID
}
