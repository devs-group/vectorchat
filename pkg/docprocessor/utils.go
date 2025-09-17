package docprocessor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FileMetadata represents metadata about a processed file
type FileMetadata struct {
	ID          uuid.UUID `json:"id"`
	Filename    string    `json:"filename"`
	Extension   string    `json:"extension"`
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	ProcessedAt time.Time `json:"processed_at"`
	ChunkCount  int       `json:"chunk_count"`
	TokenCount  int       `json:"token_count"`
}

// ProcessingResult contains the result of document processing
type ProcessingResult struct {
	File     *ProcessedFile `json:"file"`
	Metadata *FileMetadata  `json:"metadata"`
	Success  bool           `json:"success"`
	Error    string         `json:"error,omitempty"`
}

// ChunkMetadata represents metadata about a text chunk
type ChunkMetadata struct {
	Index     int       `json:"index"`
	Section   string    `json:"section"`
	TokenSize int       `json:"token_size"`
	FileID    uuid.UUID `json:"file_id"`
	DocID     string    `json:"doc_id"`
}

// GenerateFileMetadata creates metadata from a processed file
func GenerateFileMetadata(pf *ProcessedFile) *FileMetadata {
	ext := filepath.Ext(pf.Filename)
	tokenCount := len(pf.Markdown) / 4 // Rough estimate

	return &FileMetadata{
		ID:          pf.ID,
		Filename:    pf.Filename,
		Extension:   ext,
		Size:        pf.OriginalSize,
		Hash:        pf.Hash,
		ProcessedAt: pf.ProcessedAt,
		ChunkCount:  len(pf.Chunks),
		TokenCount:  tokenCount,
	}
}

// IsTextFile checks if a file is a text-based source
func IsTextFile(filename string) bool {
	return strings.HasPrefix(filename, "text-") && strings.HasSuffix(filename, ".txt")
}

// IsWebsiteFile checks if a file is a website source
func IsWebsiteFile(filename string) bool {
	return strings.HasPrefix(filename, "website-")
}

// GetFileType returns the type of file based on its name
func GetFileType(filename string) string {
	if IsTextFile(filename) {
		return "text"
	}
	if IsWebsiteFile(filename) {
		return "website"
	}
	return "file"
}

// GenerateStoredFilename creates a stored filename with chatbot ID prefix
func GenerateStoredFilename(chatbotID uuid.UUID, originalFilename string) string {
	return fmt.Sprintf("%s-%s", chatbotID.String(), filepath.Base(originalFilename))
}

// ParseStoredFilename extracts the original filename from a stored filename
func ParseStoredFilename(storedFilename string, chatbotID uuid.UUID) string {
	prefix := chatbotID.String() + "-"
	if strings.HasPrefix(storedFilename, prefix) {
		return strings.TrimPrefix(storedFilename, prefix)
	}
	return storedFilename
}

// ValidateFilename checks if a filename is valid
func ValidateFilename(filename string) error {
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename cannot contain path traversal sequences")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	return nil
}

// GetSupportedFileTypes returns a list of commonly supported file types
func GetSupportedFileTypes() []string {
	return []string{
		".txt", ".md", ".pdf", ".docx", ".doc", ".rtf",
		".html", ".htm", ".xml", ".json", ".csv",
		".xlsx", ".xls", ".pptx", ".ppt",
	}
}

// FormatFileSize formats file size in human-readable format
func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GenerateDocumentID creates a unique document ID for chunks
func GenerateDocumentID(chatbotID uuid.UUID, filename string, chunkIndex int) string {
	base := fmt.Sprintf("%s-%s", chatbotID.String(), filepath.Base(filename))
	return fmt.Sprintf("%s-%d", base, chunkIndex)
}

// CleanMarkdown performs basic cleanup on markdown text
func CleanMarkdown(markdown string) string {
	// Remove excessive whitespace
	lines := strings.Split(markdown, "\n")
	var cleaned []string
	var lastWasEmpty bool

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if !lastWasEmpty {
				cleaned = append(cleaned, "")
				lastWasEmpty = true
			}
		} else {
			cleaned = append(cleaned, line)
			lastWasEmpty = false
		}
	}

	return strings.Join(cleaned, "\n")
}

// ExtractTitle attempts to extract a title from markdown content
func ExtractTitle(markdown string) string {
	lines := strings.Split(markdown, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		}
	}
	return ""
}

// CountWords counts the approximate number of words in text
func CountWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// TruncateText truncates text to a specified length with ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	if maxLength <= 3 {
		return "..."
	}

	return text[:maxLength-3] + "..."
}
