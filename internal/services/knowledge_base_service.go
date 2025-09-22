package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/crawler"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/vectorize"
	"github.com/yourusername/vectorchat/pkg/docprocessor"
)

type KnowledgeBaseTarget struct {
	ChatbotID             *uuid.UUID
	SharedKnowledgeBaseID *uuid.UUID
}

func (t KnowledgeBaseTarget) validate() error {
	hasChatbot := t.ChatbotID != nil && *t.ChatbotID != uuid.Nil
	hasShared := t.SharedKnowledgeBaseID != nil && *t.SharedKnowledgeBaseID != uuid.Nil

	if hasChatbot == hasShared { // either both true or both false
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "knowledge base target must reference exactly one scope")
	}
	return nil
}

func (t KnowledgeBaseTarget) namespace() string {
	if t.ChatbotID != nil && *t.ChatbotID != uuid.Nil {
		return t.ChatbotID.String()
	}
	if t.SharedKnowledgeBaseID != nil && *t.SharedKnowledgeBaseID != uuid.Nil {
		return fmt.Sprintf("shared-%s", t.SharedKnowledgeBaseID.String())
	}
	return "unknown"
}

func (t KnowledgeBaseTarget) fileOwner() (chatbotID *uuid.UUID, sharedID *uuid.UUID) {
	if t.ChatbotID != nil && *t.ChatbotID != uuid.Nil {
		return t.ChatbotID, nil
	}
	return nil, t.SharedKnowledgeBaseID
}

// KnowledgeBaseService provides ingestion utilities for chatbot and shared knowledge bases.
type KnowledgeBaseService struct {
	fileRepo     *db.FileRepository
	documentRepo *db.DocumentRepository
	vectorizer   vectorize.Vectorizer
	docProcessor *docprocessor.Processor
	webCrawler   crawler.WebCrawler
	db           *db.Database

	crawlerDisabled atomic.Bool
}

func NewKnowledgeBaseService(
	fileRepo *db.FileRepository,
	documentRepo *db.DocumentRepository,
	vectorizer vectorize.Vectorizer,
	docProcessor *docprocessor.Processor,
	webCrawler crawler.WebCrawler,
	database *db.Database,
) *KnowledgeBaseService {
	return &KnowledgeBaseService{
		fileRepo:     fileRepo,
		documentRepo: documentRepo,
		vectorizer:   vectorizer,
		docProcessor: docProcessor,
		webCrawler:   webCrawler,
		db:           database,
	}
}

// IngestFile converts, chunks, and indexes an uploaded file into the specified knowledge base.
func (s *KnowledgeBaseService) IngestFile(ctx context.Context, target KnowledgeBaseTarget, fileHeader *multipart.FileHeader) (*db.File, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	processed, err := s.docProcessor.ProcessFile(ctx, fileHeader)
	if err != nil {
		return nil, err
	}

	return s.storeProcessedMarkdown(ctx, target, processed.Filename, processed.OriginalSize, processed.Markdown, processed.Hash, processed.ProcessedAt)
}

// IngestText chunks and indexes arbitrary text into the target knowledge base.
func (s *KnowledgeBaseService) IngestText(ctx context.Context, target KnowledgeBaseTarget, text string) (*db.File, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	processed, err := s.docProcessor.ProcessText(text)
	if err != nil {
		return nil, err
	}

	return s.storeProcessedMarkdown(ctx, target, processed.Filename, processed.OriginalSize, processed.Markdown, processed.Hash, processed.ProcessedAt)
}

