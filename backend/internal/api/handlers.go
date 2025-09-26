package api

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"github.com/tanaymehhta/self/backend/internal/auth"
	"github.com/tanaymehhta/self/backend/internal/middleware"
	"github.com/tanaymehhta/self/backend/internal/models"
	"github.com/tanaymehhta/self/backend/internal/services"
)

// Health check handler
func (s *Server) healthHandler(c *fiber.Ctx) error {
	// Check database health
	if err := s.db.Health(); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unhealthy",
			"database": "disconnected",
		})
	}

	return c.JSON(fiber.Map{
		"status":   "healthy",
		"database": "connected",
		"version":  "1.0.0",
	})
}

// Authentication handlers
func (s *Server) loginHandler(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Use local auth service
	authService := auth.NewLocalAuthService(s.db.DB, s.jwtManager)
	response, err := authService.Login(auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "authentication_failed",
			"message": err.Error(),
		})
	}

	return c.JSON(response)
}

func (s *Server) registerHandler(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		FullName string `json:"full_name"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Use local auth service
	authService := auth.NewLocalAuthService(s.db.DB, s.jwtManager)
	response, err := authService.Register(auth.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "registration_failed",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (s *Server) refreshTokenHandler(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	tokenPair, err := s.jwtManager.RefreshTokenPair(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "invalid_refresh_token",
			"message": "Invalid or expired refresh token",
		})
	}

	return c.JSON(tokenPair)
}

// User handlers
func (s *Server) getUserProfileHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "user_not_found",
			"message": "User not found",
		})
	}

	return c.JSON(user)
}

func (s *Server) updateUserProfileHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	var req struct {
		FullName    *string           `json:"full_name"`
		AvatarURL   *string           `json:"avatar_url"`
		Preferences models.JSONB      `json:"preferences"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Update user
	updates := map[string]interface{}{}
	if req.FullName != nil {
		updates["full_name"] = *req.FullName
	}
	if req.AvatarURL != nil {
		updates["avatar_url"] = *req.AvatarURL
	}
	if req.Preferences != nil {
		updates["preferences"] = req.Preferences
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "update_failed",
			"message": "Failed to update user profile",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Profile updated successfully",
	})
}

func (s *Server) deleteUserHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	// TODO: Implement user deletion with cascade
	return c.JSON(fiber.Map{
		"message": "Delete user endpoint - TODO: Implement with cascade deletion",
		"user_id": userID,
	})
}

// Conversation handlers
func (s *Server) getConversationsHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	var conversations []models.Conversation
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&conversations).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "query_failed",
			"message": "Failed to fetch conversations",
		})
	}

	return c.JSON(fiber.Map{
		"conversations": conversations,
		"count":        len(conversations),
	})
}

