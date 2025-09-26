package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type SearchService struct {
	db                     *gorm.DB
	embeddingService       *EmbeddingService
	answerExtractionService *AnswerExtractionService
}

type SearchResult struct {
	ID           string                 `json:"id"`
	ChunkText    string                 `json:"text"`
	ContentTitle string                 `json:"content_title"`
	ContentType  string                 `json:"content_type"`
	ChunkSpan    map[string]interface{} `json:"chunk_span"`
	Relevance    float64               `json:"relevance"`
	Source       string                `json:"source"` // "vector" or "fulltext"
}

type SearchResults struct {
	Results  []SearchResult `json:"results"`
	Strategy string         `json:"strategy"`
	Total    int           `json:"total"`
}

// QASearchResults represents answer-based search results
type QASearchResults struct {
	Answers  []*AnswerResult `json:"answers"`
	Strategy string          `json:"strategy"`
	Total    int            `json:"total"`
}

func NewSearchService(db *gorm.DB, answerExtractionService *AnswerExtractionService) *SearchService {
	return &SearchService{
		db:                     db,
		embeddingService:       NewEmbeddingService(),
		answerExtractionService: answerExtractionService,
	}
}

func (s *SearchService) Search(query string, limit int) (*SearchResults, error) {
	// 1. Vector similarity search
	vectorResults, err := s.vectorSearch(query, limit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 2. Full-text search
	textResults, err := s.fullTextSearch(query, limit)
	if err != nil {
		return nil, fmt.Errorf("fulltext search failed: %w", err)
	}

	// 3. Combine and deduplicate
	combined := s.combineResults(vectorResults, textResults, limit)

	return &SearchResults{
		Results:  combined,
		Strategy: "hybrid",
		Total:    len(combined),
	}, nil
}

func (s *SearchService) vectorSearch(query string, limit int) ([]SearchResult, error) {
	// Create embedding for query
	embedding, err := s.embeddingService.CreateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to create query embedding: %w", err)
	}

	var results []SearchResult
	var rows *sql.Rows

	// PostgreSQL vector similarity search
	// Convert float32 slice to string format for pgvector
	vectorStr := fmt.Sprintf("[%s]", strings.Join(func() []string {
		strs := make([]string, len(embedding.Vector))
		for i, v := range embedding.Vector {
			strs[i] = fmt.Sprintf("%f", v)
		}
		return strs
	}(), ","))

	sqlQuery := fmt.Sprintf(`
		SELECT c.chunk_text, c.chunk_span, ci.title, ci.content_type, c.id,
		       e.embedding <=> '%s'::vector AS distance
		FROM embeddings e
		JOIN chunks c ON e.chunk_id = c.id
		JOIN content_items ci ON c.content_item_id = ci.id
		WHERE e.embedding_model = ?
		ORDER BY e.embedding <=> '%s'::vector
		LIMIT ?
	`, vectorStr, vectorStr)

	rows, err = s.db.Raw(sqlQuery, embedding.EmbeddingModel, limit).Rows()

	if err != nil {
		return nil, fmt.Errorf("vector search query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result SearchResult
		var distance float64
		var chunkSpanJSON []byte

		err := rows.Scan(&result.ChunkText, &chunkSpanJSON, &result.ContentTitle,
						&result.ContentType, &result.ID, &distance)
		if err != nil {
			continue
		}

		result.Relevance = 1.0 - distance // Convert distance to relevance
		result.Source = "vector"

		// Parse chunk span if available
		if len(chunkSpanJSON) > 0 {
			var spanData map[string]interface{}
			if err := json.Unmarshal(chunkSpanJSON, &spanData); err == nil {
				result.ChunkSpan = spanData
			} else {
				result.ChunkSpan = map[string]interface{}{"parse_error": "invalid JSON"}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SearchService) fullTextSearch(query string, limit int) ([]SearchResult, error) {
	var results []SearchResult

	// PostgreSQL full-text search
	rows, err := s.db.Raw(`
		SELECT c.chunk_text, c.chunk_span, ci.title, ci.content_type, c.id,
		       ts_rank(to_tsvector('english', c.chunk_text), plainto_tsquery('english', ?)) as rank
		FROM chunks c
		JOIN content_items ci ON c.content_item_id = ci.id
		WHERE to_tsvector('english', c.chunk_text) @@ plainto_tsquery('english', ?)
		ORDER BY rank DESC
		LIMIT ?
	`, query, query, limit).Rows()

	if err != nil {
		return nil, fmt.Errorf("fulltext search query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result SearchResult
		var rank float64
		var chunkSpanJSON []byte

		err := rows.Scan(&result.ChunkText, &chunkSpanJSON, &result.ContentTitle,
						&result.ContentType, &result.ID, &rank)
		if err != nil {
			continue
		}

		result.Relevance = rank
		result.Source = "fulltext"

		// Parse chunk span if available
		if len(chunkSpanJSON) > 0 {
			var spanData map[string]interface{}
			if err := json.Unmarshal(chunkSpanJSON, &spanData); err == nil {
				result.ChunkSpan = spanData
			} else {
				result.ChunkSpan = map[string]interface{}{"parse_error": "invalid JSON"}
			}
		}

		results = append(results, result)
	}

	return results, nil
}

func (s *SearchService) combineResults(vectorResults, textResults []SearchResult, limit int) []SearchResult {
	// Advanced multi-modal relevance fusion with content-type weighting

	seen := make(map[string]bool)
	var allResults []SearchResult

	// Collect all results with advanced scoring
	for _, result := range vectorResults {
		if !seen[result.ID] {
			result.Relevance = s.calculateAdvancedRelevance(result, "vector")
			allResults = append(allResults, result)
			seen[result.ID] = true
		}
	}

	for _, result := range textResults {
		if !seen[result.ID] {
			result.Relevance = s.calculateAdvancedRelevance(result, "fulltext")
			allResults = append(allResults, result)
			seen[result.ID] = true
		} else {
			// If already exists from vector search, boost its score
			for i, existing := range allResults {
				if existing.ID == result.ID {
					allResults[i].Relevance = s.boostDualSourceScore(existing.Relevance, result.Relevance)
					break
				}
			}
		}
	}

	// Sort by advanced relevance score
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Relevance > allResults[j].Relevance
	})

	// Return top results
	if len(allResults) > limit {
		return allResults[:limit]
	}
	return allResults
}

func (s *SearchService) calculateAdvancedRelevance(result SearchResult, searchType string) float64 {
	baseScore := result.Relevance

	// Content type weighting factors
	contentTypeWeight := s.getContentTypeWeight(result.ContentType)

	// Information density scoring
	densityScore := s.calculateInformationDensity(result.ChunkText)

	// Context window relevance (how much of chunk is relevant)
	contextScore := s.calculateContextRelevance(result.ChunkText)

	// Source authority scoring
	authorityScore := s.calculateSourceAuthority(result.ContentType)

	// Temporal relevance (newer content slightly preferred)
	temporalScore := s.calculateTemporalRelevance(result)

	// Multi-stage scoring formula
	advancedScore := baseScore *
		contentTypeWeight *
		densityScore *
		contextScore *
		authorityScore *
		temporalScore

	return advancedScore
}

func (s *SearchService) getContentTypeWeight(contentType string) float64 {
	// Sophisticated content type weighting
	weights := map[string]float64{
		"document":    1.0,   // Full weight for documents
		"audio":       0.7,   // Reduced weight for audio (less dense)
		"video":       0.6,   // Reduced weight for video transcripts
		"image":       0.5,   // Reduced weight for image descriptions
		"webpage":     0.8,   // Medium weight for web content
		"email":       0.9,   // High weight for emails (usually focused)
	}

	if weight, exists := weights[contentType]; exists {
		return weight
	}
	return 0.8 // Default weight
}

func (s *SearchService) calculateInformationDensity(chunkText string) float64 {
	// Calculate how information-dense the chunk is
	textLength := len(chunkText)

	// Simple heuristic: longer chunks with substantive content score higher
	if textLength < 100 {
		return 0.5 // Short chunks (like brief audio mentions) get penalized
	} else if textLength < 300 {
		return 0.7
	} else if textLength < 500 {
		return 0.9
	} else {
		return 1.0 // Full chunks get full score
	}
}

func (s *SearchService) calculateContextRelevance(chunkText string) float64 {
	// Calculate what percentage of the chunk contains substantial information
	// vs noise/filler words
	words := strings.Fields(chunkText)
	if len(words) == 0 {
		return 0.5
	}

	// Count meaningful words (not stop words)
	meaningfulWords := 0
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "will": true,
		"would": true, "could": true, "should": true, "this": true, "that": true,
		"these": true, "those": true, "it": true, "its": true, "i": true, "you": true,
		"he": true, "she": true, "we": true, "they": true, "them": true, "their": true,
	}

	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ".,!?;:()[]{}\"'"))
		if !stopWords[word] && len(word) > 2 {
			meaningfulWords++
		}
	}

	ratio := float64(meaningfulWords) / float64(len(words))
	// Scale to 0.7-1.0 range (even low-density text has some value)
	return 0.7 + (ratio * 0.3)
}

