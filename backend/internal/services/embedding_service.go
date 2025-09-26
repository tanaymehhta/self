package services

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type EmbeddingService struct {
	client    *openai.Client
	model     string
	dimension int
}

func NewEmbeddingService() *EmbeddingService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// For development, we'll create a mock service
		return &EmbeddingService{
			client:    nil,
			model:     "text-embedding-ada-002",
			dimension: 1536,
		}
	}

	return &EmbeddingService{
		client:    openai.NewClient(apiKey),
		model:     "text-embedding-ada-002", // Reliable model
		dimension: 1536,
	}
}

func (e *EmbeddingService) CreateEmbedding(text string) (*Embedding, error) {
	if e.client == nil {
		// Mock embedding for development
		return e.createMockEmbedding(text), nil
	}

	resp, err := e.client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.AdaEmbeddingV2,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	return &Embedding{
		ID:               uuid.New(),
		EmbeddingModel:   e.model,
		EmbeddingDim:     e.dimension,
		Vector:           resp.Data[0].Embedding,
		EmbeddingVersion: 1,
	}, nil
}

func (e *EmbeddingService) createMockEmbedding(text string) *Embedding {
	// Create a simple mock embedding for development
	// In real usage, this would be replaced with actual OpenAI embeddings
	vector := make([]float32, e.dimension)

	// Simple hash-based mock (don't use in production)
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}

	for i := range vector {
		vector[i] = float32((hash + i) % 1000) / 1000.0
	}

	return &Embedding{
		ID:               uuid.New(),
		EmbeddingModel:   "mock-embedding-dev",
		EmbeddingDim:     e.dimension,
		Vector:           vector,
		EmbeddingVersion: 1,
	}
}

func (e *EmbeddingService) GetModel() string {
	return e.model
}

func (e *EmbeddingService) GetDimension() int {
	return e.dimension
}