func (s *Server) createConversationHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	var req struct {
		Title       *string      `json:"title"`
		AudioFormat string       `json:"audio_format"`
		Metadata    models.JSONB `json:"metadata"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	conversation := models.Conversation{
		UserID:      userID,
		Title:       req.Title,
		AudioFormat: req.AudioFormat,
		Status:      "processing",
		Metadata:    req.Metadata,
	}

	if err := s.db.Create(&conversation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "creation_failed",
			"message": "Failed to create conversation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(conversation)
}

func (s *Server) getConversationHandler(c *fiber.Ctx) error {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	conversationID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_id",
			"message": "Invalid conversation ID",
		})
	}

	var conversation models.Conversation
	if err := s.db.Preload("Transcriptions").Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "conversation_not_found",
			"message": "Conversation not found",
		})
	}

	return c.JSON(conversation)
}

func (s *Server) updateConversationHandler(c *fiber.Ctx) error {
	// TODO: Implement conversation update
	return c.JSON(fiber.Map{
		"message": "Update conversation endpoint - TODO: Implement",
	})
}

func (s *Server) deleteConversationHandler(c *fiber.Ctx) error {
	// TODO: Implement conversation deletion
	return c.JSON(fiber.Map{
		"message": "Delete conversation endpoint - TODO: Implement",
	})
}

func (s *Server) getTranscriptionsHandler(c *fiber.Ctx) error {
	// TODO: Implement transcriptions retrieval
	return c.JSON(fiber.Map{
		"message": "Get transcriptions endpoint - TODO: Implement",
	})
}

// File handlers
func (s *Server) getFileEventsHandler(c *fiber.Ctx) error {
	// TODO: Implement file events retrieval
	return c.JSON(fiber.Map{
		"message": "Get file events endpoint - TODO: Implement",
	})
}

func (s *Server) createFileEventHandler(c *fiber.Ctx) error {
	// TODO: Implement file event creation
	return c.JSON(fiber.Map{
		"message": "Create file event endpoint - TODO: Implement",
	})
}

func (s *Server) getFileEventHandler(c *fiber.Ctx) error {
	// TODO: Implement single file event retrieval
	return c.JSON(fiber.Map{
		"message": "Get file event endpoint - TODO: Implement",
	})
}

// Audio handlers
func (s *Server) uploadAudioHandler(c *fiber.Ctx) error {
	// TODO: Implement audio upload with MinIO
	return c.JSON(fiber.Map{
		"message": "Upload audio endpoint - TODO: Implement with MinIO",
	})
}

func (s *Server) getAudioHandler(c *fiber.Ctx) error {
	// TODO: Implement audio file retrieval
	return c.JSON(fiber.Map{
		"message": "Get audio endpoint - TODO: Implement",
	})
}

func (s *Server) transcribeAudioHandler(c *fiber.Ctx) error {
	// TODO: Implement audio transcription trigger
	return c.JSON(fiber.Map{
		"message": "Transcribe audio endpoint - TODO: Implement with AI service",
	})
}

// Text processing handlers
func (s *Server) uploadDocumentHandler(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("document")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_file",
			"message": "No document file provided",
		})
	}

	// Check file size (max 1GB)
	maxSize := int64(1024 * 1024 * 1024)
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "file_too_large",
			"message": "File size exceeds 1GB limit",
		})
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "file_read_error",
			"message": "Failed to read uploaded file",
		})
	}
	defer fileContent.Close()

	// Get user ID from auth context
	userID := c.Locals("user_id").(uuid.UUID)

	// Process document
	textPipeline := services.NewTextPipeline(s.db.DB)
	contentItem, err := textPipeline.ProcessDocument(userID, fileContent, file)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "processing_failed",
			"message": fmt.Sprintf("Failed to process document: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Document uploaded and processing started",
		"content_id":  contentItem.ID,
		"title":       contentItem.Title,
		"content_type": contentItem.ContentType,
		"file_size":   contentItem.FileSize,
	})
}

// Search handlers
func (s *Server) searchHandler(c *fiber.Ctx) error {
	var req struct {
		Query string `json:"query" validate:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	// Create search service (old chunk-based search)
	searchService := services.NewSearchService(s.db.DB, nil) // nil for backward compatibility
	results, err := searchService.Search(req.Query, req.Limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "search_failed",
			"message": fmt.Sprintf("Search failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"query":   req.Query,
		"results": results.Results,
		"total":   len(results.Results),
		"strategy": results.Strategy,
	})
}

