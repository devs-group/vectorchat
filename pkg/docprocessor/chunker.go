package docprocessor

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	maxEmbeddingTokens      = 7000
	metadataTokenBuffer     = 200
	defaultCharsPerTokenEst = 4
)

// MarkdownChunk represents a chunk of markdown with metadata
type MarkdownChunk struct {
	Section string
	Text    string
}

// ChunkText splits text into chunks of the specified size with optional overlap
func (p *Processor) ChunkText(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000
	}

	var chunks []string
	runes := []rune(text)

	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[start:end])
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

// ChunkMarkdown splits markdown content into semantic chunks
func (p *Processor) ChunkMarkdown(markdown string) []string {
	opts := DefaultChunkOptions()
	return p.ChunkMarkdownWithOptions(markdown, opts)
}

// ChunkMarkdownWithOptions splits markdown content with custom options
func (p *Processor) ChunkMarkdownWithOptions(markdown string, opts ChunkOptions) []string {
	rawChunks := p.chunkMarkdownInternal(markdown, opts)
	if len(rawChunks) == 0 {
		return nil
	}

	chunks := make([]string, 0, len(rawChunks))
	for _, chunk := range rawChunks {
		if strings.TrimSpace(chunk.Text) != "" {
			chunks = append(chunks, chunk.Text)
		}
	}
	return chunks
}

// WrapMarkdownWithMetadata wraps markdown chunks with metadata for vector storage
func (p *Processor) WrapMarkdownWithMetadata(markdown string, docID string, source string, fileID uuid.UUID, createdAt time.Time) []string {
	rawChunks := p.chunkMarkdownInternal(markdown, DefaultChunkOptions())
	if len(rawChunks) == 0 {
		return nil
	}

	sanitizedSource := sanitizeMetadataValue(source)
	fileIdentifier := fileID.String()
	created := createdAt.UTC().Format(time.RFC3339)
	wrapped := make([]string, 0, len(rawChunks))

	chunkCounter := 0

	for _, chunk := range rawChunks {
		section := sanitizeMetadataValue(chunk.Section)
		if section == "" {
			section = "Document"
		}

		metadataEstimate := estimateMetadataTokens(docID, fileIdentifier, sanitizedSource, section, created)
		queue := []string{chunk.Text}

		for len(queue) > 0 {
			part := queue[0]
			queue = queue[1:]

			tokenEstimate := p.EstimateTokenCount(part) + metadataEstimate
			if tokenEstimate > maxEmbeddingTokens-metadataTokenBuffer {
				subChunks := p.splitLargeChunk(part)
				if len(subChunks) > 1 {
					queue = append(subChunks, queue...)
					continue
				}
			}

			if strings.TrimSpace(part) == "" {
				continue
			}

			frontMatter := fmt.Sprintf("---\ndoc_id: %s\nfile_id: %s\nsource: \"%s\"\nsection: \"%s\"\nchunk_index: %d\ncreated_at: %s\n---\n\n",
				docID, fileIdentifier, sanitizedSource, section, chunkCounter, created)
			wrapped = append(wrapped, frontMatter+part)
			chunkCounter++
		}
	}
	return wrapped
}

func estimateMetadataTokens(docID, fileID, source, section, created string) int {
	frontMatter := fmt.Sprintf("---\ndoc_id: %s\nfile_id: %s\nsource: \"%s\"\nsection: \"%s\"\nchunk_index: 0\ncreated_at: %s\n---\n\n",
		docID, fileID, source, section, created)
	return int(math.Ceil(float64(len(frontMatter)) / float64(defaultCharsPerTokenEst)))
}

func (p *Processor) splitLargeChunk(text string) []string {
	chunkOptions := DefaultChunkOptions()
	safeTokenBudget := maxEmbeddingTokens - metadataTokenBuffer
	if safeTokenBudget <= 0 {
		safeTokenBudget = maxEmbeddingTokens
	}
	safeChunkChars := safeTokenBudget * chunkOptions.CharsPerToken
	if safeChunkChars <= 0 {
		safeChunkChars = maxEmbeddingTokens * defaultCharsPerTokenEst
	}

	overlap := int(math.Round(float64(safeChunkChars) * chunkOptions.OverlapPercent))
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= safeChunkChars {
		overlap = safeChunkChars / 4
	}

	chunks := p.ChunkTextWithOverlap(text, safeChunkChars, overlap)
	if len(chunks) > 1 {
		return chunks
	}

	return fallbackSplit(text, safeChunkChars)
}

func fallbackSplit(text string, chunkSize int) []string {
	if chunkSize <= 0 {
		return []string{text}
	}

	runes := []rune(text)
	if len(runes) == 0 {
		return []string{text}
	}

	var result []string
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		result = append(result, string(runes[start:end]))
	}
	if len(result) == 0 {
		return []string{text}
	}
	return result
}

