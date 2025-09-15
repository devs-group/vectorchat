package vectorize

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/ledongthuc/pdf"
    "github.com/tmc/langchaingo/embeddings"
)

// Vectorizer is an interface for creating vector embeddings from text
type Vectorizer interface {
	VectorizeText(ctx context.Context, text string) ([]float32, error)
	VectorizeFile(ctx context.Context, filePath string) ([]float32, error)
}

// OpenAIVectorizer implements Vectorizer using OpenAI's embeddings
type OpenAIVectorizer struct {
	client embeddings.Embedder
}

// NewOpenAIVectorizer creates a new OpenAI vectorizer
func NewOpenAIVectorizer(apiKey string) *OpenAIVectorizer {
    // Create a custom OpenAI embedder since the function isn't available
    client := &openAIEmbedder{
        apiKey: apiKey,
        model:  "text-embedding-3-small",
        http:   &http.Client{Timeout: 60 * time.Second},
    }

	return &OpenAIVectorizer{
		client: client,
	}
}

// openAIEmbedder implements the embeddings.Embedder interface
type openAIEmbedder struct {
    apiKey string
    model  string
    http   *http.Client
}

// EmbedDocuments implements the embeddings.Embedder interface
func (e *openAIEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float64, error) {
    if len(texts) == 0 {
        return [][]float64{}, nil
    }

    type embedReq struct {
        Model string   `json:"model"`
        Input []string `json:"input"`
    }
    type embedRes struct {
        Data []struct {
            Embedding []float64 `json:"embedding"`
        } `json:"data"`
        Error *struct {
            Message string `json:"message"`
        } `json:"error"`
    }

    const endpoint = "https://api.openai.com/v1/embeddings"
    const batchSize = 100
    out := make([][]float64, 0, len(texts))

    for start := 0; start < len(texts); start += batchSize {
        end := start + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        payload := embedReq{Model: e.model, Input: texts[start:end]}
        body, err := json.Marshal(payload)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal embeddings request: %w", err)
        }

        req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
        if err != nil {
            return nil, fmt.Errorf("failed to create embeddings request: %w", err)
        }
        req.Header.Set("Authorization", "Bearer "+e.apiKey)
        req.Header.Set("Content-Type", "application/json")

        client := e.http
        if client == nil {
            client = http.DefaultClient
        }
        resp, err := client.Do(req)
        if err != nil {
            return nil, fmt.Errorf("embeddings request error: %w", err)
        }
        respBytes, err := io.ReadAll(resp.Body)
        resp.Body.Close()
        if err != nil {
            return nil, fmt.Errorf("failed to read embeddings response: %w", err)
        }
        if resp.StatusCode < 200 || resp.StatusCode >= 300 {
            return nil, fmt.Errorf("embeddings API error: status %d, body: %s", resp.StatusCode, string(respBytes))
        }
        var parsed embedRes
        if err := json.Unmarshal(respBytes, &parsed); err != nil {
            return nil, fmt.Errorf("failed to parse embeddings response: %w", err)
        }
        if parsed.Error != nil {
            return nil, fmt.Errorf("embeddings API error: %s", parsed.Error.Message)
        }
        if len(parsed.Data) != end-start {
            return nil, fmt.Errorf("embeddings count mismatch: got %d, want %d", len(parsed.Data), end-start)
        }
        for _, d := range parsed.Data {
            out = append(out, d.Embedding)
        }
    }

    return out, nil
}

// EmbedQuery implements the embeddings.Embedder interface
func (e *openAIEmbedder) EmbedQuery(ctx context.Context, text string) ([]float64, error) {
	embeddings, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return embeddings[0], nil
}

// VectorizeText creates a vector embedding from text
func (v *OpenAIVectorizer) VectorizeText(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := v.client.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %v", err)
	}

	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	// Convert float64 embeddings to float32 for pgvector compatibility
	float32Embeddings := make([]float32, len(embeddings[0]))
	for i, v := range embeddings[0] {
		float32Embeddings[i] = float32(v)
	}
	return float32Embeddings, nil
}

// ExtractTextFromPDF extracts all text from a PDF file
func ExtractTextFromPDF(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	reader, err := pdf.NewReader(file, fileStatSize(file))
	if err != nil {
		return "", err
	}
	var textBuilder strings.Builder
	b, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	_, err = io.Copy(&textBuilder, b)
	if err != nil {
		return "", err
	}
	return textBuilder.String(), nil
}

// fileStatSize returns the size of the file for pdf.NewReader
func fileStatSize(f *os.File) int64 {
	fi, err := f.Stat()
	if err != nil {
		return 0
	}
	return fi.Size()
}

// VectorizeFile reads a file and creates a vector embedding from its content
func (v *OpenAIVectorizer) VectorizeFile(ctx context.Context, filePath string) ([]float32, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Simple content type detection based on file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	var text string

	switch ext {
	case ".pdf":
		// For PDFs, extract text and let the service layer handle chunking/embedding
		text, err = ExtractTextFromPDF(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract text from PDF: %v", err)
		}
		return nil, fmt.Errorf("PDF extraction handled in service layer; use VectorizeText for chunks")
	case ".txt", ".md", ".go", ".py", ".js", ".html", ".css", ".json":
		// Text files can be processed directly
		text = string(content)
	default:
		// For other file types, just use the raw content
		text = string(content)
	}

	return v.VectorizeText(ctx, text)
}
