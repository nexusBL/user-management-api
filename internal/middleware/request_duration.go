package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func RequestDuration(logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startedAt := time.Now().UTC()

		err := c.Next()

		duration := time.Since(startedAt)
		requestID, _ := c.Locals(string(RequestIDLocalKey)).(string)

		logger.Info("http request completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.String("url", c.OriginalURL()),
			zap.Int("status", c.Response().StatusCode()),
			zap.String("ip", c.IP()),
			zap.Duration("duration", duration),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)

		return err
	}
}
