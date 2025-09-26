package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tanaymehhta/self/backend/internal/models"
)

type ChatService struct {
	db            *gorm.DB
	searchService *SearchService
}

// ChatRequest represents an incoming chat message
type ChatRequest struct {
	ConversationID *uuid.UUID `json:"conversation_id,omitempty"`
	Message        string     `json:"message"`
	DocumentIDs    []uuid.UUID `json:"document_ids,omitempty"` // Optional document context
}

// ChatResponse represents the complete chat response
type ChatResponse struct {
	ConversationID uuid.UUID                `json:"conversation_id"`
	MessageID      uuid.UUID                `json:"message_id"`
	Response       string                   `json:"response"`
	Sources        []AnswerResult           `json:"sources"`
	Confidence     *float64                 `json:"confidence,omitempty"`
	Documents      []ChatDocumentReference  `json:"documents,omitempty"`
}

// ChatDocumentReference provides context about documents used
type ChatDocumentReference struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Type     string    `json:"type"`
	Relevant bool      `json:"relevant"` // Whether this doc contributed to the answer
}

// ConversationHistory represents the context of previous messages
type ConversationHistory struct {
	Messages []HistoryMessage `json:"messages"`
}

type HistoryMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func NewChatService(db *gorm.DB, searchService *SearchService) *ChatService {
	return &ChatService{
		db:            db,
		searchService: searchService,
	}
}

