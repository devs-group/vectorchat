package crawler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultUserAgent = "VectorChatBot/1.0 (+https://vectorchat.local)"

// WebCrawler defines the minimal behavior expected from a website crawler implementation.
type WebCrawler interface {
	Crawl(ctx context.Context, root string, opts Options) ([]Page, error)
}

// APIClient wraps the crawl4ai HTTP API.
type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewAPIClient validates the provided base URL and constructs a crawl4ai API client.
func NewAPIClient(baseURL string, httpClient *http.Client) (*APIClient, error) {
	trimmed := strings.TrimSpace(baseURL)
	if trimmed == "" {
		return nil, errors.New("crawler: base URL is required")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return nil, fmt.Errorf("crawler: invalid crawl4ai base URL: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("crawler: crawl4ai base URL must include scheme and host, got %q", baseURL)
	}

	client := httpClient
	if client == nil {
		client = &http.Client{}
	}

	return &APIClient{
		baseURL:    strings.TrimRight(parsed.String(), "/"),
		httpClient: client,
	}, nil
}

// Crawl requests markdown content for the provided root URL via the crawl4ai API.
func (c *APIClient) Crawl(ctx context.Context, root string, opts Options) ([]Page, error) {
	if c == nil {
		return nil, errors.New("crawler: API client is nil")
	}

	target := strings.TrimSpace(root)
	if target == "" {
		return nil, errors.New("crawler: root URL is required")
	}

	payload := map[string]any{
		"url":    target,
		"urls":   []string{target},
		"output": "markdown",
	}

	payload["browser_config"] = map[string]any{
		"type": "BrowserConfig",
		"params": map[string]any{
			"headless":   true,
			"text_mode":  true,
			"light_mode": true,
			"verbose":    false,
		},
	}

	payload["crawler_config"] = map[string]any{
		"type": "CrawlerRunConfig",
		"params": map[string]any{
			"stream":                 true,
			"cache_mode":             "ENABLED",
			"screenshot":             false,
			"pdf":                    false,
			"capture_mhtml":          false,
			"excluded_tags":          []string{"script", "style", "noscript"},
			"exclude_external_links": true,
		},
	}

	if opts.MaxDepth > 0 {
		payload["max_depth"] = opts.MaxDepth
	}
	if opts.MaxPages > 0 {
		payload["max_pages"] = opts.MaxPages
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("crawler: failed to encode crawl request: %w", err)
	}

	reqCtx := ctx
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		reqCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
	} else {
		reqCtx, cancel = context.WithTimeout(ctx, 45*time.Second)
	}
	defer cancel()

	endpoint := c.baseURL + "/crawl"
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("crawler: failed to create crawl request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", defaultUserAgent)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("crawler: crawl4ai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("crawler: failed to read crawl4ai response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		snippet := string(respBody)
		if len(snippet) > 256 {
			snippet = snippet[:256]
		}
		return nil, fmt.Errorf("crawler: crawl4ai returned status %d: %s", resp.StatusCode, strings.TrimSpace(snippet))
	}

	pages, err := parseCrawl4AIResponse(target, respBody)
	if err != nil {
		return nil, err
	}
	return pages, nil
}

// crawl4AIResponse mirrors multiple possible shapes returned by the crawl4ai API.
type crawl4AIResponse struct {
	Status   string         `json:"status"`
	Message  string         `json:"message"`
	Error    string         `json:"error"`
	Title    string         `json:"title"`
	Markdown string         `json:"markdown"`
	Result   *crawl4AIItem  `json:"result"`
	Results  []crawl4AIItem `json:"results"`
	Data     []crawl4AIItem `json:"data"`
	Pages    []crawl4AIItem `json:"pages"`
	Items    []crawl4AIItem `json:"items"`
}

type crawl4AIItem struct {
	URL             json.RawMessage `json:"url"`
	Title           json.RawMessage `json:"title"`
	Markdown        json.RawMessage `json:"markdown"`
	ContentMarkdown json.RawMessage `json:"content_markdown"`
	CleanMarkdown   json.RawMessage `json:"clean_markdown"`
	Content         json.RawMessage `json:"content"`
	Text            json.RawMessage `json:"text"`
	Body            json.RawMessage `json:"body"`
	RawContent      json.RawMessage `json:"raw"`
	Summary         json.RawMessage `json:"summary"`
}

