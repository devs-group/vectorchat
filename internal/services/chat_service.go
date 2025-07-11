package services

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pgvector/pgvector-go"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/internal/vectorize"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	pool       *pgxpool.Pool
	vectorizer vectorize.Vectorizer
	openaiKey  string
}

func NewChatService(pool *pgxpool.Pool, vectorizer vectorize.Vectorizer, openaiKey string) *ChatService {
	return &ChatService{
		pool:       pool,
		vectorizer: vectorizer,
		openaiKey:  openaiKey,
	}
}

// AddFile adds a file to the vector database
func (c *ChatService) AddFile(ctx context.Context, id string, filePath string, chatbotID uuid.UUID) error {
	fileID := uuid.New()
	filename := filepath.Base(filePath)

	// Insert into files table
	err := c.InsertFile(ctx, fileID, chatbotID, filename)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert file metadata")
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext == ".pdf" {
		// Extract text from PDF
		pdfText, err := vectorize.ExtractTextFromPDF(filePath)
		if err != nil {
			return apperrors.Wrap(err, "failed to extract text from PDF")
		}
		// Chunk the text (e.g., 1000 chars per chunk)
		const chunkSize = 1000
		chunks := chunkText(pdfText, chunkSize)
		for i, chunk := range chunks {
			embedding, err := c.vectorizer.VectorizeText(ctx, chunk)
			if err != nil {
				return apperrors.Wrapf(err, "failed to vectorize PDF chunk %d", i)
			}
			doc := Document{
				ID:         fmt.Sprintf("%s-%d", id, i),
				Content:    []byte(chunk),
				Embedding:  embedding,
				ChatbotID:  chatbotID,
				FileID:     &fileID,
				ChunkIndex: intPtr(i),
			}
			if err := c.StoreDocument(ctx, doc); err != nil {
				return apperrors.Wrapf(err, "failed to store PDF chunk %d", i)
			}
		}
		return nil
	}

	embedding, err := c.vectorizer.VectorizeFile(ctx, filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize file")
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to read file")
	}
	doc := Document{
		ID:         id,
		Content:    content,
		Embedding:  embedding,
		ChatbotID:  chatbotID,
		FileID:     &fileID,
		ChunkIndex: nil,
	}
	return c.StoreDocument(ctx, doc)
}

// chunkText splits text into chunks of the given size
func chunkText(text string, size int) []string {
	var chunks []string
	for start := 0; start < len(text); start += size {
		end := start + size
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
	}
	return chunks
}

// intPtr returns a pointer to the given int
func intPtr(i int) *int {
	return &i
}

// ChatWithChatbot sends a message to the LLM with relevant context from a specific chatbot
func (c *ChatService) ChatWithChatbot(ctx context.Context, chatbotID, userID, message string) (string, error) {
	// Retrieve the chatbot with authorization check
	chatbot, err := c.FindChatbotByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return "", err // Already wrapped
	}

	// Check if user has access to this chatbot
	if chatbot.UserID != userID && userID != "" {
		return "", apperrors.Wrapf(apperrors.ErrUnauthorizedChatbotAccess,
			"user %s does not own chatbot %s", userID, chatbotID)
	}

	// Vectorize the query
	queryEmbedding, err := c.vectorizer.VectorizeText(ctx, message)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents for this chatbot
	docs, err := c.FindSimilarDocumentsByChatbot(ctx, queryEmbedding, chatbotID, 5)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}

	// Check if any documents were found for this chatbot
	if len(docs) == 0 {
		return "", apperrors.Wrapf(apperrors.ErrNoDocumentsFound, "chatbot ID: %s", chatbotID)
	}

	// Build context from documents
	var contextBuilder strings.Builder
	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.Content))
	}
	context := contextBuilder.String()

	// Create custom prompt with chatbot's system instructions
	promptTemplate := prompts.NewPromptTemplate(
		chatbot.SystemInstructions+"\n\n"+
			"Context information is below.\n"+
			"---------------------\n"+
			"{{.context}}\n"+
			"---------------------\n"+
			"Given the context information and not prior knowledge, answer the query.\n"+
			"Query: {{.query}}\n"+
			"Answer: ",
		[]string{"context", "query"},
	)

	// Create OpenAI client with chatbot's model settings
	llm, err := openai.New(
		openai.WithToken(c.openaiKey),
		openai.WithModel(chatbot.ModelName),
	)
	if err != nil {
		return "", apperrors.Wrap(err, "failed to create OpenAI client")
	}

	// Apply temperature separately after client creation if the package doesn't support it directly
	// Or consider using a method like this if available:
	// openai.WithModelParams(map[string]interface{}{
	//     "temperature": chatbot.TemperatureParam,
	// })

	// Format the prompt
	prompt, err := promptTemplate.Format(map[string]any{
		"context": context,
		"query":   message,
	})
	if err != nil {
		return "", apperrors.Wrap(err, "failed to format prompt")
	}

	// Generate response using the LLM with chatbot's max tokens
	completion, err := llm.Call(ctx, prompt, llms.WithMaxTokens(chatbot.MaxTokens))
	if err != nil {
		return "", apperrors.Wrap(err, "failed to generate completion")
	}

	return completion, nil
}

