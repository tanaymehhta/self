package services

import (
	"context"
	"fmt"
)

// AnswerExtractionService handles extracting specific answers from text chunks using LLM
type AnswerExtractionService struct {
	llmClient LLMClient
}

// AnswerResult represents an extracted answer with metadata and confidence scoring
type AnswerResult struct {
	// Core answer data
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
	HasAnswer  bool    `json:"has_answer"`

	// Source attribution
	ChunkID     string `json:"chunk_id"`
	SourceChunk string `json:"source_chunk"`
	SourceTitle string `json:"source_title"`
	ContentType string `json:"content_type"`

	// Optional metadata
	StartTime *float64 `json:"start_time,omitempty"`
	EndTime   *float64 `json:"end_time,omitempty"`
	Speaker   *string  `json:"speaker,omitempty"`
	PageNum   *int     `json:"page_num,omitempty"`
}

// LLMResponse represents the structured response from the LLM
type LLMResponse struct {
	Answer     string  `json:"answer"`
	Confidence float64 `json:"confidence"`
	HasAnswer  bool    `json:"has_answer"`
	Reasoning  string  `json:"reasoning"`
}

// LLMClient interface allows us to swap between OpenAI, Ollama, etc.
type LLMClient interface {
	ExtractAnswer(ctx context.Context, query, chunk string) (*LLMResponse, error)
}

// NewAnswerExtractionService creates a new answer extraction service
func NewAnswerExtractionService(llmClient LLMClient) *AnswerExtractionService {
	return &AnswerExtractionService{
		llmClient: llmClient,
	}
}

// ExtractAnswer processes a chunk and query to extract a specific answer
func (s *AnswerExtractionService) ExtractAnswer(ctx context.Context, query, chunk string, sourceMetadata SourceMetadata) (*AnswerResult, error) {
	llmResponse, err := s.llmClient.ExtractAnswer(ctx, query, chunk)
	if err != nil {
		return nil, fmt.Errorf("LLM answer extraction failed: %w", err)
	}

	result := &AnswerResult{
		Answer:      llmResponse.Answer,
		Confidence:  llmResponse.Confidence,
		HasAnswer:   llmResponse.HasAnswer,
		ChunkID:     sourceMetadata.ChunkID,
		SourceChunk: chunk,
		SourceTitle: sourceMetadata.Title,
		ContentType: sourceMetadata.ContentType,
	}

	// Add content-specific metadata
	if sourceMetadata.PageNum != nil {
		result.PageNum = sourceMetadata.PageNum
	}
	if sourceMetadata.StartTime != nil {
		result.StartTime = sourceMetadata.StartTime
	}
	if sourceMetadata.EndTime != nil {
		result.EndTime = sourceMetadata.EndTime
	}
	if sourceMetadata.Speaker != nil {
		result.Speaker = sourceMetadata.Speaker
	}

	return result, nil
}

// ExtractAnswersFromChunks processes multiple chunks for a query
func (s *AnswerExtractionService) ExtractAnswersFromChunks(ctx context.Context, query string, chunks []ChunkWithMetadata) ([]*AnswerResult, error) {
	var results []*AnswerResult

	for _, chunk := range chunks {
		result, err := s.ExtractAnswer(ctx, query, chunk.Text, chunk.Metadata)
		if err != nil {
			fmt.Printf("Error extracting answer from chunk %s: %v\n", chunk.Metadata.ChunkID, err)
			continue
		}

		if result.HasAnswer && result.Confidence > 0.1 {
			results = append(results, result)
		}
	}

	return results, nil
}

// SourceMetadata contains attribution info for the chunk
type SourceMetadata struct {
	ChunkID     string
	Title       string
	ContentType string
	PageNum     *int
	StartTime   *float64
	EndTime     *float64
	Speaker     *string
}

// ChunkWithMetadata pairs chunk text with its source metadata
type ChunkWithMetadata struct {
	Text     string
	Metadata SourceMetadata
}

// RankAnswersByConfidence sorts answers by confidence score (highest first)
func RankAnswersByConfidence(answers []*AnswerResult) []*AnswerResult {
	ranked := make([]*AnswerResult, len(answers))
	copy(ranked, answers)

	for i := 0; i < len(ranked)-1; i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[i].Confidence < ranked[j].Confidence {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	return ranked
}