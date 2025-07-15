package db

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// APIKeyRepository defines the interface for API key data operations
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	FindByID(ctx context.Context, id string) (*APIKey, error)
	FindByUserID(ctx context.Context, userID string) ([]*APIKey, error)
	FindByUserIDWithPagination(ctx context.Context, userID string, offset, limit int) ([]*APIKey, int64, error)
	FindByHashComparison(ctx context.Context, compareFunc func(hashedKey string) (bool, error)) (*APIKey, error)
	Revoke(ctx context.Context, id, userID string) error
	Delete(ctx context.Context, id string) error
	IsRevoked(ctx context.Context, id string) (bool, error)
	IsExpired(ctx context.Context, id string) (bool, error)
}

// ChatbotRepository defines the interface for chatbot data operations
type ChatbotRepository interface {
	Create(ctx context.Context, chatbot *Chatbot) error
	FindByID(ctx context.Context, id uuid.UUID) (*Chatbot, error)
	FindByIDAndUserID(ctx context.Context, id uuid.UUID, userID string) (*Chatbot, error)
	FindByUserID(ctx context.Context, userID string) ([]*Chatbot, error)
	FindByUserIDWithPagination(ctx context.Context, userID string, offset, limit int) ([]*Chatbot, int64, error)
	Update(ctx context.Context, chatbot *Chatbot) error
	Delete(ctx context.Context, id uuid.UUID, userID string) error
	CheckOwnership(ctx context.Context, id uuid.UUID, userID string) (bool, error)
	UpdateBasicInfo(ctx context.Context, id uuid.UUID, userID, name, description string) error
	UpdateSystemInstructions(ctx context.Context, id uuid.UUID, userID, instructions string) error
	UpdateModelSettings(ctx context.Context, id uuid.UUID, userID, modelName string, temperature float64, maxTokens int) error
}

// DocumentRepository defines the interface for document data operations
type DocumentRepository interface {
	Store(ctx context.Context, doc *Document) error
	StoreWithEmbedding(ctx context.Context, doc *DocumentWithEmbedding) error
	FindByID(ctx context.Context, id string) (*Document, error)
	FindSimilar(ctx context.Context, embedding []float32, limit int) ([]*DocumentWithEmbedding, error)
	FindSimilarByChatbot(ctx context.Context, embedding []float32, chatbotID string, limit int) ([]*DocumentWithEmbedding, error)
	FindByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*Document, error)
	FindByFileID(ctx context.Context, fileID uuid.UUID) ([]*Document, error)
	Delete(ctx context.Context, id string) error
	DeleteByChatbotID(ctx context.Context, chatbotID uuid.UUID) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error)
}

// FileRepository defines the interface for file data operations
type FileRepository interface {
	Create(ctx context.Context, file *File) error
	FindByID(ctx context.Context, id uuid.UUID) (*File, error)
	FindByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]*File, error)
	FindByChatbotIDWithPagination(ctx context.Context, chatbotID uuid.UUID, offset, limit int) ([]*File, int64, error)
	Update(ctx context.Context, file *File) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByChatbotID(ctx context.Context, chatbotID uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	CountByChatbotID(ctx context.Context, chatbotID uuid.UUID) (int64, error)
}

// Repositories holds all repository interfaces
type Repositories struct {
	User     UserRepository
	APIKey   APIKeyRepository
	Chatbot  ChatbotRepository
	Document DocumentRepository
	File     FileRepository
}

// NewRepositories creates a new instance of all repositories
func NewRepositories(db *Database) *Repositories {
	return &Repositories{
		User:     NewUserRepository(db),
		APIKey:   NewAPIKeyRepository(db),
		Chatbot:  NewChatbotRepository(db),
		Document: NewDocumentRepository(db),
		File:     NewFileRepository(db),
	}
}

// Transaction-aware repository interfaces for operations that need transactions

// UserRepositoryTx defines transaction-aware user operations
type UserRepositoryTx interface {
	UserRepository
	CreateTx(ctx context.Context, tx *Transaction, user *User) error
	UpdateTx(ctx context.Context, tx *Transaction, user *User) error
	DeleteTx(ctx context.Context, tx *Transaction, id string) error
}

// APIKeyRepositoryTx defines transaction-aware API key operations
type APIKeyRepositoryTx interface {
	APIKeyRepository
	CreateTx(ctx context.Context, tx *Transaction, apiKey *APIKey) error
	RevokeTx(ctx context.Context, tx *Transaction, id, userID string) error
	DeleteTx(ctx context.Context, tx *Transaction, id string) error
}

// ChatbotRepositoryTx defines transaction-aware chatbot operations
type ChatbotRepositoryTx interface {
	ChatbotRepository
	CreateTx(ctx context.Context, tx *Transaction, chatbot *Chatbot) error
	UpdateTx(ctx context.Context, tx *Transaction, chatbot *Chatbot) error
	DeleteTx(ctx context.Context, tx *Transaction, id uuid.UUID, userID string) error
}

// DocumentRepositoryTx defines transaction-aware document operations
type DocumentRepositoryTx interface {
	DocumentRepository
	StoreTx(ctx context.Context, tx *Transaction, doc *Document) error
	StoreWithEmbeddingTx(ctx context.Context, tx *Transaction, doc *DocumentWithEmbedding) error
	DeleteTx(ctx context.Context, tx *Transaction, id string) error
	DeleteByChatbotIDTx(ctx context.Context, tx *Transaction, chatbotID uuid.UUID) error
	DeleteByFileIDTx(ctx context.Context, tx *Transaction, fileID uuid.UUID) error
}

// FileRepositoryTx defines transaction-aware file operations
type FileRepositoryTx interface {
	FileRepository
	CreateTx(ctx context.Context, tx *Transaction, file *File) error
	UpdateTx(ctx context.Context, tx *Transaction, file *File) error
	DeleteTx(ctx context.Context, tx *Transaction, id uuid.UUID) error
	DeleteByChatbotIDTx(ctx context.Context, tx *Transaction, chatbotID uuid.UUID) error
}

// RepositoriesTx holds all transaction-aware repository interfaces
type RepositoriesTx struct {
	User     UserRepositoryTx
	APIKey   APIKeyRepositoryTx
	Chatbot  ChatbotRepositoryTx
	Document DocumentRepositoryTx
	File     FileRepositoryTx
}

// NewRepositoriesTx creates a new instance of all transaction-aware repositories
func NewRepositoriesTx(db *Database) *RepositoriesTx {
	return &RepositoriesTx{
		User:     NewUserRepository(db),
		APIKey:   NewAPIKeyRepository(db),
		Chatbot:  NewChatbotRepository(db),
		Document: NewDocumentRepository(db),
		File:     NewFileRepository(db),
	}
}

// Repository operation options
type FindOptions struct {
	Offset int
	Limit  int
	Order  string
	Where  map[string]interface{}
}

// DefaultFindOptions returns default find options
func DefaultFindOptions() *FindOptions {
	return &FindOptions{
		Offset: 0,
		Limit:  50,
		Order:  "created_at DESC",
		Where:  make(map[string]interface{}),
	}
}

// CreateOptions represents options for create operations
type CreateOptions struct {
	SkipValidation   bool
	UpdateOnConflict bool
}

// UpdateOptions represents options for update operations
type UpdateOptions struct {
	PartialUpdate  bool
	OptimisticLock bool
}

// DeleteOptions represents options for delete operations
type DeleteOptions struct {
	SoftDelete bool
	Cascade    bool
}