func (s *SearchService) calculateSourceAuthority(contentType string) float64 {
	// Weight sources by their typical authority/reliability
	authorityWeights := map[string]float64{
		"document":    1.0,   // Documents typically authoritative
		"webpage":     0.7,   // Web content varies in quality
		"audio":       0.8,   // Meeting notes, lectures valuable
		"email":       0.9,   // Emails usually focused/intentional
	}

	if weight, exists := authorityWeights[contentType]; exists {
		return weight
	}
	return 0.8
}

func (s *SearchService) calculateTemporalRelevance(result SearchResult) float64 {
	// Since we don't have creation timestamps in SearchResult yet,
	// we'll use content type to infer temporal preferences
	// Documents tend to be more evergreen, conversations more time-sensitive

	temporalWeights := map[string]float64{
		"document": 1.0,    // Documents are timeless
		"email":    0.95,   // Emails lose relevance slowly
		"webpage":  0.9,    // Web content can become outdated
		"audio":    0.85,   // Conversations become less relevant over time
		"video":    0.85,   // Video content ages
	}

	if weight, exists := temporalWeights[result.ContentType]; exists {
		return weight
	}
	return 0.95 // Default slight preference for newer content
}

func (s *SearchService) boostDualSourceScore(vectorScore, fulltextScore float64) float64 {
	// If content appears in both vector and fulltext results,
	// it's highly relevant - boost its score
	return vectorScore * 1.2 // 20% boost for dual-source matches
}

