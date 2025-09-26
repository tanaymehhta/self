package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/tanaymehhta/self/backend/internal/auth"
	"github.com/tanaymehhta/self/backend/internal/database"
	"github.com/tanaymehhta/self/backend/internal/middleware"
	"github.com/tanaymehhta/self/backend/pkg/config"
	"github.com/tanaymehhta/self/backend/pkg/logger"
)

type Server struct {
	app        *fiber.App
	db         *database.DB
	config     *config.Config
	logger     *logger.Logger
	jwtManager *auth.JWTManager
	auth       *middleware.AuthMiddleware
}

func NewServer(
	db *database.DB,
	cfg *config.Config,
	logger *logger.Logger,
	jwtManager *auth.JWTManager,
) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler(logger),
		BodyLimit:    int(cfg.MaxFileSize),
		AppName:      "Self API v1.0",
		ServerHeader: "Self",
	})

	authMiddleware := middleware.NewAuthMiddleware(jwtManager, logger)

	server := &Server{
		app:        app,
		db:         db,
		config:     cfg,
		logger:     logger,
		jwtManager: jwtManager,
		auth:       authMiddleware,
	}

	server.setupMiddleware()
	server.setupRoutes()

	return server
}

func (s *Server) setupMiddleware() {
	// Global middleware
	s.app.Use(middleware.NewCORS(s.config))
	s.app.Use(middleware.NewRequestLogger(s.logger))
	s.app.Use(middleware.NewRateLimit(s.config))
}

func (s *Server) setupRoutes() {
	// Health check
	s.app.Get("/health", s.healthHandler)

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Public routes (no auth required)
	s.setupPublicRoutes(v1)

	// Protected routes (auth required)
	protected := v1.Group("", s.auth.RequireAuth())
	s.setupProtectedRoutes(protected)

	// WebSocket routes
	s.setupWebSocketRoutes()
}

func (s *Server) setupPublicRoutes(router fiber.Router) {
	// Auth routes
	auth := router.Group("/auth")
	auth.Post("/login", s.loginHandler)
	auth.Post("/register", s.registerHandler)
	auth.Post("/refresh", s.refreshTokenHandler)
}

func (s *Server) setupProtectedRoutes(router fiber.Router) {
	// User routes
	users := router.Group("/users")
	users.Get("/me", s.getUserProfileHandler)
	users.Put("/me", s.updateUserProfileHandler)
	users.Delete("/me", s.deleteUserHandler)

	// Text processing routes
	text := router.Group("/text")
	text.Post("/upload", s.uploadDocumentHandler)
	text.Post("/test-upload", s.testPipelineHandler) // New testing endpoint with detailed logging
	text.Post("/search", s.searchHandler)
	text.Get("/items", s.getContentItemsHandler)
	text.Get("/items/:id", s.getContentItemHandler)

	// Conversation routes
	conversations := router.Group("/conversations")
	conversations.Get("/", s.getConversationsHandler)
	conversations.Post("/", s.createConversationHandler)
	conversations.Get("/:id", s.getConversationHandler)
	conversations.Put("/:id", s.updateConversationHandler)
	conversations.Delete("/:id", s.deleteConversationHandler)
	conversations.Get("/:id/transcriptions", s.getTranscriptionsHandler)

	// File routes
	files := router.Group("/files")
	files.Get("/events", s.getFileEventsHandler)
	files.Post("/events", s.createFileEventHandler)
	files.Get("/events/:id", s.getFileEventHandler)

	// Audio routes with upload rate limiting
	audio := router.Group("/audio", middleware.NewUploadRateLimit())
	audio.Post("/upload", s.uploadAudioHandler)
	audio.Get("/:id", s.getAudioHandler)
	audio.Post("/transcribe", s.transcribeAudioHandler)

	// Search routes
	search := router.Group("/search")
	search.Get("/", s.searchHandler)
	search.Post("/qa", s.qaSearchHandler) // New QA-based search endpoint
	search.Post("/semantic", s.semanticSearchHandler)

	// Entity routes
	entities := router.Group("/entities")
	entities.Get("/", s.getEntitiesHandler)
	entities.Post("/", s.createEntityHandler)
	entities.Get("/:id", s.getEntityHandler)
	entities.Put("/:id", s.updateEntityHandler)
	entities.Delete("/:id", s.deleteEntityHandler)

	// Insight routes
	insights := router.Group("/insights")
	insights.Get("/", s.getInsightsHandler)
	insights.Put("/:id/acknowledge", s.acknowledgeInsightHandler)
	insights.Delete("/:id", s.dismissInsightHandler)

	// Integration routes
	integrations := router.Group("/integrations")
	integrations.Get("/", s.getIntegrationsHandler)
	integrations.Post("/:service/connect", s.connectIntegrationHandler)
	integrations.Delete("/:service", s.disconnectIntegrationHandler)
	integrations.Post("/:service/sync", s.syncIntegrationHandler)

	// Chat routes
	chat := router.Group("/chat")
	chat.Post("/conversations", s.createChatConversationHandler)
	chat.Post("/conversations/:id/message", s.sendChatMessageHandler)
	chat.Get("/conversations", s.getChatConversationsHandler)
	chat.Get("/conversations/:id/messages", s.getChatMessagesHandler)
}

func (s *Server) setupWebSocketRoutes() {
	// WebSocket upgrade middleware
	s.app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket routes
	s.app.Get("/ws/transcription/:conversationId", websocket.New(s.transcriptionWebSocketHandler))
	s.app.Get("/ws/insights", websocket.New(s.insightsWebSocketHandler))
}

func (s *Server) Listen(port string) error {
	s.logger.Info("Starting server", "port", port, "env", s.config.Env)
	return s.app.Listen(":" + port)
}

func (s *Server) Shutdown() error {
	s.logger.Info("Shutting down server...")
	return s.app.Shutdown()
}

// Error handler
func errorHandler(logger *logger.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to 500 server error
		code := fiber.StatusInternalServerError
		message := "Internal Server Error"

		// If it's a Fiber error, get the code and message
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
			message = e.Message
		}

		// Log error
		logger.LogError(err, "HTTP error",
			"status", code,
			"path", c.Path(),
			"method", c.Method(),
		)

		// Return error response
		return c.Status(code).JSON(fiber.Map{
			"error":   "server_error",
			"message": message,
		})
	}
}