// chunkMarkdownInternal performs the actual markdown chunking logic
func (p *Processor) chunkMarkdownInternal(markdown string, opts ChunkOptions) []MarkdownChunk {
	maxChars := opts.MaxTokens * opts.CharsPerToken
	minChars := opts.MinTokens * opts.CharsPerToken

	lines := strings.Split(markdown, "\n")
	if len(lines) == 0 {
		return nil
	}

	var chunks []MarkdownChunk
	var currentLines []string
	currentLength := 0
	currentSection := ""
	activeSection := ""
	inCodeBlock := false
	inTable := false
	lastBlankIndex := -1

	flush := func(force bool) {
		text := strings.TrimSpace(strings.Join(currentLines, "\n"))
		if text == "" {
			currentLines = nil
			currentLength = 0
			currentSection = ""
			lastBlankIndex = -1
			return
		}

		section := currentSection
		if section == "" {
			section = activeSection
		}
		if section == "" {
			section = "Document"
		}

		chunks = append(chunks, MarkdownChunk{Section: section, Text: text})
		currentLines = nil
		currentLength = 0
		currentSection = ""
		lastBlankIndex = -1
	}

	appendLine := func(line string) {
		if len(currentLines) == 0 {
			currentSection = activeSection
		}
		currentLines = append(currentLines, line)
		currentLength += len(line) + 1
		if strings.TrimSpace(line) == "" {
			lastBlankIndex = len(currentLines) - 1
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle code blocks
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				inCodeBlock = true
			} else {
				inCodeBlock = false
			}
		}

		// Handle headings (only when not in code block)
		isHeading := !inCodeBlock && strings.HasPrefix(trimmed, "#")
		if isHeading {
			level := strings.TrimLeft(trimmed, "#")
			activeSection = strings.TrimSpace(level)
			if len(currentLines) > 0 {
				flush(true)
			}
			appendLine(line)
			continue
		}

		// Handle tables (only when not in code block)
		if !inCodeBlock {
			if strings.Contains(line, "|") && strings.Count(line, "|") >= 2 && trimmed != "" {
				inTable = true
			} else if trimmed == "" {
				inTable = false
			}
		}

		appendLine(line)

		// Skip chunking logic while in code blocks or tables
		if inCodeBlock || inTable {
			continue
		}

		// Check if we need to split due to size
		if currentLength >= maxChars {
			splitIndex := len(currentLines)
			if lastBlankIndex >= 0 {
				splitIndex = lastBlankIndex + 1
			}
			if splitIndex <= 0 || splitIndex > len(currentLines) {
				splitIndex = len(currentLines)
			}

			chunkLines := make([]string, splitIndex)
			copy(chunkLines, currentLines[:splitIndex])
			remainder := currentLines[splitIndex:]

			currentLines = chunkLines
			flush(true)

			currentLines = append([]string{}, remainder...)
			currentLength = 0
			lastBlankIndex = -1
			for idx, remLine := range currentLines {
				currentLength += len(remLine) + 1
				if strings.TrimSpace(remLine) == "" {
					lastBlankIndex = idx
				}
			}
			if len(currentLines) == 0 {
				currentSection = ""
			} else if currentSection == "" {
				currentSection = activeSection
			}
			continue
		}

		// Split at natural boundaries when we have enough content
		if currentLength >= minChars && trimmed == "" {
			flush(false)
		}
	}

	// Flush any remaining content
	if len(currentLines) > 0 {
		flush(true)
	}

	return chunks
}

// sanitizeMetadataValue cleans metadata values for YAML front matter
func sanitizeMetadataValue(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\"", "'")
	return value
}

// ChunkTextWithOverlap chunks text with configurable overlap between chunks
func (p *Processor) ChunkTextWithOverlap(text string, chunkSize int, overlapSize int) []string {
	if chunkSize <= 0 {
		chunkSize = 1000
	}
	if overlapSize < 0 {
		overlapSize = 0
	}
	if overlapSize >= chunkSize {
		overlapSize = chunkSize / 4 // Default to 25% overlap if invalid
	}

	var chunks []string
	runes := []rune(text)
	step := chunkSize - overlapSize

	for start := 0; start < len(runes); start += step {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}

		chunk := string(runes[start:end])
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}

		// Break if we've reached the end
		if end >= len(runes) {
			break
		}
	}

	return chunks
}

// EstimateTokenCount estimates the number of tokens in text
func (p *Processor) EstimateTokenCount(text string) int {
	const charsPerToken = 4
	return len(text) / charsPerToken
}

// SplitOnSentences splits text on sentence boundaries for better chunk quality
func (p *Processor) SplitOnSentences(text string) []string {
	sentences := []string{}
	current := ""

	runes := []rune(text)
	for i, r := range runes {
		current += string(r)

		// Look for sentence endings
		if r == '.' || r == '!' || r == '?' {
			// Check if this is actually the end of a sentence
			if i < len(runes)-1 {
				next := runes[i+1]
				// If followed by whitespace and capital letter, likely sentence end
				if (next == ' ' || next == '\n') && i < len(runes)-2 {
					afterNext := runes[i+2]
					if afterNext >= 'A' && afterNext <= 'Z' {
						sentences = append(sentences, strings.TrimSpace(current))
						current = ""
					}
				}
			} else {
				// End of text
				sentences = append(sentences, strings.TrimSpace(current))
				current = ""
			}
		}
	}

	// Add any remaining text
	if strings.TrimSpace(current) != "" {
		sentences = append(sentences, strings.TrimSpace(current))
	}

	return sentences
}
