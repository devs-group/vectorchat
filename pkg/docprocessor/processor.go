package docprocessor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// Processor handles document processing operations including chunking and conversion
type Processor struct {
	markitdown   *MarkitdownClient
	extMu        sync.RWMutex
	supportedExt map[string]struct{}
}

// ProcessedFile represents a processed file with metadata
type ProcessedFile struct {
	ID           uuid.UUID
	Filename     string
	OriginalSize int64
	Hash         string
	Markdown     string
	Chunks       []string
	ProcessedAt  time.Time
}

// ChunkOptions contains options for chunking text
type ChunkOptions struct {
	MaxTokens      int
	MinTokens      int
	CharsPerToken  int
	OverlapPercent float64
}

// DefaultChunkOptions returns default chunking options
func DefaultChunkOptions() ChunkOptions {
	return ChunkOptions{
		MaxTokens:      1200,
		MinTokens:      800,
		CharsPerToken:  4,
		OverlapPercent: 0.1,
	}
}

// NewProcessor creates a new document processor
func NewProcessor(markitdown *MarkitdownClient) *Processor {
	return &Processor{
		markitdown:   markitdown,
		supportedExt: make(map[string]struct{}),
	}
}

// ProcessFile processes an uploaded file by converting to markdown and chunking
func (p *Processor) ProcessFile(ctx context.Context, fileHeader *multipart.FileHeader) (*ProcessedFile, error) {
	// Validate file size
	const maxFileBytes = 10 * 1024 * 1024 // 10 MB
	if fileHeader.Size > maxFileBytes {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file exceeds maximum size (10MB)")
	}

	filename := filepath.Base(fileHeader.Filename)
	if filename == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file name is required")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file extension is required")
	}

	// Check if extension is supported
	if err := p.ensureSupportedExtensions(ctx); err != nil {
		return nil, apperrors.Wrap(err, "failed to load supported file types")
	}
	if !p.isExtensionSupported(ext) {
		return nil, apperrors.Wrapf(apperrors.ErrInvalidChatbotParameters, "unsupported file type: %s", ext)
	}

	// Read file content
	src, err := fileHeader.Open()
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	var buf bytes.Buffer
	hasher := sha256.New()
	if _, err := io.Copy(io.MultiWriter(&buf, hasher), src); err != nil {
		return nil, apperrors.Wrap(err, "failed to read file")
	}

	// Convert to markdown
	markdown, err := p.convertFileToMarkdown(ctx, filename, buf.Bytes())
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to convert file to markdown")
	}

	// Chunk the markdown
	chunks := p.ChunkMarkdown(markdown)
	if len(chunks) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file did not produce any indexable content")
	}

	return &ProcessedFile{
		ID:           uuid.New(),
		Filename:     filename,
		OriginalSize: fileHeader.Size,
		Hash:         hex.EncodeToString(hasher.Sum(nil)),
		Markdown:     markdown,
		Chunks:       chunks,
		ProcessedAt:  time.Now().UTC(),
	}, nil
}

// ProcessText processes plain text by chunking it
func (p *Processor) ProcessText(text string) (*ProcessedFile, error) {
	if strings.TrimSpace(text) == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text is required")
	}

	const maxTextLength = 200_000 // 200 KB
	if len(text) > maxTextLength {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text exceeds maximum allowed length")
	}

	chunks := p.ChunkText(text, 1000)
	if len(chunks) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text did not produce any chunks")
	}

	hasher := sha256.New()
	hasher.Write([]byte(text))

	return &ProcessedFile{
		ID:           uuid.New(),
		Filename:     fmt.Sprintf("text-%s.txt", time.Now().Format("20060102-150405")),
		OriginalSize: int64(len(text)),
		Hash:         hex.EncodeToString(hasher.Sum(nil)),
		Markdown:     text,
		Chunks:       chunks,
		ProcessedAt:  time.Now().UTC(),
	}, nil
}

// SaveFileToDirectory saves a file to the specified directory with a prefix
func (p *Processor) SaveFileToDirectory(fileHeader *multipart.FileHeader, directory, prefix string) (string, error) {
	filename := filepath.Base(fileHeader.Filename)
	storedName := fmt.Sprintf("%s-%s", prefix, filename)
	storedPath := filepath.Join(directory, storedName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", apperrors.Wrap(err, "failed to open uploaded file")
	}
	defer src.Close()

	dst, err := os.Create(storedPath)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(storedPath) // Clean up on error
		return "", apperrors.Wrap(err, "failed to save file")
	}

	return storedPath, nil
}

// DeleteFile removes a file from the filesystem
func (p *Processor) DeleteFile(filepath string) error {
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return apperrors.Wrap(err, "failed to delete file")
	}
	return nil
}

// ensureSupportedExtensions loads supported file extensions from markitdown
func (p *Processor) ensureSupportedExtensions(ctx context.Context) error {
	if p.markitdown == nil {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "markitdown client is not configured")
	}

	p.extMu.RLock()
	if len(p.supportedExt) > 0 {
		p.extMu.RUnlock()
		return nil
	}
	p.extMu.RUnlock()

	p.extMu.Lock()
	defer p.extMu.Unlock()
	if len(p.supportedExt) > 0 {
		return nil
	}

	exts, err := p.markitdown.SupportedExtensions(ctx)
	if err != nil {
		return err
	}
	m := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		m[ext] = struct{}{}
	}
	p.supportedExt = m
	return nil
}

// isExtensionSupported checks if a file extension is supported
func (p *Processor) isExtensionSupported(ext string) bool {
	p.extMu.RLock()
	_, ok := p.supportedExt[ext]
	p.extMu.RUnlock()
	return ok
}

// convertFileToMarkdown converts a file to markdown using the markitdown client
func (p *Processor) convertFileToMarkdown(ctx context.Context, filename string, data []byte) (string, error) {
	if p.markitdown == nil {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "markitdown client is not configured")
	}
	markdown, err := p.markitdown.Convert(ctx, filename, data)
	if err != nil {
		return "", err
	}
	markdown = strings.TrimSpace(markdown)
	if markdown == "" {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "converted markdown is empty")
	}
	return markdown, nil
}

// GetSupportedExtensions returns the list of supported file extensions
func (p *Processor) GetSupportedExtensions(ctx context.Context) ([]string, error) {
	if err := p.ensureSupportedExtensions(ctx); err != nil {
		return nil, err
	}

	p.extMu.RLock()
	defer p.extMu.RUnlock()

	extensions := make([]string, 0, len(p.supportedExt))
	for ext := range p.supportedExt {
		extensions = append(extensions, ext)
	}
	return extensions, nil
}
