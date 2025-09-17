package docprocessor

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// MockMarkitdownClient for testing
type MockMarkitdownClient struct {
	convertFunc             func(ctx context.Context, filename string, data []byte) (string, error)
	supportedExtensionsFunc func(ctx context.Context) ([]string, error)
}

func (m *MockMarkitdownClient) Convert(ctx context.Context, filename string, data []byte) (string, error) {
	if m.convertFunc != nil {
		return m.convertFunc(ctx, filename, data)
	}
	return "# Mock Markdown\nThis is converted content.", nil
}

func (m *MockMarkitdownClient) SupportedExtensions(ctx context.Context) ([]string, error) {
	if m.supportedExtensionsFunc != nil {
		return m.supportedExtensionsFunc(ctx)
	}
	return []string{".txt", ".md", ".pdf", ".docx"}, nil
}

func TestChunkText(t *testing.T) {
	processor := &Processor{}

	text := "This is a test text that should be chunked into smaller pieces for processing."
	chunks := processor.ChunkText(text, 20)

	if len(chunks) == 0 {
		t.Error("Expected chunks to be created, got none")
	}

	for i, chunk := range chunks {
		if len(chunk) > 25 { // Allow some buffer
			t.Errorf("Chunk %d is too long: %d characters", i, len(chunk))
		}
		if strings.TrimSpace(chunk) == "" {
			t.Errorf("Chunk %d is empty", i)
		}
	}
}

func TestChunkMarkdown(t *testing.T) {
	processor := &Processor{}

	markdown := `# Main Title
This is the introduction section.

## Section 1
Content for section 1 with some text that should be chunked appropriately.

### Subsection
More detailed content here.

## Section 2
Another section with different content.

` + strings.Repeat("This is repeated text to make it longer. ", 50)

	chunks := processor.ChunkMarkdown(markdown)

	if len(chunks) == 0 {
		t.Error("Expected markdown chunks to be created, got none")
	}

	// Verify all chunks have content
	for i, chunk := range chunks {
		if strings.TrimSpace(chunk) == "" {
			t.Errorf("Chunk %d is empty", i)
		}
	}
}

func TestWrapMarkdownWithMetadata(t *testing.T) {
	processor := &Processor{}

	markdown := "# Test Document\nThis is test content."
	docID := "test-doc-123"
	source := "test.md"
	fileID := uuid.New()
	createdAt := time.Now()

	wrapped := processor.WrapMarkdownWithMetadata(markdown, docID, source, fileID, createdAt)

	if len(wrapped) == 0 {
		t.Error("Expected wrapped chunks to be created, got none")
	}

	// Check that metadata is present in the first chunk
	firstChunk := wrapped[0]
	if !strings.Contains(firstChunk, "doc_id: "+docID) {
		t.Error("First chunk should contain doc_id metadata")
	}
	if !strings.Contains(firstChunk, "file_id: "+fileID.String()) {
		t.Error("First chunk should contain file_id metadata")
	}
	if !strings.Contains(firstChunk, `source: "`+source+`"`) {
		t.Error("First chunk should contain source metadata")
	}
	if !strings.Contains(firstChunk, "---") {
		t.Error("First chunk should contain YAML front matter delimiters")
	}
}

func TestWrapMarkdownWithMetadataSplitsLargeBlocks(t *testing.T) {
	processor := &Processor{}

	var builder strings.Builder
	builder.WriteString("# Massive Document\n\n")
	builder.WriteString("```\n")
	for i := 0; i < 3000; i++ {
		builder.WriteString("This is a very long line of code that should force splitting across multiple chunks due to token limits.\n")
	}
	builder.WriteString("```\n")

	docID := "massive-doc"
	source := "massive.md"
	fileID := uuid.New()
	createdAt := time.Now()

	wrapped := processor.WrapMarkdownWithMetadata(builder.String(), docID, source, fileID, createdAt)

	if len(wrapped) <= 1 {
		t.Fatalf("expected wrapped content to produce multiple chunks, got %d", len(wrapped))
	}

	for i, chunk := range wrapped {
		estimate := processor.EstimateTokenCount(chunk)
		if estimate > maxEmbeddingTokens {
			t.Fatalf("chunk %d exceeds token estimate: %d > %d", i, estimate, maxEmbeddingTokens)
		}
		if !strings.Contains(chunk, "chunk_index: ") {
			t.Fatalf("chunk %d missing chunk_index metadata", i)
		}
	}
}

func TestProcessText(t *testing.T) {
	processor := &Processor{}

	text := "This is a test text for processing."
	result, err := processor.ProcessText(text)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.Markdown != text {
		t.Error("Expected markdown to match original text")
	}

	if len(result.Chunks) == 0 {
		t.Error("Expected chunks to be created")
	}

	if result.Hash == "" {
		t.Error("Expected hash to be generated")
	}

	if result.ProcessedAt.IsZero() {
		t.Error("Expected ProcessedAt to be set")
	}
}

func TestProcessTextEmpty(t *testing.T) {
	processor := &Processor{}

	_, err := processor.ProcessText("")
	if err == nil {
		t.Error("Expected error for empty text")
	}

	_, err = processor.ProcessText("   ")
	if err == nil {
		t.Error("Expected error for whitespace-only text")
	}
}

func TestProcessTextTooLarge(t *testing.T) {
	processor := &Processor{}

	// Create text larger than 200KB
	largeText := strings.Repeat("a", 200001)
	_, err := processor.ProcessText(largeText)

	if err == nil {
		t.Error("Expected error for text exceeding maximum length")
	}
}

func TestEstimateTokenCount(t *testing.T) {
	processor := &Processor{}

	text := "This is a test text with multiple words."
	tokenCount := processor.EstimateTokenCount(text)

	// Should be roughly text length / 4
	expectedRange := len(text) / 6 // Allow some variance
	if tokenCount < expectedRange {
		t.Errorf("Token count seems too low: got %d, expected around %d", tokenCount, len(text)/4)
	}
}

func TestSplitOnSentences(t *testing.T) {
	processor := &Processor{}

	text := "This is the first sentence. This is the second sentence! And this is the third sentence?"
	sentences := processor.SplitOnSentences(text)

	if len(sentences) < 3 {
		t.Errorf("Expected at least 3 sentences, got %d", len(sentences))
	}

	for i, sentence := range sentences {
		if strings.TrimSpace(sentence) == "" {
			t.Errorf("Sentence %d is empty", i)
		}
	}
}

func TestChunkTextWithOverlap(t *testing.T) {
	processor := &Processor{}

	text := "This is a long text that will be chunked with overlap to test the functionality."
	chunks := processor.ChunkTextWithOverlap(text, 30, 10)

	if len(chunks) < 2 {
		t.Error("Expected multiple chunks with overlap")
	}

	// Check that there's some overlap between consecutive chunks
	if len(chunks) >= 2 {
		chunk1 := chunks[0]
		chunk2 := chunks[1]

		// There should be some common text between chunks due to overlap
		overlap := false
		words1 := strings.Fields(chunk1)
		words2 := strings.Fields(chunk2)

		for _, word1 := range words1[len(words1)-5:] { // Check last 5 words of first chunk
			for _, word2 := range words2[:5] { // Check first 5 words of second chunk
				if word1 == word2 {
					overlap = true
					break
				}
			}
			if overlap {
				break
			}
		}

		if !overlap {
			t.Error("Expected some overlap between consecutive chunks")
		}
	}
}
