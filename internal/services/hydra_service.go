package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

// HydraOAuthClientMetadata captures the metadata we attach to OAuth clients.
type HydraOAuthClientMetadata struct {
	UserID    string     `json:"user_id"`
	Name      string     `json:"name,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// HydraOAuthClient represents the subset of Hydra's OAuth client payload we care about.
type HydraOAuthClient struct {
	ClientID     string                    `json:"client_id"`
	ClientSecret string                    `json:"client_secret,omitempty"`
	ClientName   string                    `json:"client_name,omitempty"`
	Owner        string                    `json:"owner,omitempty"`
	GrantTypes   []string                  `json:"grant_types,omitempty"`
	Metadata     *HydraOAuthClientMetadata `json:"metadata,omitempty"`
	CreatedAt    *time.Time                `json:"created_at,omitempty"`
	UpdatedAt    *time.Time                `json:"updated_at,omitempty"`
}

// HydraService wraps calls to the Hydra admin APIs.
type HydraService struct {
	adminURL   string
	publicURL  string
	httpClient *http.Client
}

// NewHydraService creates a new Hydra service using the provided admin URL.
func NewHydraService(adminURL, publicURL string) *HydraService {
	return &HydraService{
		adminURL:   strings.TrimRight(adminURL, "/"),
		publicURL:  strings.TrimRight(publicURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// HydraTokenResponse represents the subset of fields returned by Hydra's token endpoint.
type HydraTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope,omitempty"`
}

// CreateMachineToMachineClient provisions a new client_credentials OAuth client tied to a user.
func (s *HydraService) CreateMachineToMachineClient(ctx context.Context, name, userID string, expiresAt *time.Time) (*HydraOAuthClient, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidUserData, "user id is required to create oauth client")
	}

	metadata := &HydraOAuthClientMetadata{
		UserID:    userID,
		Name:      name,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: expiresAt,
	}

	body := map[string]any{
		"client_name":                name,
		"grant_types":                []string{"client_credentials"},
		"token_endpoint_auth_method": "client_secret_basic",
		"owner":                      userID,
		"metadata":                   metadata,
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to marshal hydra client request")
	}

	endpoint := fmt.Sprintf("%s/admin/clients", s.adminURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(buf))
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create hydra client request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to execute hydra create client request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return nil, apperrors.Wrap(fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data))), "failed to create oauth client")
	}

	var client HydraOAuthClient
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode hydra client response")
	}

	return &client, nil
}

// ListClientsForUser returns all Hydra OAuth clients owned by the provided user.
func (s *HydraService) ListClientsForUser(ctx context.Context, userID string) ([]HydraOAuthClient, error) {
	if userID == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidUserData, "user id is required to list oauth clients")
	}

	endpoint := fmt.Sprintf("%s/admin/clients?owner=%s&limit=500", s.adminURL, url.QueryEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create hydra list clients request")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to execute hydra list clients request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return nil, apperrors.Wrap(fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data))), "failed to list oauth clients")
	}

	var clients []HydraOAuthClient
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode hydra list clients response")
	}

	return clients, nil
}

// DeleteClient removes an OAuth client by ID.
func (s *HydraService) DeleteClient(ctx context.Context, clientID string) error {
	if clientID == "" {
		return apperrors.Wrap(apperrors.ErrAPIKeyNotFound, "client id is required to revoke oauth client")
	}

	endpoint := fmt.Sprintf("%s/admin/clients/%s", s.adminURL, url.PathEscape(clientID))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return apperrors.Wrap(err, "failed to create hydra delete client request")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return apperrors.Wrap(err, "failed to execute hydra delete client request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apperrors.ErrAPIKeyNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return apperrors.Wrap(fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data))), "failed to delete oauth client")
	}

	return nil
}

// GetClient returns details for a single OAuth client.
func (s *HydraService) GetClient(ctx context.Context, clientID string) (*HydraOAuthClient, error) {
	if clientID == "" {
		return nil, apperrors.Wrap(apperrors.ErrAPIKeyNotFound, "client id is required to fetch oauth client")
	}

	endpoint := fmt.Sprintf("%s/admin/clients/%s", s.adminURL, url.PathEscape(clientID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create hydra get client request")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to execute hydra get client request")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperrors.ErrAPIKeyNotFound
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return nil, apperrors.Wrap(fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(data))), "failed to fetch oauth client")
	}

	var client HydraOAuthClient
	if err := json.NewDecoder(resp.Body).Decode(&client); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode hydra get client response")
	}

	return &client, nil
}

// ExchangeClientCredentials requests an access token from Hydra's public endpoint using the client credentials flow.
func (s *HydraService) ExchangeClientCredentials(ctx context.Context, clientID, clientSecret string) (*HydraTokenResponse, error) {
	if clientID == "" || clientSecret == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidAPIKey, "client credentials are required")
	}
	if s.publicURL == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidAPIKey, "hydra public url is not configured")
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")

	endpoint := fmt.Sprintf("%s/oauth2/token", s.publicURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to create hydra token request")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to execute hydra token request")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to read hydra token response")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var tokenErr map[string]any
		if decodeErr := json.Unmarshal(body, &tokenErr); decodeErr == nil {
			if desc, ok := tokenErr["error_description"].(string); ok && desc != "" {
				return nil, apperrors.Wrap(apperrors.ErrInvalidAPIKey, desc)
			}
			if errCode, ok := tokenErr["error"].(string); ok && errCode != "" {
				return nil, apperrors.Wrap(apperrors.ErrInvalidAPIKey, errCode)
			}
		}
		return nil, apperrors.Wrap(fmt.Errorf("status %d: %s", resp.StatusCode, strings.TrimSpace(string(body))), "hydra token request failed")
	}

	var token HydraTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, apperrors.Wrap(err, "failed to decode hydra token response")
	}

	return &token, nil
}
