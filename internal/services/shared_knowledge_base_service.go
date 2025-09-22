package services

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/pkg/models"
)

type SharedKnowledgeBaseService struct {
	*CommonService
	repo         *db.SharedKnowledgeBaseRepository
	fileRepo     *db.FileRepository
	documentRepo *db.DocumentRepository
	ingestion    *KnowledgeBaseService
}

func NewSharedKnowledgeBaseService(
	repo *db.SharedKnowledgeBaseRepository,
	fileRepo *db.FileRepository,
	documentRepo *db.DocumentRepository,
	ingestion *KnowledgeBaseService,
) *SharedKnowledgeBaseService {
	return &SharedKnowledgeBaseService{
		CommonService: NewCommonService(),
		repo:          repo,
		fileRepo:      fileRepo,
		documentRepo:  documentRepo,
		ingestion:     ingestion,
	}
}

func (s *SharedKnowledgeBaseService) Create(ctx context.Context, ownerID string, req *models.SharedKnowledgeBaseCreateRequest) (*models.SharedKnowledgeBaseResponse, error) {
	if ownerID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "owner id is required")
	}
	if req == nil || req.Name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	kb := &db.SharedKnowledgeBase{
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, kb); err != nil {
		return nil, err
	}

	return toSharedKnowledgeBaseResponse(kb), nil
}

func (s *SharedKnowledgeBaseService) Update(ctx context.Context, ownerID string, kbID uuid.UUID, req *models.SharedKnowledgeBaseUpdateRequest) (*models.SharedKnowledgeBaseResponse, error) {
	kb, err := s.ensureOwnership(ctx, ownerID, kbID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name cannot be empty")
		}
		kb.Name = *req.Name
	}
	if req.Description != nil {
		kb.Description = req.Description
	}

	if err := s.repo.Update(ctx, kb); err != nil {
		return nil, err
	}

	return toSharedKnowledgeBaseResponse(kb), nil
}

func (s *SharedKnowledgeBaseService) Delete(ctx context.Context, ownerID string, kbID uuid.UUID) error {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, kbID)
}

func (s *SharedKnowledgeBaseService) Get(ctx context.Context, ownerID string, kbID uuid.UUID) (*models.SharedKnowledgeBaseResponse, error) {
	kb, err := s.ensureOwnership(ctx, ownerID, kbID)
	if err != nil {
		return nil, err
	}
	return toSharedKnowledgeBaseResponse(kb), nil
}

func (s *SharedKnowledgeBaseService) List(ctx context.Context, ownerID string) (*models.SharedKnowledgeBaseListResponse, error) {
	kbs, err := s.repo.ListByOwner(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	responses := make([]models.SharedKnowledgeBaseResponse, 0, len(kbs))
	for _, kb := range kbs {
		responses = append(responses, *toSharedKnowledgeBaseResponse(kb))
	}

	return &models.SharedKnowledgeBaseListResponse{KnowledgeBases: responses}, nil
}

func (s *SharedKnowledgeBaseService) ListFiles(ctx context.Context, ownerID string, kbID uuid.UUID) (*models.SharedKnowledgeBaseFilesResponse, error) {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return nil, err
	}

	files, err := s.fileRepo.FindNonTextBySharedKnowledgeBaseID(ctx, kbID)
	if err != nil {
		return nil, err
	}

	respFiles := make([]models.FileInfo, 0, len(files))
	for _, f := range files {
		respFiles = append(respFiles, models.FileInfo{
			Filename:   f.Filename,
			ID:         f.ID,
			Size:       f.SizeBytes,
			UploadedAt: f.UploadedAt,
		})
	}

	return &models.SharedKnowledgeBaseFilesResponse{
		KnowledgeBaseID: kbID,
		Files:           respFiles,
	}, nil
}

func (s *SharedKnowledgeBaseService) ListTextSources(ctx context.Context, ownerID string, kbID uuid.UUID) (*models.SharedKnowledgeBaseTextSourcesResponse, error) {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return nil, err
	}

	files, err := s.fileRepo.FindTextBySharedKnowledgeBaseID(ctx, kbID)
	if err != nil {
		return nil, err
	}

	sources := make([]models.TextSourceInfo, 0, len(files))
	for _, f := range files {
		sources = append(sources, models.TextSourceInfo{
			ID:         f.ID,
			Title:      f.Filename,
			Size:       f.SizeBytes,
			UploadedAt: f.UploadedAt,
		})
	}

	return &models.SharedKnowledgeBaseTextSourcesResponse{
		KnowledgeBaseID: kbID,
		Sources:         sources,
	}, nil
}

