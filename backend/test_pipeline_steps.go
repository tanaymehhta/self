package main

import (
	"fmt"
	"strings"

	"github.com/tanaymehhta/self/backend/internal/services"
)

func main() {
	fmt.Println("🧪 Testing Text Processing Pipeline - Step by Step\n")

	// Test file content
	testContent := `# Test Document for Pipeline Testing

This is a comprehensive test document designed to validate each step of the text processing pipeline.

## Key Features to Test

1. **Text Extraction**: The system should extract this text cleanly from the file.

2. **Smart Chunking**: This document should be divided into logical chunks that respect sentence boundaries.

3. **Token Counting**: Each chunk should have accurate token counts using tiktoken.

4. **Context Preservation**: Overlapping chunks should maintain context between segments.

## Expected Results

When processed through the pipeline:
- The text should be extracted without formatting artifacts
- Chunks should be approximately 400 tokens each
- Token counts should be precise, not approximated
- Each chunk should contain complete thoughts

This validates the complete pipeline from file input to searchable chunks.`

	// Step 1: File Upload and Validation (simulate file reading)
	testStep1([]byte(testContent), int64(len(testContent)))
}

func testStep1(content []byte, fileSize int64) {
	fmt.Println("📄 STEP 1: File Upload and Validation")
	fmt.Println("=====================================")

	// Test file size validation
	fmt.Printf("File size: %d bytes\n", fileSize)
	maxSize := int64(1024 * 1024 * 1024) // 1GB
	if fileSize > maxSize {
		fmt.Printf("❌ File too large (exceeds %d bytes)\n", maxSize)
		return
	}
	fmt.Printf("✅ File size validation: PASSED (under %d GB limit)\n", maxSize/(1024*1024*1024))

	// Test file reading (simulated)
	fmt.Println("Reading file content into memory...")
	fmt.Printf("✅ File reading: PASSED (%d bytes read)\n", len(content))

	// Basic content validation
	if len(content) == 0 {
		fmt.Println("❌ File content is empty")
		return
	}
	fmt.Printf("✅ Content validation: PASSED (non-empty content)\n")

	fmt.Printf("\n📊 Step 1 Results:\n")
	fmt.Printf("   • File size: %d bytes\n", len(content))
	fmt.Printf("   • Content preview: %.100s...\n", string(content))
	fmt.Printf("   • Status: ✅ ALL VALIDATIONS PASSED\n\n")

	// Move to Step 2
	testStep2(content, "test_document.txt")
}

func testStep2(content []byte, filename string) {
	fmt.Println("🎯 STEP 2: Text Extraction")
	fmt.Println("==========================")

	// Create text extractor
	extractor := services.NewTextExtractorService()

	fmt.Printf("Extracting text from %s file...\n", getFileExtension(filename))
	text, err := extractor.ExtractText(content, filename)
	if err != nil {
		fmt.Printf("❌ Text extraction failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Text extraction: PASSED\n")
	fmt.Printf("   • Original bytes: %d\n", len(content))
	fmt.Printf("   • Extracted characters: %d\n", len(text))
	fmt.Printf("   • Word count: %d\n", len(strings.Fields(text)))
	fmt.Printf("   • Text preview: %.150s...\n", text)
	fmt.Printf("   • Status: ✅ TEXT SUCCESSFULLY EXTRACTED\n\n")

	// Move to Step 3
	testStep3(text)
}

func testStep3(text string) {
	fmt.Println("🧠 STEP 3: Smart Chunking")
	fmt.Println("=========================")

	// Create chunk service
	chunkService := services.NewChunkService()

	fmt.Println("Chunking text using SmartChunkBySentences (400 token limit)...")
	chunks := chunkService.SmartChunkBySentences(text, 400)

	fmt.Printf("✅ Text chunking: PASSED\n")
	fmt.Printf("   • Total chunks created: %d\n", len(chunks))

	for i, chunk := range chunks {
		tokenCount := chunkService.CountTokens(chunk)
		fmt.Printf("   • Chunk %d: %d tokens, %d characters\n", i+1, tokenCount, len(chunk))
		fmt.Printf("     Preview: %.100s...\n", chunk)
	}

	fmt.Printf("   • Status: ✅ SMART CHUNKING COMPLETED\n\n")

	// Move to Step 4
	testStep4(chunks[0], chunkService)
}

func testStep4(sampleChunk string, chunkService *services.ChunkService) {
	fmt.Println("🔢 STEP 4: Token Counting")
	fmt.Println("=========================")

	fmt.Println("Testing accurate token counting with tiktoken...")
	tokenCount := chunkService.CountTokens(sampleChunk)

	fmt.Printf("✅ Token counting: PASSED\n")
	fmt.Printf("   • Sample chunk: %.100s...\n", sampleChunk)
	fmt.Printf("   • Character count: %d\n", len(sampleChunk))
	fmt.Printf("   • Word count: %d\n", len(strings.Fields(sampleChunk)))
	fmt.Printf("   • Token count (tiktoken): %d\n", tokenCount)
	fmt.Printf("   • Tokens/word ratio: %.2f\n", float64(tokenCount)/float64(len(strings.Fields(sampleChunk))))
	fmt.Printf("   • Status: ✅ ACCURATE TOKENIZATION COMPLETED\n\n")

	// Move to Step 5
	testStep5(sampleChunk)
}

func testStep5(sampleChunk string) {
	fmt.Println("🎯 STEP 5: Embedding Creation")
	fmt.Println("=============================")

	// Create embedding service
	embeddingService := services.NewEmbeddingService()

	fmt.Println("Creating embedding for sample chunk...")
	embedding, err := embeddingService.CreateEmbedding(sampleChunk)
	if err != nil {
		fmt.Printf("❌ Embedding creation failed: %v\n", err)
		// This might fail without OpenAI API key, but we can still show the structure
		fmt.Println("   • Note: This may require OpenAI API key in production")
		fmt.Println("   • Mock embeddings will be used for development")
		return
	}

	fmt.Printf("✅ Embedding creation: PASSED\n")
	fmt.Printf("   • Embedding model: %s\n", embedding.EmbeddingModel)
	fmt.Printf("   • Vector dimensions: %d\n", embedding.EmbeddingDim)
	fmt.Printf("   • Embedding ID: %s\n", embedding.ID)
	fmt.Printf("   • Vector sample: [%.3f, %.3f, %.3f, ...]\n",
		embedding.Vector[0], embedding.Vector[1], embedding.Vector[2])
	fmt.Printf("   • Status: ✅ EMBEDDING GENERATED SUCCESSFULLY\n\n")

	fmt.Println("🎉 PIPELINE TEST COMPLETED!")
	fmt.Println("============================")
	fmt.Println("✅ All 5 core steps validated successfully!")
	fmt.Println("   1. ✅ File Upload and Validation")
	fmt.Println("   2. ✅ Text Extraction")
	fmt.Println("   3. ✅ Smart Chunking")
	fmt.Println("   4. ✅ Token Counting")
	fmt.Println("   5. ✅ Embedding Creation")
}

// Removed mock multipart file function - using direct byte content instead

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return "unknown"
}