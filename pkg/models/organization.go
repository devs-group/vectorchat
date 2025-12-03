package models

import (
	"time"

	"github.com/google/uuid"
)

type OrganizationCreateRequest struct {
	Name         string  `json:"name" example:"Acme Inc"`
	Slug         *string `json:"slug,omitempty" example:"acme"`
	Description  *string `json:"description,omitempty" example:"Team workspace for Acme"`
	BillingEmail *string `json:"billing_email,omitempty" example:"billing@acme.com"`
}

type OrganizationUpdateRequest struct {
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Description  *string `json:"description,omitempty"`
	BillingEmail *string `json:"billing_email,omitempty"`
	PlanTier     *string `json:"plan_tier,omitempty"`
}

type OrganizationResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	BillingEmail *string   `json:"billing_email,omitempty"`
	PlanTier     string    `json:"plan_tier"`
	Role         string    `json:"role"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
}

type OrganizationMemberResponse struct {
	UserID       string     `json:"user_id"`
	Role         string     `json:"role"`
	JoinedAt     time.Time  `json:"joined_at"`
	LastActiveAt *time.Time `json:"last_active_at,omitempty"`
	InvitedBy    *string    `json:"invited_by,omitempty"`
}

type OrganizationInviteRequest struct {
	Email   string  `json:"email"`
	Role    string  `json:"role"`
	Message *string `json:"message,omitempty"`
}

type OrganizationInviteResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Email          string     `json:"email"`
	Role           string     `json:"role"`
	ExpiresAt      time.Time  `json:"expires_at"`
	AcceptedAt     *time.Time `json:"accepted_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type OrganizationSwitchRequest struct {
	OrganizationID *uuid.UUID `json:"organization_id"` // nil means personal workspace
}