// CreateChatbot creates a new chatbot with default settings
func (s *ChatService) CreateChatbot(ctx context.Context, userID, name, description, systemInstructions string) (*Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}
	if name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Set default values
	if systemInstructions == "" {
		systemInstructions = "You are a helpful AI assistant."
	}

	now := time.Now()
	chatbot := Chatbot{
		ID:                 uuid.New(),
		UserID:             userID,
		Name:               name,
		Description:        description,
		SystemInstructions: systemInstructions,
		ModelName:          "gpt-3.5-turbo", // Default model
		TemperatureParam:   0.7,             // Default temperature
		MaxTokens:          2000,            // Default max tokens
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err := s.CreateChatbotDB(ctx, &chatbot)
	if err != nil {
		return nil, err // Already wrapped in the store
	}

	return &chatbot, nil
}

// GetChatbot retrieves a chatbot by ID and validates ownership
func (s *ChatService) GetChatbot(ctx context.Context, chatbotID, userID string) (*Chatbot, error) {
	chatbot, err := s.FindChatbotByIDAndUserID(ctx, uuid.MustParse(chatbotID), userID)
	if err != nil {
		return nil, err // Already wrapped in the store
	}

	// Check ownership
	if chatbot.UserID != userID {
		return nil, apperrors.Wrapf(apperrors.ErrUnauthorizedChatbotAccess,
			"user %s does not own chatbot %s", userID, chatbotID)
	}

	return chatbot, nil
}

// ListChatbots lists all chatbots owned by a user
func (s *ChatService) ListChatbots(ctx context.Context, userID string) ([]Chatbot, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "user ID is required")
	}

	return s.FindChatbotsByUserID(ctx, userID)
}

// UpdateChatbot updates a chatbot's basic information
func (s *ChatService) UpdateChatbot(ctx context.Context, chatbotID, userID, name, description string) (*Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if name == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "name is required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.Name = name
	chatbot.Description = description
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.UpdateChatbotDB(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateSystemInstructions updates a chatbot's system instructions
func (s *ChatService) UpdateSystemInstructions(ctx context.Context, chatbotID, userID, instructions string) (*Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if instructions == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "instructions are required")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.SystemInstructions = instructions
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.UpdateChatbotDB(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// UpdateModelSettings updates a chatbot's LLM model settings
func (s *ChatService) UpdateModelSettings(ctx context.Context, chatbotID, userID, modelName string, temperature float64, maxTokens int) (*Chatbot, error) {
	// Validate inputs
	if chatbotID == "" || userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}
	if modelName == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "model name is required")
	}
	if temperature < 0 || temperature > 2 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "temperature must be between 0 and 2")
	}
	if maxTokens <= 0 {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "max tokens must be positive")
	}

	// Get the existing chatbot to check ownership
	chatbot, err := s.GetChatbot(ctx, chatbotID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	chatbot.ModelName = modelName
	chatbot.TemperatureParam = temperature
	chatbot.MaxTokens = maxTokens
	chatbot.UpdatedAt = time.Now()

	// Save changes
	err = s.UpdateChatbotDB(ctx, *chatbot)
	if err != nil {
		return nil, err
	}

	return chatbot, nil
}

// DeleteChatbot deletes a chatbot
func (s *ChatService) DeleteChatbot(ctx context.Context, chatbotID, userID string) error {
	if chatbotID == "" || userID == "" {
		return apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "chatbot ID and user ID are required")
	}

	return s.DeleteChatbotDB(ctx, uuid.MustParse(chatbotID), userID)
}

// Database operations for chatbots