func (s *SharedKnowledgeBaseService) ProcessFileUpload(ctx context.Context, ownerID string, kbID uuid.UUID, fileHeader *multipart.FileHeader) (*models.SharedKnowledgeBaseFileUploadResponse, error) {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return nil, err
	}

	target := KnowledgeBaseTarget{SharedKnowledgeBaseID: &kbID}
	file, err := s.ingestion.IngestFile(ctx, target, fileHeader)
	if err != nil {
		return nil, err
	}

	return &models.SharedKnowledgeBaseFileUploadResponse{
		Message:         "File processed successfully",
		KnowledgeBaseID: kbID,
		File:            file.Filename,
		Filename:        file.Filename,
		Size:            file.SizeBytes,
	}, nil
}

func (s *SharedKnowledgeBaseService) ProcessTextUpload(ctx context.Context, ownerID string, kbID uuid.UUID, text string) error {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "text is required")
	}

	target := KnowledgeBaseTarget{SharedKnowledgeBaseID: &kbID}
	_, err := s.ingestion.IngestText(ctx, target, text)
	return err
}

func (s *SharedKnowledgeBaseService) ProcessWebsiteUpload(ctx context.Context, ownerID string, kbID uuid.UUID, rootURL string) error {
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return err
	}

	target := KnowledgeBaseTarget{SharedKnowledgeBaseID: &kbID}
	_, err := s.ingestion.IngestWebsite(ctx, target, rootURL)
	return err
}

func (s *SharedKnowledgeBaseService) DeleteFile(ctx context.Context, ownerID string, kbID uuid.UUID, filename string) error {
	if filename == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "filename is required")
	}
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return err
	}

	file, err := s.fileRepo.FindBySharedKnowledgeBaseIDAndFilename(ctx, kbID, filename)
	if err != nil {
		return err
	}

	tx, err := s.ingestion.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.documentRepo.DeleteByFileIDTx(ctx, tx, file.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete file documents")
	}
	if err := s.fileRepo.DeleteTx(ctx, tx, file.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete file metadata")
	}

	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit delete transaction")
	}

	return nil
}

func (s *SharedKnowledgeBaseService) DeleteTextSource(ctx context.Context, ownerID string, kbID uuid.UUID, sourceID string) error {
	if sourceID == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "source id is required")
	}
	if _, err := s.ensureOwnership(ctx, ownerID, kbID); err != nil {
		return err
	}

	fid, err := uuid.Parse(sourceID)
	if err != nil {
		return apperrors.Wrap(err, "invalid text source id")
	}

	file, err := s.fileRepo.FindByID(ctx, fid)
	if err != nil {
		return err
	}
	if file.SharedKnowledgeBaseID == nil || *file.SharedKnowledgeBaseID != kbID {
		return apperrors.ErrUnauthorizedKnowledgeBaseAccess
	}
	if !strings.HasPrefix(file.Filename, "text-") {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "not a text source")
	}

	tx, err := s.ingestion.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if err := s.documentRepo.DeleteByFileIDTx(ctx, tx, file.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete text documents")
	}
	if err := s.fileRepo.DeleteTx(ctx, tx, file.ID); err != nil {
		return apperrors.Wrap(err, "failed to delete text source")
	}

	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit delete transaction")
	}

	return nil
}

func (s *SharedKnowledgeBaseService) ensureOwnership(ctx context.Context, ownerID string, kbID uuid.UUID) (*db.SharedKnowledgeBase, error) {
	if ownerID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "owner id is required")
	}
	kb, err := s.repo.FindByID(ctx, kbID)
	if err != nil {
		return nil, err
	}
	if kb.OwnerID != ownerID {
		return nil, apperrors.ErrUnauthorizedKnowledgeBaseAccess
	}
	return kb, nil
}

func toSharedKnowledgeBaseResponse(kb *db.SharedKnowledgeBase) *models.SharedKnowledgeBaseResponse {
	return &models.SharedKnowledgeBaseResponse{
		ID:          kb.ID,
		OwnerID:     kb.OwnerID,
		Name:        kb.Name,
		Description: kb.Description,
		CreatedAt:   kb.CreatedAt,
		UpdatedAt:   kb.UpdatedAt,
	}
}
