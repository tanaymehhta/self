package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"github.com/tanaymehhta/self/backend/pkg/config"
)

func NewRateLimit(cfg *config.Config) fiber.Handler {
	var max int
	var expiration time.Duration

	if cfg.IsDevelopment() {
		// More lenient in development
		max = 1000
		expiration = time.Minute
	} else {
		// Stricter in production
		max = 100
		expiration = time.Minute
	}

	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit by IP, but use user ID if authenticated
			if userID, exists := GetUserID(c); exists {
				return userID.String()
			}
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "rate_limit_exceeded",
				"message": "Too many requests, please try again later",
			})
		},
	})
}

func NewUploadRateLimit() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        10,              // 10 uploads per hour
		Expiration: time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			if userID, exists := GetUserID(c); exists {
				return "upload:" + userID.String()
			}
			return "upload:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "upload_rate_limit_exceeded",
				"message": "Too many uploads, please try again later",
			})
		},
	})
}