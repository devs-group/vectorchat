package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/vectorchat/pkg/chat"
	"github.com/yourusername/vectorchat/pkg/db"
	"github.com/yourusername/vectorchat/pkg/vectorize"
)

func TestIntegration(t *testing.T) {
	// Skip if not explicitly testing integration
	if os.Getenv("TEST_INTEGRATION") != "true" {
		t.Skip("Skipping integration test. Set TEST_INTEGRATION=true to run")
	}

	// Get environment variables
	pgConnStr := os.Getenv("PG_CONNECTION_STRING")
	if pgConnStr == "" {
		pgConnStr = "postgres://postgres:postgres@localhost:5432/vectordb_test?sslmode=disable"
	}

	openaiKey := os.Getenv("OPENAI_API_KEY")
	if openaiKey == "" {
		t.Skip("OPENAI_API_KEY environment variable is required for integration test")
	}

	// Initialize components
	database, err := db.NewPgVectorDB(pgConnStr)
	require.NoError(t, err)
	defer database.Close()

	vectorizer := vectorize.NewOpenAIVectorizer(openaiKey)
	chatService := chat.NewChatService(database, vectorizer, openaiKey)

	ctx := context.Background()

	// Add test documents
	docs := []struct {
		id      string
		content string
	}{
		{"golang", "Go is a statically typed, compiled programming language designed at Google."},
		{"python", "Python is an interpreted, high-level, general-purpose programming language."},
		{"javascript", "JavaScript is a high-level, interpreted programming language that conforms to the ECMAScript specification."},
	}

	for _, doc := range docs {
		err = chatService.AddDocument(ctx, doc.id, doc.content)
		require.NoError(t, err)
	}

	// Test chat with context
	testCases := []struct {
		query       string
		shouldMatch string
	}{
		{
			query:       "Which language was designed at Google?",
			shouldMatch: "Go",
		},
		{
			query:       "Which language is interpreted?",
			shouldMatch: "Python",
		},
		{
			query:       "Tell me about ECMAScript",
			shouldMatch: "JavaScript",
		},
	}

	for _, tc := range testCases {
		response, err := chatService.Chat(ctx, tc.query)
		require.NoError(t, err)
		assert.Contains(t, response, tc.shouldMatch)
	}
} 