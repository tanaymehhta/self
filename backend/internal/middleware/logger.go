package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/tanaymehhta/self/backend/pkg/logger"
)

func NewRequestLogger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("X-Request-ID", requestID)

		// Start timer
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user ID if available
		var userID string
		if uid, exists := GetUserID(c); exists {
			userID = uid.String()
		}

		// Log request
		log.WithRequest(requestID, userID).LogHTTP(
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			duration,
			"ip", c.IP(),
			"user_agent", c.Get("User-Agent"),
			"request_size", len(c.Body()),
			"response_size", len(c.Response().Body()),
		)

		return err
	}
}