package models

// TokenRequest represents the payload for exchanging client credentials for an access token.
type TokenRequest struct {
	ClientID     string `json:"client_id" example:"public-vectorchat-client"`
	ClientSecret string `json:"client_secret" example:"super-secret-value"`
}

// TokenResponse represents the standard OAuth2 token response returned by Hydra.
type TokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresIn   int    `json:"expires_in" example:"3599"`
	TokenType   string `json:"token_type" example:"bearer"`
	Scope       string `json:"scope,omitempty" example:"chat:read chat:write"`
}
