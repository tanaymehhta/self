package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/tanaymehhta/self/backend/internal/models"
)

type TextPipeline struct {
	db               *gorm.DB
	embeddingService *EmbeddingService
	chunkService     *ChunkService
	textExtractor    *TextExtractorService
}

type ContentItem struct {
	ID           uuid.UUID    `json:"id"`
	UserID       uuid.UUID    `json:"user_id"`
	ContentType  string       `json:"content_type"`
	Title        string       `json:"title"`
	FilePath     string       `json:"file_path"`
	FileSize     int64        `json:"file_size"`
	Checksum     string       `json:"checksum"`
	Language     string       `json:"language"`
	SourceMeta   models.JSONB `json:"source_metadata"`
}

type Chunk struct {
	ID             uuid.UUID    `json:"id"`
	ContentItemID  uuid.UUID    `json:"content_item_id"`
	ChunkText      string       `json:"chunk_text"`
	ChunkIndex     int          `json:"chunk_index"`
	TokenCount     int          `json:"token_count"`
	ChunkSpan      models.JSONB `json:"chunk_span"`
}

type Embedding struct {
	ID              uuid.UUID `json:"id"`
	ChunkID         uuid.UUID `json:"chunk_id"`
	EmbeddingModel  string    `json:"embedding_model"`
	EmbeddingDim    int       `json:"embedding_dim"`
	Vector          []float32 `json:"embedding"`
	EmbeddingVersion int      `json:"embedding_version"`
}

func NewTextPipeline(db *gorm.DB) *TextPipeline {
	return &TextPipeline{
		db:               db,
		embeddingService: NewEmbeddingService(),
		chunkService:     NewChunkService(),
		textExtractor:    NewTextExtractorService(),
	}
}

func (t *TextPipeline) ProcessDocument(userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*ContentItem, error) {
	return t.ProcessDocumentWithLogging(userID, file, header, nil)
}

func (t *TextPipeline) ProcessDocumentWithLogging(userID uuid.UUID, file multipart.File, header *multipart.FileHeader, logger *PipelineLogger) (*ContentItem, error) {
	if logger == nil {
		logger = NewPipelineLogger()
	}

	logger.LogStart("file_validation", fmt.Sprintf("Processing file: %s (%.2f KB)", header.Filename, float64(header.Size)/1024))

	// 1. Save file to storage
	filePath := fmt.Sprintf("uploads/documents/%s", header.Filename)

	// 2. Extract text content
	logger.LogStart("file_reading", "Reading file content into memory")
	fileContent, err := io.ReadAll(file)
	if err != nil {
		logger.LogError("file_reading", "Failed to read file", err)
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	logger.LogSuccess("file_reading", fmt.Sprintf("Successfully read %d bytes", len(fileContent)), nil)

	// 3. Text extraction
	logger.LogStart("text_extraction", fmt.Sprintf("Extracting text from %s file", filepath.Ext(header.Filename)))
	text, err := t.textExtractor.ExtractText(fileContent, header.Filename)
	if err != nil {
		logger.LogError("text_extraction", "Failed to extract text", err)
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	textStats := map[string]interface{}{
		"character_count": len(text),
		"word_count":      len(strings.Fields(text)),
		"preview":         func() string {
			if len(text) > 200 {
				return text[:200] + "..."
			}
			return text
		}(),
	}
	logger.LogSuccess("text_extraction", fmt.Sprintf("Extracted text: %d characters, %d words",
		textStats["character_count"], textStats["word_count"]), textStats)

	// 4. Create content item
	logger.LogStart("content_item_creation", "Creating content item record")
	contentItem := &ContentItem{
		ID:          uuid.New(),
		UserID:      userID,
		ContentType: "document",
		Title:       strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)),
		FilePath:    filePath,
		FileSize:    header.Size,
		Language:    "en", // TODO: Detect language
		SourceMeta:  models.JSONB{
			"filename":  header.Filename,
			"mime_type": header.Header.Get("Content-Type"),
		},
	}

	// 5. Save to database
	err = t.db.Table("content_items").Create(contentItem).Error
	if err != nil {
		logger.LogError("content_item_creation", "Failed to save content item", err)
		return nil, fmt.Errorf("failed to save content item: %w", err)
	}
	logger.LogSuccess("content_item_creation", fmt.Sprintf("Created content item with ID: %s", contentItem.ID),
		map[string]interface{}{"content_id": contentItem.ID, "title": contentItem.Title})

	// 6. Process text into chunks and embeddings with detailed logging
	logger.LogStart("async_processing", "Starting chunking and embedding process")
	go t.processTextAsyncWithLogging(contentItem.ID, text, logger)

	return contentItem, nil
}

