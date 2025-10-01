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
	httpClient *http.Client
}

// NewHydraService creates a new Hydra service using the provided admin URL.
func NewHydraService(adminURL string) *HydraService {
	return &HydraService{
		adminURL:   strings.TrimRight(adminURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
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
		"token_endpoint_auth_method": "client_secret_post",
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