// CreateChatbotDB creates a new chatbot in the database
func (s *ChatService) CreateChatbotDB(ctx context.Context, chatbot *Chatbot) error {
	// Initialize a new UUID if not set
	if chatbot.ID == uuid.Nil {
		chatbot.ID = uuid.New()
	}

	query := `
		INSERT INTO chatbots (
			id, user_id, name, description, system_instructions,
			model_name, temperature_param, max_tokens
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	return s.pool.QueryRow(
		ctx,
		query,
		chatbot.ID,
		chatbot.UserID,
		chatbot.Name,
		chatbot.Description,
		chatbot.SystemInstructions,
		chatbot.ModelName,
		chatbot.TemperatureParam,
		chatbot.MaxTokens,
	).Scan(&chatbot.CreatedAt, &chatbot.UpdatedAt)
}

// FindChatbotByIDAndUserID retrieves a chatbot by its ID and user ID
func (s *ChatService) FindChatbotByIDAndUserID(ctx context.Context, id uuid.UUID, userID string) (*Chatbot, error) {
	query := `
		SELECT id, user_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens,
		       created_at, updated_at
		FROM chatbots
		WHERE id = $1 AND user_id = $2
	`

	var chatbot Chatbot
	err := s.pool.QueryRow(ctx, query, id, userID).Scan(
		&chatbot.ID,
		&chatbot.UserID,
		&chatbot.Name,
		&chatbot.Description,
		&chatbot.SystemInstructions,
		&chatbot.ModelName,
		&chatbot.TemperatureParam,
		&chatbot.MaxTokens,
		&chatbot.CreatedAt,
		&chatbot.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrChatbotNotFound
		}
		return nil, apperrors.Wrap(err, "failed to get chatbot")
	}

	return &chatbot, nil
}

// FindChatbotsByUserID retrieves all chatbots owned by a specific user
func (s *ChatService) FindChatbotsByUserID(ctx context.Context, userID string) ([]Chatbot, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, user_id, name, description, system_instructions,
			model_name, temperature_param, max_tokens, created_at, updated_at
		FROM chatbots
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to query chatbots: %v", err)
	}
	defer rows.Close()

	var chatbots []Chatbot
	for rows.Next() {
		var chatbot Chatbot
		err := rows.Scan(
			&chatbot.ID, &chatbot.UserID, &chatbot.Name, &chatbot.Description,
			&chatbot.SystemInstructions, &chatbot.ModelName, &chatbot.TemperatureParam,
			&chatbot.MaxTokens, &chatbot.CreatedAt, &chatbot.UpdatedAt,
		)
		if err != nil {
			return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to scan chatbot row: %v", err)
		}
		chatbots = append(chatbots, chatbot)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation, "error iterating chatbot rows: %v", err)
	}

	return chatbots, nil
}

// UpdateChatbotDB updates an existing chatbot
func (s *ChatService) UpdateChatbotDB(ctx context.Context, chatbot Chatbot) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE chatbots
		SET name = $1,
		description = $2,
		system_instructions = $3,
		model_name = $4,
		temperature_param = $5,
		max_tokens = $6,
		updated_at = $7
		WHERE id = $8 AND user_id = $9
	`, chatbot.Name, chatbot.Description, chatbot.SystemInstructions,
		chatbot.ModelName, chatbot.TemperatureParam, chatbot.MaxTokens,
		chatbot.UpdatedAt, chatbot.ID, chatbot.UserID)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to update chatbot: %v", err)
	}

	return nil
}

// DeleteChatbotDB deletes a chatbot by ID and owner
func (s *ChatService) DeleteChatbotDB(ctx context.Context, id uuid.UUID, userID string) error {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM chatbots
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to delete chatbot: %v", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return apperrors.Wrapf(apperrors.ErrChatbotNotFound,
			"chatbot with ID %s not found or not owned by user %s", id, userID)
	}

	return nil
}

