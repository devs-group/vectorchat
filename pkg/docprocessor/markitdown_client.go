package docprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// MarkitdownClient wraps the HTTP calls to the MarkItDown service.
type MarkitdownClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMarkitdownClient creates a MarkItDown client with the provided base URL.
func NewMarkitdownClient(baseURL string) (*MarkitdownClient, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "MARKITDOWN_API_URL is required")
	}

	client := &http.Client{Timeout: 60 * time.Second}

	return &MarkitdownClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: client,
	}, nil
}

// Convert uploads the file bytes to the MarkItDown service and returns Markdown.
func (c *MarkitdownClient) Convert(ctx context.Context, filename string, data []byte) (string, error) {
	// If no data provided, return error
	if len(data) == 0 {
		return "", apperrors.Wrap(apperrors.ErrInvalidChatbotParameters, "file data is empty")
	}

	// Ensure filename is not empty - use a default if needed
	if strings.TrimSpace(filename) == "" {
		filename = "uploaded_file"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", apperrors.Wrapf(err, "failed to create multipart payload for file: %s", filename)
	}
	if _, err := part.Write(data); err != nil {
		return "", apperrors.Wrapf(err, "failed to write file payload for file: %s (size: %d bytes)", filename, len(data))
	}
	if err := writer.Close(); err != nil {
		return "", apperrors.Wrap(err, "failed to finalize multipart payload")
	}

	endpoint := fmt.Sprintf("%s/convert", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &buf)
	if err != nil {
		return "", apperrors.Wrapf(err, "failed to create markitdown request to %s for file: %s", endpoint, filename)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", apperrors.Wrapf(err, "failed to call markitdown convert endpoint %s for file: %s", endpoint, filename)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", apperrors.Wrapf(err, "failed to read markitdown response for file: %s", filename)
	}

	if resp.StatusCode != http.StatusOK {
		message := decodeErrorMessage(body, resp.Status)
		return "", apperrors.Wrapf(apperrors.ErrInvalidChatbotParameters, "markitdown conversion failed for file %s (status %d): %s", filename, resp.StatusCode, message)
	}

	result := strings.TrimSpace(string(body))
	if result == "" {
		return "", apperrors.Wrapf(apperrors.ErrInvalidChatbotParameters, "markitdown service returned empty content for file: %s", filename)
	}

	return result, nil
}

// SupportedExtensions fetches the list of supported file extensions from the MarkItDown service.
func (c *MarkitdownClient) SupportedExtensions(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("%s/supported-extensions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, apperrors.Wrapf(err, "failed to create request to %s", endpoint)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrapf(err, "failed to call supported-extensions endpoint %s", endpoint)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrapf(err, "failed to read supported-extensions response")
	}

	if resp.StatusCode != http.StatusOK {
		message := decodeErrorMessage(body, resp.Status)
		return nil, apperrors.Wrapf(apperrors.ErrInvalidChatbotParameters, "supported-extensions request failed (status %d): %s", resp.StatusCode, message)
	}

	var response struct {
		Extensions []string `json:"extensions"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, apperrors.Wrapf(err, "failed to parse supported-extensions response")
	}

	return response.Extensions, nil
}

func normalizeExtension(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return ext
}

func decodeErrorMessage(body []byte, fallback string) string {
	if len(body) == 0 {
		return fallback
	}

	var payload struct {
		Detail  any    `json:"detail"`
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &payload); err == nil {
		switch v := payload.Detail.(type) {
		case string:
			if strings.TrimSpace(v) != "" {
				return strings.TrimSpace(v)
			}
		case []any:
			// FastAPI details may include a list of error objects.
			if len(v) > 0 {
				if m, ok := v[0].(map[string]any); ok {
					if msg, ok := m["msg"].(string); ok && strings.TrimSpace(msg) != "" {
						return strings.TrimSpace(msg)
					}
				}
			}
		}
		if strings.TrimSpace(payload.Message) != "" {
			return strings.TrimSpace(payload.Message)
		}
		if strings.TrimSpace(payload.Error) != "" {
			return strings.TrimSpace(payload.Error)
		}
	}

	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fallback
	}
	return msg
}
