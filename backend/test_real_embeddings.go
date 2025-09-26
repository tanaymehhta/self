package main

import (
	"fmt"
	"os"

	"github.com/tanaymehhta/self/backend/internal/services"
)

func main() {
	fmt.Println("🎯 Testing Real OpenAI Embeddings\n")

	// Load environment variables
	os.Setenv("OPENAI_API_KEY", "your-openai-api-key-here")

	embeddingService := services.NewEmbeddingService()

	testText := "The artificial intelligence system can process natural language and generate semantic embeddings for text analysis."

	fmt.Println("📄 Test Text:")
	fmt.Printf("   %s\n\n", testText)

	fmt.Println("🔄 Creating embedding with OpenAI...")
	embedding, err := embeddingService.CreateEmbedding(testText)
	if err != nil {
		fmt.Printf("❌ Failed to create embedding: %v\n", err)
		fmt.Println("   This might be due to API limits or network issues")
		return
	}

	fmt.Printf("✅ Real OpenAI Embedding Created!\n")
	fmt.Printf("   • ID: %s\n", embedding.ID)
	fmt.Printf("   • Model: %s\n", embedding.EmbeddingModel)
	fmt.Printf("   • Dimensions: %d\n", embedding.EmbeddingDim)
	fmt.Printf("   • Vector Sample: [%.6f, %.6f, %.6f, %.6f, %.6f, ...]\n",
		embedding.Vector[0], embedding.Vector[1], embedding.Vector[2], embedding.Vector[3], embedding.Vector[4])

	// Verify it's not a mock
	if embedding.EmbeddingModel == "mock-embedding-dev" {
		fmt.Println("❌ Still using mock embeddings!")
	} else {
		fmt.Println("✅ Real OpenAI embeddings working!")
	}

	// Test cosine similarity between two similar texts
	fmt.Println("\n🔬 Testing Semantic Similarity...")

	text1 := "Machine learning algorithms can analyze data patterns."
	text2 := "AI systems are capable of data pattern analysis."
	text3 := "I like to eat pizza and pasta for dinner."

	emb1, _ := embeddingService.CreateEmbedding(text1)
	emb2, _ := embeddingService.CreateEmbedding(text2)
	emb3, _ := embeddingService.CreateEmbedding(text3)

	if emb1 != nil && emb2 != nil && emb3 != nil {
		sim12 := cosineSimilarity(emb1.Vector, emb2.Vector)
		sim13 := cosineSimilarity(emb1.Vector, emb3.Vector)

		fmt.Printf("   • Similarity (ML ↔ AI): %.4f\n", sim12)
		fmt.Printf("   • Similarity (ML ↔ Food): %.4f\n", sim13)

		if sim12 > sim13 {
			fmt.Println("✅ Semantic similarity working correctly!")
		} else {
			fmt.Println("⚠️ Unexpected similarity results")
		}
	}

	fmt.Println("\n🎉 Real embeddings test complete!")
}

func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (normA * normB)
}