// ProcessMessage is the main entry point for chat functionality
func (cs *ChatService) ProcessMessage(ctx context.Context, userID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	// 1. Get or create conversation
	conversation, err := cs.getOrCreateConversation(userID, req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get/create conversation: %w", err)
	}

	// 2. Save user message
	_, err = cs.saveMessage(conversation.ID, "user", req.Message, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// 3. Get conversation context for better queries
	history, err := cs.getConversationHistory(conversation.ID, 5) // Last 5 messages
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// 4. Enhance query with context (resolve pronouns, add context)
	enhancedQuery := cs.enhanceQueryWithContext(req.Message, history)

	// 5. Perform QA search using existing pipeline
	qaResults, err := cs.searchService.QASearch(ctx, enhancedQuery, 5)
	if err != nil {
		return nil, fmt.Errorf("QA search failed: %w", err)
	}

	// 6. Format answer for chat context
	chatResponse, confidence := cs.formatChatResponse(qaResults, req.Message)

	// 7. Prepare sources for response
	sources := cs.prepareSources(qaResults.Answers)

	// 8. Save AI response
	aiMessage, err := cs.saveMessage(conversation.ID, "assistant", chatResponse, sources, confidence)
	if err != nil {
		return nil, fmt.Errorf("failed to save AI message: %w", err)
	}

	// 9. Get document references
	documents, err := cs.getConversationDocuments(conversation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation documents: %w", err)
	}

	return &ChatResponse{
		ConversationID: conversation.ID,
		MessageID:      aiMessage.ID,
		Response:       chatResponse,
		Sources:        sources,
		Confidence:     confidence,
		Documents:      documents,
	}, nil
}

// getOrCreateConversation handles conversation management
func (cs *ChatService) getOrCreateConversation(userID uuid.UUID, conversationID *uuid.UUID) (*models.ChatConversation, error) {
	if conversationID != nil {
		// Fetch existing conversation
		var conversation models.ChatConversation
		err := cs.db.Where("id = ? AND user_id = ?", *conversationID, userID).First(&conversation).Error
		if err == nil {
			return &conversation, nil
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
	}

	// Create new conversation
	conversation := models.ChatConversation{
		ID:           uuid.New(),
		UserID:       userID,
		Title:        nil, // Will be auto-generated later
		MessageCount: 0,
		LastActivity: time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := cs.db.Create(&conversation).Error; err != nil {
		return nil, err
	}

	return &conversation, nil
}

// saveMessage saves a chat message to the database
func (cs *ChatService) saveMessage(conversationID uuid.UUID, role, content string, sources []AnswerResult, confidence *float64) (*models.ChatMessage, error) {
	sourcesJSON := models.JSONB{}
	if sources != nil {
		sourcesBytes, _ := json.Marshal(sources)
		json.Unmarshal(sourcesBytes, &sourcesJSON)
	}

	message := models.ChatMessage{
		ID:             uuid.New(),
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		Sources:        sourcesJSON,
		Confidence:     confidence,
		Metadata:       models.JSONB{},
		CreatedAt:      time.Now(),
	}

	if err := cs.db.Create(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

// getConversationHistory retrieves recent conversation context
func (cs *ChatService) getConversationHistory(conversationID uuid.UUID, limit int) (*ConversationHistory, error) {
	var messages []models.ChatMessage
	err := cs.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// Reverse to get chronological order
	history := ConversationHistory{Messages: make([]HistoryMessage, len(messages))}
	for i, msg := range messages {
		history.Messages[len(messages)-1-i] = HistoryMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt,
		}
	}

	return &history, nil
}

// enhanceQueryWithContext improves queries using conversation history
func (cs *ChatService) enhanceQueryWithContext(query string, history *ConversationHistory) string {
	if history == nil || len(history.Messages) == 0 {
		return query
	}

	// Simple context enhancement (can be made more sophisticated)
	// This helps resolve pronouns and provide context
	contextualQuery := query

	// If the query has pronouns or references, add recent context
	pronouns := []string{"it", "this", "that", "they", "them", "he", "she", "his", "her"}
	hasPronouns := false
	lowerQuery := strings.ToLower(query)

	for _, pronoun := range pronouns {
		if strings.Contains(lowerQuery, pronoun) {
			hasPronouns = true
			break
		}
	}

	if hasPronouns && len(history.Messages) > 0 {
		// Get the last user message for context
		for i := len(history.Messages) - 1; i >= 0; i-- {
			if history.Messages[i].Role == "user" {
				contextualQuery = fmt.Sprintf("Previous context: %s\n\nCurrent question: %s",
					history.Messages[i].Content, query)
				break
			}
		}
	}

	return contextualQuery
}

// formatChatResponse converts QA results into a natural chat response
func (cs *ChatService) formatChatResponse(qaResults *QASearchResults, originalQuery string) (string, *float64) {
	if len(qaResults.Answers) == 0 {
		return "I couldn't find relevant information to answer your question. Could you try rephrasing or provide more context?", nil
	}

	// Get the best answer
	bestAnswer := qaResults.Answers[0]

	if !bestAnswer.HasAnswer {
		return "I found some related content, but couldn't extract a specific answer to your question. Could you be more specific?", nil
	}

	// Create a natural chat response
	response := bestAnswer.Answer

	// Add source context if helpful
	if len(qaResults.Answers) > 1 {
		response += "\n\n"
		additionalSources := 0
		for i := 1; i < len(qaResults.Answers) && i < 3; i++ {
			if qaResults.Answers[i].HasAnswer && qaResults.Answers[i].Confidence > 0.6 {
				additionalSources++
			}
		}
		if additionalSources > 0 {
			response += fmt.Sprintf("I found this information across %d sources.", additionalSources+1)
		}
	}

	return response, &bestAnswer.Confidence
}

// prepareSources formats sources for the chat response
func (cs *ChatService) prepareSources(answers []*AnswerResult) []AnswerResult {
	sources := make([]AnswerResult, 0, len(answers))
	for _, answer := range answers {
		if answer.HasAnswer {
			sources = append(sources, *answer)
		}
	}
	return sources
}

// getConversationDocuments retrieves documents associated with a conversation
func (cs *ChatService) getConversationDocuments(conversationID uuid.UUID) ([]ChatDocumentReference, error) {
	var docs []ChatDocumentReference

	// This would need to join with the content_items table
	// For now, return empty - can be implemented when document management is added

	return docs, nil
}

// GetConversations retrieves user's chat conversations
func (cs *ChatService) GetConversations(userID uuid.UUID, limit int) ([]models.ChatConversation, error) {
	var conversations []models.ChatConversation
	err := cs.db.Where("user_id = ?", userID).
		Order("last_activity DESC").
		Limit(limit).
		Find(&conversations).Error

	return conversations, err
}

// GetConversationMessages retrieves messages for a conversation
func (cs *ChatService) GetConversationMessages(userID uuid.UUID, conversationID uuid.UUID) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage

	// First verify user owns the conversation
	var conversation models.ChatConversation
	err := cs.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error
	if err != nil {
		return nil, err
	}

	// Get messages
	err = cs.db.Where("conversation_id = ?", conversationID).
		Order("created_at ASC").
		Find(&messages).Error

	return messages, err
}

// AddDocumentToConversation links a document to a conversation
func (cs *ChatService) AddDocumentToConversation(userID, conversationID, contentItemID uuid.UUID) error {
	// First verify user owns both the conversation and the document
	var conversation models.ChatConversation
	err := cs.db.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error
	if err != nil {
		return err
	}

	// Create the link
	conversationDoc := models.ConversationDocument{
		ID:             uuid.New(),
		ConversationID: conversationID,
		ContentItemID:  contentItemID,
		AddedAt:        time.Now(),
	}

	return cs.db.Create(&conversationDoc).Error
}