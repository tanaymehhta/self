package services

import (
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

type TokenizerService struct {
	encoder *tiktoken.Tiktoken
}

func NewTokenizerService() (*TokenizerService, error) {
	// Use GPT-3.5-turbo encoder (cl100k_base) - compatible with most modern models
	encoder, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoder: %w", err)
	}

	return &TokenizerService{
		encoder: encoder,
	}, nil
}

func (t *TokenizerService) CountTokens(text string) int {
	if t.encoder == nil {
		// Fallback to rough approximation
		return len(strings.Fields(text))
	}

	tokens := t.encoder.Encode(text, nil, nil)
	return len(tokens)
}

func (t *TokenizerService) Tokenize(text string) []string {
	if t.encoder == nil {
		// Fallback to word splitting
		return strings.Fields(text)
	}

	tokens := t.encoder.Encode(text, nil, nil)
	result := make([]string, len(tokens))

	for i, token := range tokens {
		// Decode individual tokens back to strings
		decoded := t.encoder.Decode([]int{token})
		result[i] = decoded
	}

	return result
}

func (t *TokenizerService) TruncateToTokenLimit(text string, maxTokens int) string {
	if t.encoder == nil {
		// Fallback: truncate by words
		words := strings.Fields(text)
		if len(words) <= maxTokens {
			return text
		}
		return strings.Join(words[:maxTokens], " ")
	}

	tokens := t.encoder.Encode(text, nil, nil)
	if len(tokens) <= maxTokens {
		return text
	}

	// Truncate tokens and decode back
	truncatedTokens := tokens[:maxTokens]
	return t.encoder.Decode(truncatedTokens)
}

func (t *TokenizerService) SplitIntoTokenChunks(text string, chunkSize int) []string {
	if t.encoder == nil {
		// Fallback to word-based chunking
		words := strings.Fields(text)
		var chunks []string

		for i := 0; i < len(words); i += chunkSize {
			end := i + chunkSize
			if end > len(words) {
				end = len(words)
			}
			chunks = append(chunks, strings.Join(words[i:end], " "))
		}
		return chunks
	}

	tokens := t.encoder.Encode(text, nil, nil)
	var chunks []string

	for i := 0; i < len(tokens); i += chunkSize {
		end := i + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}

		chunkTokens := tokens[i:end]
		chunkText := t.encoder.Decode(chunkTokens)
		chunks = append(chunks, chunkText)
	}

	return chunks
}