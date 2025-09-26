package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/tanaymehhta/self/backend/pkg/config"
)

func NewCORS(cfg *config.Config) fiber.Handler {
	var allowOrigins string
	var allowCredentials bool

	if cfg.IsDevelopment() {
		// Allow all origins in development
		allowOrigins = "*"
		allowCredentials = false
	} else {
		// Restrict origins in production
		allowOrigins = "https://yourdomain.com,https://app.yourdomain.com"
		allowCredentials = true
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		ExposeHeaders:    "Content-Length,Content-Range",
		AllowCredentials: allowCredentials,
		MaxAge:           86400, // 24 hours
	})
}