func (t *TextPipeline) processTextAsync(contentItemID uuid.UUID, text string) {
	logger := NewPipelineLogger()
	t.processTextAsyncWithLogging(contentItemID, text, logger)
}

func (t *TextPipeline) processTextAsyncWithLogging(contentItemID uuid.UUID, text string, logger *PipelineLogger) {
	defer func() {
		logger.Complete()
		logger.Print() // Print to console for debugging
	}()

	// 1. Smart chunk the text with overlap for better context
	logger.LogStart("text_chunking", fmt.Sprintf("Chunking text into smart sentence-based chunks (400 token limit)"))
	chunks := t.chunkService.SmartChunkBySentences(text, 400) // 400 tokens per chunk for better LLM processing

	chunkStats := make([]map[string]interface{}, 0)
	for i, chunk := range chunks {
		tokenCount := t.chunkService.CountTokens(chunk)
		chunkStats = append(chunkStats, map[string]interface{}{
			"index":       i,
			"token_count": tokenCount,
			"char_count":  len(chunk),
			"preview":     func() string {
				if len(chunk) > 100 {
					return chunk[:100] + "..."
				}
				return chunk
			}(),
		})
	}

	logger.LogSuccess("text_chunking", fmt.Sprintf("Created %d chunks", len(chunks)),
		map[string]interface{}{
			"chunk_count": len(chunks),
			"chunks":      chunkStats,
		})

	// 2. Process each chunk
	successfulChunks := 0
	successfulEmbeddings := 0

	for i, chunkText := range chunks {
		// Save chunk to database
		logger.LogStart(fmt.Sprintf("chunk_%d_save", i), fmt.Sprintf("Saving chunk %d to database", i))

		chunk := &Chunk{
			ID:            uuid.New(),
			ContentItemID: contentItemID,
			ChunkText:     chunkText,
			ChunkIndex:    i,
			TokenCount:    t.chunkService.CountTokens(chunkText),
			ChunkSpan:     models.JSONB{
				"chunk_index":     i,
				"start_sentence":  i * 3,              // Approximate sentence tracking
				"method":         "smart_sentences",   // Track chunking method
				"token_count":    t.chunkService.CountTokens(chunkText),
			},
		}

		err := t.db.Create(chunk).Error
		if err != nil {
			logger.LogError(fmt.Sprintf("chunk_%d_save", i), "Failed to save chunk", err)
			continue
		}

		logger.LogSuccess(fmt.Sprintf("chunk_%d_save", i), fmt.Sprintf("Saved chunk with ID: %s", chunk.ID),
			map[string]interface{}{
				"chunk_id":    chunk.ID,
				"token_count": chunk.TokenCount,
			})
		successfulChunks++

		// Create embedding
		logger.LogStart(fmt.Sprintf("chunk_%d_embedding", i), fmt.Sprintf("Creating embedding for chunk %d", i))

		embedding, err := t.embeddingService.CreateEmbedding(chunkText)
		if err != nil {
			logger.LogError(fmt.Sprintf("chunk_%d_embedding", i), "Failed to create embedding", err)
			continue
		}

		embedding.ChunkID = chunk.ID
		err = t.db.Create(embedding).Error
		if err != nil {
			logger.LogError(fmt.Sprintf("chunk_%d_embedding", i), "Failed to save embedding", err)
			continue
		}

		logger.LogSuccess(fmt.Sprintf("chunk_%d_embedding", i), fmt.Sprintf("Created embedding with ID: %s", embedding.ID),
			map[string]interface{}{
				"embedding_id":    embedding.ID,
				"embedding_model": embedding.EmbeddingModel,
				"vector_dim":      embedding.EmbeddingDim,
			})
		successfulEmbeddings++
	}

	// Final summary
	logger.LogSuccess("processing_complete", fmt.Sprintf("Processing complete: %d/%d chunks saved, %d/%d embeddings created",
		successfulChunks, len(chunks), successfulEmbeddings, len(chunks)),
		map[string]interface{}{
			"total_chunks":          len(chunks),
			"successful_chunks":     successfulChunks,
			"successful_embeddings": successfulEmbeddings,
			"content_id":           contentItemID,
		})
}

// Legacy method - now handled by TextExtractorService
// Keeping for backward compatibility but redirecting to new service
func (t *TextPipeline) extractTextFromFile(content []byte, filename string) string {
	text, err := t.textExtractor.ExtractText(content, filename)
	if err != nil {
		fmt.Printf("Text extraction failed for %s: %v\n", filename, err)
		// Fallback to treating as plain text
		return string(content)
	}
	return text
}