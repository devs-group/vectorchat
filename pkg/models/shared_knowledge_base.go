package models

import (
	"time"

	"github.com/google/uuid"
)

type SharedKnowledgeBaseCreateRequest struct {
	Name        string  `json:"name" binding:"required" example:"Company Handbook"`
	Description *string `json:"description,omitempty" example:"Central policies and guidelines"`
}

type SharedKnowledgeBaseUpdateRequest struct {
	Name        *string `json:"name,omitempty" example:"Updated Handbook"`
	Description *string `json:"description,omitempty" example:"Additional onboarding details"`
}

type SharedKnowledgeBaseResponse struct {
	ID          uuid.UUID `json:"id" example:"3f5f5f4e-1234-5678-a9ab-0123456789ab"`
	OwnerID     string    `json:"owner_id" example:"user_123"`
	Name        string    `json:"name" example:"Support FAQs"`
	Description *string   `json:"description,omitempty" example:"Frequently asked questions for agents"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

type SharedKnowledgeBaseListResponse struct {
	KnowledgeBases []SharedKnowledgeBaseResponse `json:"knowledge_bases"`
}

type ChatbotKnowledgeBaseLinkRequest struct {
	SharedKnowledgeBaseIDs []uuid.UUID `json:"shared_knowledge_base_ids"`
}

type SharedKnowledgeBaseFileUploadResponse struct {
	Message         string    `json:"message" example:"File processed successfully"`
	KnowledgeBaseID uuid.UUID `json:"knowledge_base_id"`
	File            string    `json:"file" example:"document.pdf"`
	Filename        string    `json:"filename,omitempty" example:"document.pdf"`
	Size            int64     `json:"size,omitempty" example:"1024"`
}

type SharedKnowledgeBaseFilesResponse struct {
	KnowledgeBaseID uuid.UUID  `json:"knowledge_base_id"`
	Files           []FileInfo `json:"files"`
}

type SharedKnowledgeBaseTextSourcesResponse struct {
	KnowledgeBaseID uuid.UUID        `json:"knowledge_base_id"`
	Sources         []TextSourceInfo `json:"sources"`
}