func (i crawl4AIItem) primaryText() string {
	candidates := []json.RawMessage{
		i.Markdown,
		i.ContentMarkdown,
		i.CleanMarkdown,
		i.Content,
		i.Text,
		i.Body,
		i.RawContent,
		i.Summary,
	}
	for _, raw := range candidates {
		if s := strings.TrimSpace(rawToString(raw)); s != "" {
			return s
		}
	}
	return ""
}

func parseCrawl4AIResponse(root string, data []byte) ([]Page, error) {
	var resp crawl4AIResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		// Try to parse as an array of items directly
		var direct []crawl4AIItem
		if err2 := json.Unmarshal(data, &direct); err2 == nil {
			return itemsToPages(root, direct, ""), nil
		}
		// Try simple markdown string
		var markdown string
		if err3 := json.Unmarshal(data, &markdown); err3 == nil && strings.TrimSpace(markdown) != "" {
			return []Page{{URL: root, Title: "", Text: markdown}}, nil
		}
		return nil, fmt.Errorf("crawler: invalid crawl4ai response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("crawler: crawl4ai error: %s", strings.TrimSpace(resp.Error))
	}
	if strings.EqualFold(resp.Status, "error") && strings.TrimSpace(resp.Message) != "" {
		return nil, fmt.Errorf("crawler: crawl4ai error: %s", strings.TrimSpace(resp.Message))
	}

	combined := make([]crawl4AIItem, 0, 8)
	if resp.Result != nil {
		combined = append(combined, *resp.Result)
	}
	combined = append(combined, resp.Results...)
	combined = append(combined, resp.Data...)
	combined = append(combined, resp.Pages...)
	combined = append(combined, resp.Items...)

	pages := itemsToPages(root, combined, resp.Title)
	if len(pages) > 0 {
		return pages, nil
	}

	if strings.TrimSpace(resp.Markdown) != "" {
		return []Page{{URL: root, Title: strings.TrimSpace(resp.Title), Text: resp.Markdown}}, nil
	}

	return nil, errors.New("crawler: crawl4ai response did not include markdown content")
}

func itemsToPages(root string, items []crawl4AIItem, fallbackTitle string) []Page {
	if len(items) == 0 {
		return nil
	}

	pages := make([]Page, 0, len(items))
	seen := make(map[string]struct{})
	for _, item := range items {
		text := strings.TrimSpace(item.primaryText())
		if text == "" {
			continue
		}
		url := strings.TrimSpace(rawToString(item.URL))
		if url == "" {
			url = root
		}
		title := strings.TrimSpace(rawToString(item.Title))
		if title == "" {
			title = fallbackTitle
		}

		key := url + "|" + title
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		pages = append(pages, Page{URL: url, Title: title, Text: text})
	}
	return pages
}

func rawToString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return strings.Join(arr, "\n\n")
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err == nil {
		return anyToString(generic)
	}
	return ""
}

func anyToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	case float64, float32, int, int64, int32, uint64, uint32, uint16, uint8, int16, int8, uint:
		return fmt.Sprint(v)
	case bool:
		return fmt.Sprint(v)
	case []byte:
		return string(val)
	case []any:
		parts := make([]string, 0, len(val))
		for _, item := range val {
			if s := strings.TrimSpace(anyToString(item)); s != "" {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, "\n\n")
	case map[string]any:
		for _, key := range []string{"markdown", "clean_markdown", "content_markdown", "content", "text", "body", "raw", "summary", "value"} {
			if v, ok := val[key]; ok {
				if s := strings.TrimSpace(anyToString(v)); s != "" {
					return s
				}
			}
		}
		for _, nested := range val {
			if s := strings.TrimSpace(anyToString(nested)); s != "" {
				return s
			}
		}
	}
	return ""
}
