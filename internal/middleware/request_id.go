package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDLocalKey contextKey = "request_id"
	requestIDHeader              = "X-Request-ID"
)

func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Locals(string(RequestIDLocalKey), requestID)
		c.Set(requestIDHeader, requestID)
		c.SetUserContext(context.WithValue(c.UserContext(), RequestIDLocalKey, requestID))

		return c.Next()
	}
}