// Simple search (fallback when embeddings aren't available)
func (s *SearchService) SimpleSearch(query string, limit int) (*SearchResults, error) {
	var results []SearchResult

	query = strings.ToLower(query)

	rows, err := s.db.Raw(`
		SELECT c.chunk_text, c.chunk_span, ci.title, ci.content_type, c.id
		FROM chunks c
		JOIN content_items ci ON c.content_item_id = ci.id
		WHERE LOWER(c.chunk_text) LIKE ?
		ORDER BY c.created_at DESC
		LIMIT ?
	`, "%"+query+"%", limit).Rows()

	if err != nil {
		return nil, fmt.Errorf("simple search query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result SearchResult
		var chunkSpanJSON []byte

		err := rows.Scan(&result.ChunkText, &chunkSpanJSON, &result.ContentTitle,
						&result.ContentType, &result.ID)
		if err != nil {
			continue
		}

		result.Relevance = 0.5 // Default relevance for simple search
		result.Source = "simple"

		results = append(results, result)
	}

	return &SearchResults{
		Results:  results,
		Strategy: "simple",
		Total:    len(results),
	}, nil
}

// QASearch performs two-stage search: retrieval -> answer extraction
func (s *SearchService) QASearch(ctx context.Context, query string, limit int) (*QASearchResults, error) {
	// Stage 1: Retrieve candidate chunks (more than final limit)
	candidateLimit := limit * 3 // Get 3x candidates for better answer extraction

	// 1. Vector similarity search
	vectorResults, err := s.vectorSearch(query, candidateLimit)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 2. Full-text search
	textResults, err := s.fullTextSearch(query, candidateLimit)
	if err != nil {
		return nil, fmt.Errorf("fulltext search failed: %w", err)
	}

	// 3. Combine and get candidate chunks
	candidateChunks := s.prepareCandidateChunks(vectorResults, textResults, candidateLimit)

	// Stage 2: Extract answers from candidate chunks
	answers, err := s.answerExtractionService.ExtractAnswersFromChunks(ctx, query, candidateChunks)
	if err != nil {
		return nil, fmt.Errorf("answer extraction failed: %w", err)
	}

	// Rank answers by confidence
	rankedAnswers := RankAnswersByConfidence(answers)

	// Limit final results
	if len(rankedAnswers) > limit {
		rankedAnswers = rankedAnswers[:limit]
	}

	return &QASearchResults{
		Answers:  rankedAnswers,
		Strategy: "qa-hybrid",
		Total:    len(rankedAnswers),
	}, nil
}

// prepareCandidateChunks converts search results into chunks with metadata
func (s *SearchService) prepareCandidateChunks(vectorResults, textResults []SearchResult, limit int) []ChunkWithMetadata {
	seen := make(map[string]bool)
	var chunks []ChunkWithMetadata

	// Process vector results
	for _, result := range vectorResults {
		if !seen[result.ID] && len(chunks) < limit {
			chunks = append(chunks, ChunkWithMetadata{
				Text: result.ChunkText,
				Metadata: SourceMetadata{
					ChunkID:     result.ID,
					Title:       result.ContentTitle,
					ContentType: result.ContentType,
				},
			})
			seen[result.ID] = true
		}
	}

	// Process text results
	for _, result := range textResults {
		if !seen[result.ID] && len(chunks) < limit {
			chunks = append(chunks, ChunkWithMetadata{
				Text: result.ChunkText,
				Metadata: SourceMetadata{
					ChunkID:     result.ID,
					Title:       result.ContentTitle,
					ContentType: result.ContentType,
				},
			})
			seen[result.ID] = true
		}
	}

	return chunks
}