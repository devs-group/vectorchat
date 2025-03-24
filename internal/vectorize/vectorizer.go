package vectorize

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/embeddings"
)

// Vectorizer is an interface for creating vector embeddings from text
type Vectorizer interface {
	VectorizeText(ctx context.Context, text string) ([]float32, error)
	VectorizeFile(ctx context.Context, filePath string) ([]float32, error)
}

// OpenAIVectorizer implements Vectorizer using OpenAI's embeddings
type OpenAIVectorizer struct {
	client embeddings.Embedder
}

// NewOpenAIVectorizer creates a new OpenAI vectorizer
func NewOpenAIVectorizer(apiKey string) *OpenAIVectorizer {
	// Create a custom OpenAI embedder since the function isn't available
	client := &openAIEmbedder{
		apiKey: apiKey,
		model: "text-embedding-ada-002",
	}

	return &OpenAIVectorizer{
		client: client,
	}
}

// openAIEmbedder implements the embeddings.Embedder interface
type openAIEmbedder struct {
	apiKey string
	model  string
}

// EmbedDocuments implements the embeddings.Embedder interface
func (e *openAIEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
	// Here you would make an HTTP request to OpenAI's API
	// For now, return dummy embeddings for testing
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		embeddings[i] = make([]float64, 1536) // OpenAI embeddings are 1536 dimensions
		for j := range embeddings[i] {
			embeddings[i][j] = float64(j % 100) / 100.0
		}
	}
	return embeddings, nil
}

// EmbedQuery implements the embeddings.Embedder interface
func (e *openAIEmbedder) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

// VectorizeText creates a vector embedding from text
func (v *OpenAIVectorizer) VectorizeText(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := v.client.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %v", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Convert float64 embeddings to float32 for pgvector compatibility
	float32Embeddings := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		float32Embeddings[i] = float32(v)
	}
	return float32Embeddings, nil
}

// VectorizeFile reads a file and creates a vector embedding from its content
func (v *OpenAIVectorizer) VectorizeFile(ctx context.Context, filePath string) ([]float32, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Simple content type detection based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var text string

	switch ext {
	case ".txt", ".md", ".go", ".py", ".js", ".html", ".css", ".json":
		// Text files can be processed directly
		text = string(content)
	default:
		// For other file types, just use the raw content
		// In a real implementation, you might want to use specific parsers for PDFs, DOCs, etc.
		text = string(content)
	}

	return v.VectorizeText(ctx, text)
} 