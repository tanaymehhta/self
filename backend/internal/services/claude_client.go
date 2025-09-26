package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClaudeClient implements LLMClient interface using Anthropic's Claude API
type ClaudeClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// ClaudeRequest represents the request structure for Claude API
type ClaudeRequest struct {
	Model     string         `json:"model"`
	MaxTokens int           `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
	System    string         `json:"system,omitempty"`
}

// ClaudeMessage represents a message in Claude format
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeResponse represents the response from Claude API
type ClaudeResponse struct {
	Content []ClaudeContent `json:"content"`
}

// ClaudeContent represents content in Claude response
type ClaudeContent struct {
	Text string `json:"text"`
}

// NewClaudeClient creates a new Claude client
func NewClaudeClient(apiKey, model string) *ClaudeClient {
	if model == "" {
		model = "claude-3-haiku-20240307" // Default to fast, cost-effective model
	}

	return &ClaudeClient{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExtractAnswer implements LLMClient interface
func (c *ClaudeClient) ExtractAnswer(ctx context.Context, query, chunk string) (*LLMResponse, error) {
	// Create the system prompt for answer extraction
	systemPrompt := `You are an expert at extracting specific answers from text chunks.

Your task:
1. Read the text chunk carefully
2. Determine if it contains information that answers the user's query
3. If it does, extract the most precise answer
4. If it doesn't, indicate there's no relevant answer

Respond with a JSON object containing:
- "answer": The extracted answer (or empty string if no answer)
- "confidence": Float between 0.0-1.0 indicating your confidence
- "has_answer": Boolean indicating if chunk contains relevant answer
- "reasoning": Brief explanation of your decision

Guidelines:
- Be precise and concise in answers
- Only extract information actually present in the chunk
- Don't make assumptions or add external knowledge
- Confidence should reflect how directly the chunk answers the query`

	userPrompt := fmt.Sprintf(`Query: %s

Text Chunk:
%s

Extract the answer from this chunk:`, query, chunk)

	// Prepare Claude request
	request := ClaudeRequest{
		Model:     c.model,
		MaxTokens: 500,
		System:    systemPrompt,
		Messages: []ClaudeMessage{
			{Role: "user", Content: userPrompt},
		},
	}

	// Make API call
	response, err := c.callClaude(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("Claude API call failed: %w", err)
	}

	// Parse Claude response
	if len(response.Content) == 0 {
		return nil, fmt.Errorf("no content returned from Claude")
	}

	content := response.Content[0].Text

	// Parse JSON response from LLM
	var llmResponse LLMResponse
	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		// If JSON parsing fails, create a fallback response
		return &LLMResponse{
			Answer:     "",
			Confidence: 0.0,
			HasAnswer:  false,
			Reasoning:  fmt.Sprintf("Failed to parse Claude response: %v", err),
		}, nil
	}

	// Validate confidence score
	if llmResponse.Confidence < 0.0 {
		llmResponse.Confidence = 0.0
	}
	if llmResponse.Confidence > 1.0 {
		llmResponse.Confidence = 1.0
	}

	return &llmResponse, nil
}

// callClaude makes the actual HTTP request to Claude API
func (c *ClaudeClient) callClaude(ctx context.Context, request ClaudeRequest) (*ClaudeResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Claude API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response ClaudeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}