// CheckChatbotOwnership verifies if a user owns a specific chatbot
func (s *ChatService) CheckChatbotOwnership(ctx context.Context, chatbotID uuid.UUID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM chatbots
			WHERE id = $1 AND user_id = $2
		)
	`, chatbotID, userID).Scan(&exists)

	if err != nil {
		return false, apperrors.Wrapf(apperrors.ErrDatabaseOperation,
			"failed to check chatbot ownership: %v", err)
	}

	return exists, nil
}

// GetChatbotByID retrieves a chatbot by its ID
func (s *ChatService) GetChatbotByID(ctx context.Context, id uuid.UUID) (*Chatbot, error) {
	query := `
		SELECT id, user_id, name, description, system_instructions,
		       model_name, temperature_param, max_tokens,
		       created_at, updated_at
		FROM chatbots
		WHERE id = $1
	`

	var chatbot Chatbot
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&chatbot.ID,
		&chatbot.UserID,
		&chatbot.Name,
		&chatbot.Description,
		&chatbot.SystemInstructions,
		&chatbot.ModelName,
		&chatbot.TemperatureParam,
		&chatbot.MaxTokens,
		&chatbot.CreatedAt,
		&chatbot.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrChatbotNotFound
		}
		return nil, apperrors.Wrap(err, "failed to get chatbot")
	}

	return &chatbot, nil
}

// Document database operations

// StoreDocument stores a document with its vector embedding
func (s *ChatService) StoreDocument(ctx context.Context, doc Document) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO documents (id, content, embedding, chatbot_id, file_id, chunk_index)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE
		SET content = $2, embedding = $3, chatbot_id = $4, file_id = $5, chunk_index = $6
	`, doc.ID, doc.Content, pgvector.NewVector(doc.Embedding), doc.ChatbotID, doc.FileID, doc.ChunkIndex)

	if err != nil {
		return apperrors.Wrapf(apperrors.ErrDatabaseOperation, "failed to store document: %v", err)
	}

	return nil
}

// FindSimilarDocuments finds documents similar to the given embedding
func (s *ChatService) FindSimilarDocuments(ctx context.Context, embedding []float32, limit int) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
		FROM documents
		ORDER BY embedding <=> $1
		LIMIT $2
	`, pgvector.NewVector(embedding), limit)

	if err != nil {
		return nil, fmt.Errorf("failed to query similar documents: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
			return nil, fmt.Errorf("failed to scan document row: %v", err)
		}

		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating document rows: %v", err)
	}

	return documents, nil
}

// FindSimilarDocumentsByChatbot finds documents similar to the given embedding that belong to a specific chatbot
func (s *ChatService) FindSimilarDocumentsByChatbot(ctx context.Context, embedding []float32, chatbotID string, limit int) ([]Document, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, content, embedding, chatbot_id, file_id, chunk_index
		FROM documents
		WHERE chatbot_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3
	`, chatbotID, pgvector.NewVector(embedding), limit)

	if err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation,
			"failed to query similar documents by chatbot: %v", err)
	}
	defer rows.Close()

	var documents []Document
	for rows.Next() {
		var doc Document
		var pgvec pgvector.Vector

		if err := rows.Scan(&doc.ID, &doc.Content, &pgvec, &doc.ChatbotID, &doc.FileID, &doc.ChunkIndex); err != nil {
			return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation,
				"failed to scan document row: %v", err)
		}

		doc.Embedding = pgvec.Slice()
		documents = append(documents, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrapf(apperrors.ErrDatabaseOperation,
			"error iterating document rows: %v", err)
	}

	return documents, nil
}

// DeleteDocument removes a document from the database
func (s *ChatService) DeleteDocument(ctx context.Context, id string) error {
	_, err := s.pool.Exec(ctx, `
		DELETE FROM documents
		WHERE id = $1
	`, id)

	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

// InsertFile inserts a new file record into the files table
func (s *ChatService) InsertFile(ctx context.Context, fileID uuid.UUID, chatbotID uuid.UUID, filename string) error {
	_, err := s.pool.Exec(ctx, `INSERT INTO files (id, chatbot_id, filename) VALUES ($1, $2, $3)`, fileID, chatbotID, filename)
	if err != nil {
		return apperrors.Wrap(err, "failed to insert file metadata")
	}
	return nil
}

// GetFilesByChatbotID retrieves all files for a given chatbot_id
func (s *ChatService) GetFilesByChatbotID(ctx context.Context, chatbotID uuid.UUID) ([]File, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, chatbot_id, filename, uploaded_at
		FROM files
		WHERE chatbot_id = $1
		ORDER BY uploaded_at DESC
	`, chatbotID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to query files by chatbot_id")
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var f File
		if err := rows.Scan(&f.ID, &f.ChatbotID, &f.Filename, &f.UploadedAt); err != nil {
			return nil, apperrors.Wrap(err, "failed to scan file row")
		}
		files = append(files, f)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, "error iterating file rows")
	}
	return files, nil
}