// IngestWebsite crawls a website starting from rootURL and indexes discovered content.
func (s *KnowledgeBaseService) IngestWebsite(ctx context.Context, target KnowledgeBaseTarget, rootURL string) (*db.File, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}
	if strings.TrimSpace(rootURL) == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "url is required")
	}

	host := rootURL
	if u, err := s.ParseURL(rootURL); err == nil {
		host = u.Hostname()
	}

	fileID := uuid.New()
	file := &db.File{
		ID:                    fileID,
		Filename:              fmt.Sprintf("website-%s-%s", host, time.Now().Format("20060102-150405")),
		UploadedAt:            time.Now().UTC(),
		SizeBytes:             0,
		ChatbotID:             nil,
		SharedKnowledgeBaseID: nil,
	}
	if chatbotID, sharedID := target.fileOwner(); chatbotID != nil {
		file.ChatbotID = chatbotID
	} else {
		file.SharedKnowledgeBaseID = sharedID
	}

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.fileRepo.CreateTx(ctx, tx, file); err != nil {
		return nil, apperrors.Wrap(err, "failed to insert website source")
	}

	pages, err := s.crawlWebsite(ctx, rootURL, crawler.Options{MaxPages: 25, MaxDepth: 2, Timeout: 40 * time.Second})
	if err != nil {
		return nil, err
	}

	const chunkSize = 1000
	var totalBytes int64
	for pi, page := range pages {
		if strings.TrimSpace(page.Text) == "" {
			continue
		}
		chunks := s.docProcessor.ChunkText(page.Text, chunkSize)
		for ci, chunk := range chunks {
			totalBytes += int64(len(chunk))
			emb, err := s.vectorizer.VectorizeText(ctx, chunk)
			if err != nil {
				return nil, apperrors.Wrapf(err, "failed to vectorize page %d chunk %d", pi, ci)
			}

			docID := fmt.Sprintf("%s-web-%d-%d-%s", target.namespace(), pi, ci, uuid.New().String())
			doc := &db.DocumentWithEmbedding{
				ID:                    docID,
				Content:               []byte(chunk),
				Embedding:             emb,
				ChatbotID:             target.ChatbotID,
				SharedKnowledgeBaseID: target.SharedKnowledgeBaseID,
				FileID:                &fileID,
				ChunkIndex:            intPtr(ci),
			}

			if err := s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc); err != nil {
				return nil, apperrors.Wrapf(err, "failed to store page %d chunk %d", pi, ci)
			}
		}
	}

	file.SizeBytes = totalBytes
	if err := s.fileRepo.UpdateTx(ctx, tx, file); err != nil {
		return nil, apperrors.Wrap(err, "failed to update website source size")
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Wrap(err, "failed to commit website ingestion")
	}

	return file, nil
}

func (s *KnowledgeBaseService) storeProcessedMarkdown(ctx context.Context, target KnowledgeBaseTarget, originalFilename string, originalSize int64, markdown string, docID string, ingestedAt time.Time) (*db.File, error) {
	if err := target.validate(); err != nil {
		return nil, err
	}

	fileID := uuid.New()
	baseName := filepath.Base(originalFilename)
	file := &db.File{
		ID:        fileID,
		Filename:  baseName,
		SizeBytes: originalSize,
		UploadedAt: func() time.Time {
			if ingestedAt.IsZero() {
				return time.Now().UTC()
			}
			return ingestedAt.UTC()
		}(),
	}
	if chatbotID, sharedID := target.fileOwner(); chatbotID != nil {
		file.ChatbotID = chatbotID
	} else {
		file.SharedKnowledgeBaseID = sharedID
	}

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.fileRepo.CreateTx(ctx, tx, file); err != nil {
		return nil, apperrors.Wrap(err, "failed to insert file metadata")
	}

	chunks := s.docProcessor.WrapMarkdownWithMetadata(markdown, docID, baseName, fileID, file.UploadedAt)
	if len(chunks) == 0 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file did not produce any indexable content")
	}

	docIDBase := fmt.Sprintf("%s-%s", target.namespace(), baseName)
	for idx, chunk := range chunks {
		emb, err := s.vectorizer.VectorizeText(ctx, chunk)
		if err != nil {
			return nil, apperrors.Wrapf(err, "failed to vectorize chunk %d", idx)
		}

		doc := &db.DocumentWithEmbedding{
			ID:                    fmt.Sprintf("%s-%d", docIDBase, idx),
			Content:               []byte(chunk),
			Embedding:             emb,
			ChatbotID:             target.ChatbotID,
			SharedKnowledgeBaseID: target.SharedKnowledgeBaseID,
			FileID:                &fileID,
			ChunkIndex:            intPtr(idx),
		}

		if err := s.documentRepo.StoreWithEmbeddingTx(ctx, tx, doc); err != nil {
			return nil, apperrors.Wrapf(err, "failed to store chunk %d", idx)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, apperrors.Wrap(err, "failed to commit knowledge ingestion")
	}

	return file, nil
}

func (s *KnowledgeBaseService) crawlWebsite(ctx context.Context, rootURL string, opts crawler.Options) ([]crawler.Page, error) {
	if s.webCrawler != nil && !s.crawlerDisabled.Load() {
		pages, err := s.webCrawler.Crawl(ctx, rootURL, opts)
		if err == nil && len(pages) > 0 {
			return pages, nil
		}
		if err != nil {
			disable := false
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				disable = errors.Is(opErr.Err, syscall.ECONNREFUSED)
			}
			if disable && s.crawlerDisabled.CompareAndSwap(false, true) {
				slog.Info("crawl4ai unavailable; disabling external crawler", "url", rootURL, "error", err)
			} else {
				slog.Warn("crawl4ai crawl failed; using fallback crawler", "error", err, "url", rootURL)
			}
		} else {
			slog.Warn("crawl4ai returned no pages; using fallback crawler", "url", rootURL)
		}
	}
	return crawler.CrawlWebsite(ctx, rootURL, opts)
}

// ParseURL parses a string URL for reuse in handlers/services.
func (s *KnowledgeBaseService) ParseURL(raw string) (*url.URL, error) {
	return url.Parse(raw)
}
