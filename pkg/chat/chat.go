package chat

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/prompts"
	"github.com/yourusername/vectorchat/pkg/db"
	apperrors "github.com/yourusername/vectorchat/pkg/errors"
	"github.com/yourusername/vectorchat/pkg/vectorize"
)

// ChatService handles chat interactions with context from vector database
type ChatService struct {
	db         db.VectorDB
	vectorizer vectorize.Vectorizer
	llm        llms.LLM
}

// NewChatService creates a new chat service
func NewChatService(database *db.PgVectorDB, vectorizer vectorize.Vectorizer, openaiKey string) *ChatService {
	llm, err := openai.New(
		openai.WithToken(openaiKey),
		openai.WithModel("gpt-3.5-turbo"),
	)
	
	if err != nil {
		panic(fmt.Sprintf("Failed to create OpenAI LLM: %v", err))
	}

	return &ChatService{
		db:         database,
		vectorizer: vectorizer,
		llm:        llm,
	}
}

// AddDocument adds a document to the vector database
func (c *ChatService) AddDocument(ctx context.Context, id string, content string) error {
	embedding, err := c.vectorizer.VectorizeText(ctx, content)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize document")
	}

	doc := db.Document{
		ID:        id,
		Content:   content,
		Embedding: embedding,
	}

	return c.db.StoreDocument(ctx, doc)
}

// AddFile adds a file to the vector database
func (c *ChatService) AddFile(ctx context.Context, id string, filePath string) error {
	embedding, err := c.vectorizer.VectorizeFile(ctx, filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to vectorize file")
	}

	// Read file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return apperrors.Wrap(err, "failed to read file")
	}

	doc := db.Document{
		ID:        id,
		Content:   string(content),
		Embedding: embedding,
	}

	return c.db.StoreDocument(ctx, doc)
}

// Chat sends a message to the LLM with relevant context from the vector database
func (c *ChatService) Chat(ctx context.Context, message string) (string, error) {
	return c.ChatWithID(ctx, "default", message)
}

// ChatWithID sends a message to the LLM with relevant context from the specified chat session
func (c *ChatService) ChatWithID(ctx context.Context, chatID string, message string) (string, error) {
	// Vectorize the query
	queryEmbedding, err := c.vectorizer.VectorizeText(ctx, message)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrVectorizationFailed, "query: %v", err)
	}

	// Find relevant documents
	docs, err := c.db.FindSimilarDocumentsByChatID(ctx, queryEmbedding, chatID, 3)
	if err != nil {
		return "", apperrors.Wrapf(apperrors.ErrDatabaseOperation, "find similar documents: %v", err)
	}
	
	// Check if any documents were found for this chat ID
	if len(docs) == 0 {
		return "", apperrors.Wrapf(apperrors.ErrNoDocumentsFound, "chat ID: %s", chatID)
	}

	// Build context from documents
	var contextBuilder strings.Builder
	for i, doc := range docs {
		contextBuilder.WriteString(fmt.Sprintf("Document %d:\n%s\n\n", i+1, doc.Content))
	}
	context := contextBuilder.String()

	// Create a prompt template
	promptTemplate := prompts.NewPromptTemplate(
		"Context information is below.\n"+
			"---------------------\n"+
			"{{.context}}\n"+
			"---------------------\n"+
			"Given the context information and not prior knowledge, answer the query.\n"+
			"Query: {{.query}}\n"+
			"Answer: ",
		[]string{"context", "query"},
	)

	// Use the LLM directly since chains.NewLLMChain expects a different interface
	prompt, err := promptTemplate.Format(map[string]any{
		"context": context,
		"query":   message,
	})
	if err != nil {
		return "", apperrors.Wrap(err, "failed to format prompt")
	}

	// Generate response using the LLM
	completion, err := c.llm.Call(ctx, prompt, llms.WithMaxTokens(1000))
	if err != nil {
		return "", apperrors.Wrap(err, "failed to generate completion")
	}

	return completion, nil
} 