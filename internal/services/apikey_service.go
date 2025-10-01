package services

import (
	"context"
	"sort"
	"time"

	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/pkg/models"
)

// APIKeyService now acts as a thin adapter over Hydra's OAuth client APIs.
type APIKeyService struct {
	*CommonService
	hydra *HydraService
}

// NewAPIKeyService creates a new API key service instance backed by Hydra.
func NewAPIKeyService(hydra *HydraService) *APIKeyService {
	return &APIKeyService{
		CommonService: NewCommonService(),
		hydra:         hydra,
	}
}

// ParseAPIKeyRequest parses and validates an API key creation request.
func (s *APIKeyService) ParseAPIKeyRequest(req *models.APIKeyCreateRequest) (string, *time.Time, error) {
	name := req.Name
	var expiresAt *time.Time

	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err != nil {
			return "", nil, apperrors.Wrap(err, "invalid expiration date format")
		}
		expiresAt = &parsedTime
	}

	return name, expiresAt, nil
}

// CreateAPIKey provisions a new OAuth client through Hydra and returns its credentials.
func (s *APIKeyService) CreateAPIKey(ctx context.Context, userID, name string, expiresAt *time.Time) (*models.APIKeyResponse, string, error) {
	client, err := s.hydra.CreateMachineToMachineClient(ctx, name, userID, expiresAt)
	if err != nil {
		return nil, "", apperrors.Wrap(err, "failed to create oauth client")
	}

	response := toAPIKeyResponse(client)
	return response, client.ClientSecret, nil
}

// GetAPIKeysWithPagination lists OAuth clients for the user using in-memory pagination.
func (s *APIKeyService) GetAPIKeysWithPagination(ctx context.Context, userID string, page, limit, offset int) (*models.APIKeysListResponse, error) {
	clients, err := s.hydra.ListClientsForUser(ctx, userID)
	if err != nil {
		return nil, apperrors.Wrap(err, "failed to list oauth clients")
	}

	responses := make([]*models.APIKeyResponse, 0, len(clients))
	for i := range clients {
		client := clients[i]
		res := toAPIKeyResponse(&client)
		responses = append(responses, res)
	}

	sort.Slice(responses, func(i, j int) bool {
		return responses[i].CreatedAt.After(responses[j].CreatedAt)
	})

	total := len(responses)
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	if offset < 0 {
		offset = 0
	}

	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	sliced := responses[start:end]

	totalPages := total / limit
	if total%limit != 0 {
		totalPages++
	}

	return &models.APIKeysListResponse{
		APIKeys: sliced,
		Pagination: &models.PaginationMetadata{
			Page:       page,
			Limit:      limit,
			Total:      int64(total),
			TotalPages: totalPages,
			HasNext:    page < totalPages,
			HasPrev:    page > 1,
		},
	}, nil
}

// RevokeAPIKey removes an OAuth client owned by the user.
func (s *APIKeyService) RevokeAPIKey(ctx context.Context, clientID string, userID string) error {
	if clientID == "" {
		return apperrors.Wrap(apperrors.ErrAPIKeyNotFound, "api key id is required")
	}

	clients, err := s.hydra.ListClientsForUser(ctx, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to verify client ownership")
	}

	allowed := false
	for _, client := range clients {
		if client.ClientID == clientID {
			allowed = true
			break
		}
	}
	if !allowed {
		return apperrors.ErrAPIKeyNotFound
	}

	if err := s.hydra.DeleteClient(ctx, clientID); err != nil {
		return apperrors.Wrap(err, "failed to revoke oauth client")
	}

	return nil
}

func toAPIKeyResponse(client *HydraOAuthClient) *models.APIKeyResponse {
	var createdAt time.Time
	if client.Metadata != nil && !client.Metadata.CreatedAt.IsZero() {
		createdAt = client.Metadata.CreatedAt
	} else if client.CreatedAt != nil && !client.CreatedAt.IsZero() {
		createdAt = client.CreatedAt.UTC()
	} else {
		createdAt = time.Now().UTC()
	}

	var expiresAt *time.Time
	if client.Metadata != nil && client.Metadata.ExpiresAt != nil {
		expiresAt = client.Metadata.ExpiresAt
	}

	var name *string
	if client.Metadata != nil && client.Metadata.Name != "" {
		copyName := client.Metadata.Name
		name = &copyName
	} else if client.ClientName != "" {
		copyName := client.ClientName
		name = &copyName
	}

	userID := client.Owner
	if userID == "" && client.Metadata != nil {
		userID = client.Metadata.UserID
	}

	return &models.APIKeyResponse{
		ID:        client.ClientID,
		ClientID:  client.ClientID,
		UserID:    userID,
		Name:      name,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
	}
}
