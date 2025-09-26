package services

import (
	"regexp"
	"strings"
)

type ChunkService struct {
	maxTokens int
	tokenizer *TokenizerService
}

func NewChunkService() *ChunkService {
	tokenizer, err := NewTokenizerService()
	if err != nil {
		// Fallback to nil tokenizer (will use approximation)
		tokenizer = nil
	}

	return &ChunkService{
		maxTokens: 500, // tokens per chunk
		tokenizer: tokenizer,
	}
}

func (c *ChunkService) ChunkText(text string, maxTokens int) []string {
	if maxTokens == 0 {
		maxTokens = c.maxTokens
	}

	// Simple sentence-based chunking for now
	// TODO: Implement smarter chunking with overlap
	sentences := c.splitIntoSentences(text)

	var chunks []string
	var currentChunk strings.Builder
	currentTokens := 0

	for _, sentence := range sentences {
		sentenceTokens := c.CountTokens(sentence)

		// If adding this sentence would exceed limit, start new chunk
		if currentTokens+sentenceTokens > maxTokens && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
			currentTokens = 0
		}

		currentChunk.WriteString(sentence)
		currentChunk.WriteString(" ")
		currentTokens += sentenceTokens
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

func (c *ChunkService) splitIntoSentences(text string) []string {
	// Simple sentence splitting on periods, exclamation marks, question marks
	// TODO: Use a proper NLP library for better sentence segmentation

	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Split on sentence endings
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})

	var result []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 0 {
			result = append(result, sentence)
		}
	}

	return result
}

func (c *ChunkService) CountTokens(text string) int {
	if c.tokenizer != nil {
		return c.tokenizer.CountTokens(text)
	}
	// Fallback approximation: 1 token ≈ 0.75 words for English text
	return len(strings.Fields(text))
}

// ChunkWithOverlap creates overlapping chunks for better context preservation
func (c *ChunkService) ChunkWithOverlap(text string, chunkSize, overlap int) []string {
	if c.tokenizer != nil {
		return c.chunkWithOverlapTokenized(text, chunkSize, overlap)
	}
	return c.chunkWithOverlapWordBased(text, chunkSize, overlap)
}

// chunkWithOverlapTokenized uses proper tokenization
func (c *ChunkService) chunkWithOverlapTokenized(text string, chunkSize, overlap int) []string {
	tokens := c.tokenizer.encoder.Encode(text, nil, nil)

	if len(tokens) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(tokens) {
		end := start + chunkSize
		if end > len(tokens) {
			end = len(tokens)
		}

		chunkTokens := tokens[start:end]
		chunkText := c.tokenizer.encoder.Decode(chunkTokens)
		chunks = append(chunks, strings.TrimSpace(chunkText))

		// Move start position with overlap
		if end == len(tokens) {
			break
		}
		start = end - overlap
		if start <= 0 {
			start = 1
		}
	}

	return chunks
}

// chunkWithOverlapWordBased falls back to word-based chunking
func (c *ChunkService) chunkWithOverlapWordBased(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)

	if len(words) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(words) {
		end := start + chunkSize
		if end > len(words) {
			end = len(words)
		}

		chunkWords := words[start:end]
		chunkText := strings.Join(chunkWords, " ")
		chunks = append(chunks, chunkText)

		// Move start position with overlap
		if end == len(words) {
			break
		}
		start = end - overlap
		if start <= 0 {
			start = 1
		}
	}

	return chunks
}

// SmartChunkBySentences - Improved sentence-aware chunking
func (c *ChunkService) SmartChunkBySentences(text string, maxTokens int) []string {
	sentences := c.smartSentenceSplit(text)

	var chunks []string
	var currentChunk strings.Builder
	currentTokens := 0

	for _, sentence := range sentences {
		sentenceTokens := c.CountTokens(sentence)

		// If adding this sentence would exceed limit, start new chunk
		if currentTokens+sentenceTokens > maxTokens && currentChunk.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
			currentChunk.Reset()
			currentTokens = 0
		}

		// If single sentence is too long, split it
		if sentenceTokens > maxTokens {
			// Split long sentence into smaller parts
			subChunks := c.ChunkWithOverlap(sentence, maxTokens, 50) // 50 token overlap
			for i, subChunk := range subChunks {
				if i == 0 && currentChunk.Len() > 0 {
					// Add first part to current chunk if there's space
					currentChunk.WriteString(" ")
					currentChunk.WriteString(subChunk)
					chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
					currentChunk.Reset()
					currentTokens = 0
				} else {
					chunks = append(chunks, subChunk)
				}
			}
		} else {
			currentChunk.WriteString(sentence)
			currentChunk.WriteString(" ")
			currentTokens += sentenceTokens
		}
	}

	// Add final chunk
	if currentChunk.Len() > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk.String()))
	}

	return chunks
}

// smartSentenceSplit - Better sentence boundary detection
func (c *ChunkService) smartSentenceSplit(text string) []string {
	// Clean up text
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Enhanced sentence splitting with regex for better accuracy
	// Handles abbreviations, decimals, etc.
	sentenceRegex := regexp.MustCompile(`([.!?]+)(\s+|$)`)

	// Split on sentence boundaries but keep the punctuation
	parts := sentenceRegex.Split(text, -1)
	matches := sentenceRegex.FindAllString(text, -1)

	var sentences []string
	for i, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}

		sentence := strings.TrimSpace(part)
		if i < len(matches) {
			sentence += strings.TrimSpace(matches[i])
		}

		// Filter out very short sentences (likely abbreviations or fragments)
		if len(sentence) > 10 {
			sentences = append(sentences, sentence)
		}
	}

	return sentences
}