// QA Search handler - New answer-based search
func (s *Server) qaSearchHandler(c *fiber.Ctx) error {
	var req struct {
		Query string `json:"query" validate:"required"`
		Limit int    `json:"limit"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	if req.Limit == 0 {
		req.Limit = 5 // Fewer answers than chunks by default
	}

	// Create Claude client using config
	if s.config.ClaudeAPIKey == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "service_unavailable",
			"message": "Claude API key not configured",
		})
	}

	claudeClient := services.NewClaudeClient(s.config.ClaudeAPIKey, "claude-3-haiku-20240307")

	// Create answer extraction service
	answerService := services.NewAnswerExtractionService(claudeClient)

	// Create search service with answer extraction
	searchService := services.NewSearchService(s.db.DB, answerService)

	// Perform QA search
	results, err := searchService.QASearch(c.Context(), req.Query, req.Limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "qa_search_failed",
			"message": fmt.Sprintf("QA search failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"query":    req.Query,
		"answers":  results.Answers,
		"total":    results.Total,
		"strategy": results.Strategy,
	})
}

// Get content items handler
func (s *Server) getContentItemsHandler(c *fiber.Ctx) error {
	contentType := c.Query("type", "") // filter by type
	limit := c.QueryInt("limit", 20)

	var items []services.ContentItem
	query := s.db.DB.Order("created_at DESC").Limit(limit)

	if contentType != "" {
		query = query.Where("content_type = ?", contentType)
	}

	err := query.Find(&items).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "database_error",
			"message": "Failed to fetch content items",
		})
	}

	return c.JSON(fiber.Map{
		"items": items,
		"total": len(items),
	})
}

// Get content item details with chunks
func (s *Server) getContentItemHandler(c *fiber.Ctx) error {
	itemID := c.Params("id")

	var item services.ContentItem
	err := s.db.DB.First(&item, "id = ?", itemID).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "not_found",
			"message": "Content item not found",
		})
	}

	// Get chunks for this item
	var chunks []services.Chunk
	err = s.db.DB.Where("content_item_id = ?", itemID).Order("chunk_index").Find(&chunks).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "database_error",
			"message": "Failed to fetch chunks",
		})
	}

	return c.JSON(fiber.Map{
		"item":   item,
		"chunks": chunks,
	})
}

// Test Pipeline handler - uploads document with detailed step logging
func (s *Server) testPipelineHandler(c *fiber.Ctx) error {
	// Get file from form
	file, err := c.FormFile("document")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "no_file",
			"message": "No document file provided",
		})
	}

	// Check file size (max 1GB)
	maxSize := int64(1024 * 1024 * 1024)
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "file_too_large",
			"message": "File size exceeds 1GB limit",
		})
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "file_read_error",
			"message": "Failed to read uploaded file",
		})
	}
	defer fileContent.Close()

	// Get user ID from auth context
	userID := c.Locals("user_id").(uuid.UUID)

	// Create pipeline logger for detailed tracking
	logger := services.NewPipelineLogger()

	// Process document with detailed logging
	textPipeline := services.NewTextPipeline(s.db.DB)
	contentItem, err := textPipeline.ProcessDocumentWithLogging(userID, fileContent, file, logger)

	// Always return the pipeline logs, even if processing failed
	logger.Complete()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "processing_failed",
			"message": fmt.Sprintf("Failed to process document: %v", err),
			"pipeline_log": fiber.Map{
				"steps":   logger.Steps,
				"summary": logger.GetSummary(),
			},
		})
	}

	// Return success with detailed pipeline information
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Document processed successfully with detailed logging",
		"content_id":  contentItem.ID,
		"title":       contentItem.Title,
		"content_type": contentItem.ContentType,
		"file_size":   contentItem.FileSize,
		"pipeline_log": fiber.Map{
			"steps":   logger.Steps,
			"summary": logger.GetSummary(),
		},
	})
}

func (s *Server) semanticSearchHandler(c *fiber.Ctx) error {
	// TODO: Implement semantic search with embeddings
	return c.JSON(fiber.Map{
		"message": "Semantic search endpoint - TODO: Implement with vectors",
	})
}

// Entity handlers
func (s *Server) getEntitiesHandler(c *fiber.Ctx) error {
	// TODO: Implement entities retrieval
	return c.JSON(fiber.Map{
		"message": "Get entities endpoint - TODO: Implement",
	})
}

func (s *Server) createEntityHandler(c *fiber.Ctx) error {
	// TODO: Implement entity creation
	return c.JSON(fiber.Map{
		"message": "Create entity endpoint - TODO: Implement",
	})
}

func (s *Server) getEntityHandler(c *fiber.Ctx) error {
	// TODO: Implement single entity retrieval
	return c.JSON(fiber.Map{
		"message": "Get entity endpoint - TODO: Implement",
	})
}

func (s *Server) updateEntityHandler(c *fiber.Ctx) error {
	// TODO: Implement entity update
	return c.JSON(fiber.Map{
		"message": "Update entity endpoint - TODO: Implement",
	})
}

func (s *Server) deleteEntityHandler(c *fiber.Ctx) error {
	// TODO: Implement entity deletion
	return c.JSON(fiber.Map{
		"message": "Delete entity endpoint - TODO: Implement",
	})
}

// Insight handlers
func (s *Server) getInsightsHandler(c *fiber.Ctx) error {
	// TODO: Implement insights retrieval
	return c.JSON(fiber.Map{
		"message": "Get insights endpoint - TODO: Implement",
	})
}

func (s *Server) acknowledgeInsightHandler(c *fiber.Ctx) error {
	// TODO: Implement insight acknowledgment
	return c.JSON(fiber.Map{
		"message": "Acknowledge insight endpoint - TODO: Implement",
	})
}

func (s *Server) dismissInsightHandler(c *fiber.Ctx) error {
	// TODO: Implement insight dismissal
	return c.JSON(fiber.Map{
		"message": "Dismiss insight endpoint - TODO: Implement",
	})
}

// Integration handlers
func (s *Server) getIntegrationsHandler(c *fiber.Ctx) error {
	// TODO: Implement integrations retrieval
	return c.JSON(fiber.Map{
		"message": "Get integrations endpoint - TODO: Implement",
	})
}

func (s *Server) connectIntegrationHandler(c *fiber.Ctx) error {
	// TODO: Implement integration connection
	return c.JSON(fiber.Map{
		"message": "Connect integration endpoint - TODO: Implement OAuth flow",
	})
}

func (s *Server) disconnectIntegrationHandler(c *fiber.Ctx) error {
	// TODO: Implement integration disconnection
	return c.JSON(fiber.Map{
		"message": "Disconnect integration endpoint - TODO: Implement",
	})
}

func (s *Server) syncIntegrationHandler(c *fiber.Ctx) error {
	// TODO: Implement integration sync
	return c.JSON(fiber.Map{
		"message": "Sync integration endpoint - TODO: Implement",
	})
}

// WebSocket handlers
func (s *Server) transcriptionWebSocketHandler(c *websocket.Conn) {
	// TODO: Implement real-time transcription WebSocket
	defer c.Close()

	for {
		// Read message from client
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			s.logger.LogError(err, "WebSocket read error")
			break
		}

		s.logger.Debug("WebSocket message received", "type", messageType, "message", string(msg))

		// Echo back for now - TODO: Implement real transcription streaming
		if err := c.WriteMessage(messageType, []byte("Transcription WebSocket - TODO: Implement real-time processing")); err != nil {
			s.logger.LogError(err, "WebSocket write error")
			break
		}
	}
}

func (s *Server) insightsWebSocketHandler(c *websocket.Conn) {
	// TODO: Implement real-time insights WebSocket
	defer c.Close()

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			s.logger.LogError(err, "Insights WebSocket read error")
			break
		}

		s.logger.Debug("Insights WebSocket message", "type", messageType, "message", string(msg))

		if err := c.WriteMessage(messageType, []byte("Insights WebSocket - TODO: Implement real-time insights")); err != nil {
			s.logger.LogError(err, "Insights WebSocket write error")
			break
		}
	}
}

// ============================================================================
// CHAT HANDLERS
// ============================================================================

// Create new chat conversation
func (s *Server) createChatConversationHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	// For now, just create the conversation record directly
	conversation := models.ChatConversation{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        nil,
		MessageCount: 0,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.DB.Create(&conversation).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "database_error",
			"message": "Failed to create conversation",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"conversation": conversation,
	})
}

// Send message in chat conversation
func (s *Server) sendChatMessageHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	var req services.ChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	if req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Message cannot be empty",
		})
	}

	// Create search service instance
	searchService := services.NewSearchService(
		s.db.DB,
		services.NewAnswerExtractionService(services.NewClaudeClient(s.config.ClaudeAPIKey, "claude-3-haiku-20240307")),
	)

	// Create chat service
	chatService := services.NewChatService(s.db.DB, searchService)

	// Process the message
	response, err := chatService.ProcessMessage(c.Context(), userID, req)
	if err != nil {
		s.logger.LogError(err, "Failed to process chat message")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "processing_failed",
			"message": fmt.Sprintf("Failed to process message: %v", err),
		})
	}

	return c.JSON(response)
}

// Get chat conversations for user
func (s *Server) getChatConversationsHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)
	limit := c.QueryInt("limit", 20)

	// Create search service instance
	searchService := services.NewSearchService(
		s.db.DB,
		services.NewAnswerExtractionService(services.NewClaudeClient(s.config.ClaudeAPIKey, "claude-3-haiku-20240307")),
	)

	// Create chat service
	chatService := services.NewChatService(s.db.DB, searchService)

	conversations, err := chatService.GetConversations(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "database_error",
			"message": "Failed to fetch conversations",
		})
	}

	return c.JSON(fiber.Map{
		"conversations": conversations,
		"total":         len(conversations),
	})
}

// Get messages for a chat conversation
func (s *Server) getChatMessagesHandler(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	conversationIDStr := c.Params("id")
	conversationID, err := uuid.Parse(conversationIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid conversation ID",
		})
	}

	// Create search service instance
	searchService := services.NewSearchService(
		s.db.DB,
		services.NewAnswerExtractionService(services.NewClaudeClient(s.config.ClaudeAPIKey, "claude-3-haiku-20240307")),
	)

	// Create chat service
	chatService := services.NewChatService(s.db.DB, searchService)

	messages, err := chatService.GetConversationMessages(userID, conversationID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "database_error",
			"message": "Failed to fetch messages",
		})
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"total":    len(messages),
	})
}