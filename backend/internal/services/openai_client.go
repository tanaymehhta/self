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

// OpenAIClient implements LLMClient interface using OpenAI's API
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a response choice
type Choice struct {
	Message Message `json:"message"`
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini" // Default to cost-effective model
	}

	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1/chat/completions",
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExtractAnswer implements LLMClient interface
func (c *OpenAIClient) ExtractAnswer(ctx context.Context, query, chunk string) (*LLMResponse, error) {
	// Create the prompt for answer extraction
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

	// Prepare OpenAI request
	request := OpenAIRequest{
		Model:       c.model,
		Temperature: 0.1, // Low temperature for consistent extraction
		MaxTokens:   500,  // Reasonable limit for answer extraction
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// Make API call
	response, err := c.callOpenAI(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Parse LLM response
	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	content := response.Choices[0].Message.Content

	// Parse JSON response from LLM
	var llmResponse LLMResponse
	if err := json.Unmarshal([]byte(content), &llmResponse); err != nil {
		// If JSON parsing fails, create a fallback response
		return &LLMResponse{
			Answer:     "",
			Confidence: 0.0,
			HasAnswer:  false,
			Reasoning:  fmt.Sprintf("Failed to parse LLM response: %v", err),
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

// callOpenAI makes the actual HTTP request to OpenAI API
func (c *OpenAIClient) callOpenAI(ctx context.Context, request OpenAIRequest) (*OpenAIResponse, error) {
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
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}