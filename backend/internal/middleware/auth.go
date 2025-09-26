package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/tanaymehhta/self/backend/internal/auth"
	"github.com/tanaymehhta/self/backend/pkg/logger"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
	logger     *logger.Logger
}

// ContextKey type for context keys
type ContextKey string

const (
	UserIDKey ContextKey = "user_id"
	ClaimsKey ContextKey = "claims"
)

func NewAuthMiddleware(jwtManager *auth.JWTManager, logger *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
		logger:     logger.WithComponent("auth_middleware"),
	}
}

// RequireAuth middleware that requires valid JWT authentication
func (m *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing authorization header", "path", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Authorization header required",
			})
		}

		// Extract token from header
		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			m.logger.Warn("Invalid authorization header", "error", err.Error(), "path", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid authorization header format",
			})
		}

		// Validate token
		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			m.logger.Warn("Invalid access token", "error", err.Error(), "path", c.Path())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid or expired token",
			})
		}

		// Parse user ID
		userID, err := auth.GetUserIDFromClaims(claims)
		if err != nil {
			m.logger.Error("Invalid user ID in token", "error", err.Error(), "user_id", claims.UserID)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthorized",
				"message": "Invalid user ID in token",
			})
		}

		// Store user information in context
		c.Locals(string(UserIDKey), userID)
		c.Locals(string(ClaimsKey), claims)

		// Add user ID to logger context for this request
		reqLogger := m.logger.WithUser(claims.UserID)
		c.Locals("logger", reqLogger)

		return c.Next()
	}
}

// OptionalAuth middleware that optionally validates JWT if present
func (m *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		tokenString, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Next()
		}

		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			return c.Next()
		}

		userID, err := auth.GetUserIDFromClaims(claims)
		if err != nil {
			return c.Next()
		}

		// Store user information in context
		c.Locals(string(UserIDKey), userID)
		c.Locals(string(ClaimsKey), claims)

		reqLogger := m.logger.WithUser(claims.UserID)
		c.Locals("logger", reqLogger)

		return c.Next()
	}
}

// GetUserID extracts user ID from Fiber context
func GetUserID(c *fiber.Ctx) (uuid.UUID, bool) {
	userID := c.Locals(string(UserIDKey))
	if userID == nil {
		return uuid.Nil, false
	}

	if uid, ok := userID.(uuid.UUID); ok {
		return uid, true
	}

	return uuid.Nil, false
}

// GetClaims extracts JWT claims from Fiber context
func GetClaims(c *fiber.Ctx) (*auth.Claims, bool) {
	claims := c.Locals(string(ClaimsKey))
	if claims == nil {
		return nil, false
	}

	if cl, ok := claims.(*auth.Claims); ok {
		return cl, true
	}

	return nil, false
}

// GetLogger extracts request logger from Fiber context
func GetLogger(c *fiber.Ctx) *logger.Logger {
	if l := c.Locals("logger"); l != nil {
		if logger, ok := l.(*logger.Logger); ok {
			return logger
		}
	}
	// Fallback to basic logger
	return &logger.Logger{}
}