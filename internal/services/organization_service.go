package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/vectorchat/internal/db"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
	"github.com/yourusername/vectorchat/pkg/models"
)

const (
	OrgRoleOwner    = "owner"
	OrgRoleAdmin    = "admin"
	OrgRoleMember   = "member"
	OrgRoleBilling  = "billing"
	OrgRolePersonal = "personal"
)

var allowedRoles = map[string]struct{}{
	OrgRoleOwner:   {},
	OrgRoleAdmin:   {},
	OrgRoleMember:  {},
	OrgRoleBilling: {},
}

type OrganizationContext struct {
	ID   *uuid.UUID
	Role string
}

func (o *OrganizationContext) IsPersonal() bool {
	return o == nil || o.ID == nil
}

func (o *OrganizationContext) HasRole(roles ...string) bool {
	if o == nil {
		return false
	}
	for _, r := range roles {
		if o.Role == r {
			return true
		}
	}
	return false
}

type OrganizationService struct {
	orgRepo    *db.OrganizationRepository
	memberRepo *db.OrganizationMemberRepository
	inviteRepo *db.OrganizationInviteRepository
	userRepo   *db.UserRepository
}

func NewOrganizationService(
	orgRepo *db.OrganizationRepository,
	memberRepo *db.OrganizationMemberRepository,
	inviteRepo *db.OrganizationInviteRepository,
	userRepo *db.UserRepository,
) *OrganizationService {
	return &OrganizationService{
		orgRepo:    orgRepo,
		memberRepo: memberRepo,
		inviteRepo: inviteRepo,
		userRepo:   userRepo,
	}
}

func (s *OrganizationService) Create(ctx context.Context, userID string, req *models.OrganizationCreateRequest) (*models.OrganizationResponse, error) {
	if req == nil || strings.TrimSpace(req.Name) == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidUserData, "name is required")
	}

	slug := strings.TrimSpace(strings.ToLower(coalesce(req.Slug, slugify(req.Name))))
	if slug == "" {
		return nil, apperrors.Wrap(apperrors.ErrInvalidUserData, "slug is required")
	}

	org := &db.Organization{
		Name:        strings.TrimSpace(req.Name),
		Slug:        slug,
		Description: req.Description,
		CreatedBy:   userID,
		PlanTier:    "free",
	}
	if req.BillingEmail != nil && strings.TrimSpace(*req.BillingEmail) != "" {
		org.BillingEmail = req.BillingEmail
	}

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	member := &db.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         userID,
		Role:           OrgRoleOwner,
		JoinedAt:       time.Now().UTC(),
	}
	if err := s.memberRepo.Upsert(ctx, member); err != nil {
		return nil, err
	}

	return toOrgResponse(org, OrgRoleOwner), nil
}

func (s *OrganizationService) ListForUser(ctx context.Context, userID string) (*models.OrganizationListResponse, error) {
	orgs, err := s.orgRepo.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]models.OrganizationResponse, 0, len(orgs)+1)
	// Personal workspace sentinel
	res = append(res, models.OrganizationResponse{
		ID:        uuid.Nil,
		Name:      "Personal",
		Slug:      "personal",
		PlanTier:  "free",
		Role:      OrgRolePersonal,
		CreatedBy: userID,
	})

	for _, o := range orgs {
		res = append(res, *toOrgResponse(&o.Organization, o.Role))
	}

	return &models.OrganizationListResponse{Organizations: res}, nil
}

func (s *OrganizationService) Get(ctx context.Context, orgID uuid.UUID, userID string) (*models.OrganizationResponse, error) {
	org, role, err := s.loadOrgWithRole(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	return toOrgResponse(org, role), nil
}

func (s *OrganizationService) Update(ctx context.Context, orgID uuid.UUID, userID string, req *models.OrganizationUpdateRequest) (*models.OrganizationResponse, error) {
	org, role, err := s.loadOrgWithRole(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin(role) {
		return nil, apperrors.ErrUnauthorizedOrganizationAccess
	}

	if req.Name != nil {
		org.Name = strings.TrimSpace(*req.Name)
	}
	if req.Slug != nil {
		org.Slug = strings.TrimSpace(strings.ToLower(*req.Slug))
	}
	if req.Description != nil {
		org.Description = req.Description
	}
	if req.BillingEmail != nil {
		org.BillingEmail = req.BillingEmail
	}
	if req.PlanTier != nil && *req.PlanTier != "" {
		org.PlanTier = *req.PlanTier
	}

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}
	return toOrgResponse(org, role), nil
}

func (s *OrganizationService) Delete(ctx context.Context, orgID uuid.UUID, userID string) error {
	_, role, err := s.loadOrgWithRole(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if role != OrgRoleOwner {
		return apperrors.ErrUnauthorizedOrganizationAccess
	}
	return s.orgRepo.Delete(ctx, orgID)
}

func (s *OrganizationService) ListMembers(ctx context.Context, orgID uuid.UUID, userID string) ([]models.OrganizationMemberResponse, error) {
	_, role, err := s.loadOrgWithRole(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin(role) {
		return nil, apperrors.ErrUnauthorizedOrganizationAccess
	}

	members, err := s.memberRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	out := make([]models.OrganizationMemberResponse, 0, len(members))
	for _, m := range members {
		out = append(out, models.OrganizationMemberResponse{
			UserID:       m.UserID,
			Role:         m.Role,
			JoinedAt:     m.JoinedAt,
			LastActiveAt: m.LastActiveAt,
			InvitedBy:    m.InvitedBy,
		})
	}
	return out, nil
}

func (s *OrganizationService) UpdateMemberRole(ctx context.Context, orgID uuid.UUID, targetUserID, actorID, role string) error {
	_, actorRole, err := s.loadOrgWithRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if !isAdmin(actorRole) {
		return apperrors.ErrUnauthorizedOrganizationAccess
	}
	if _, ok := allowedRoles[role]; !ok {
		return apperrors.Wrap(apperrors.ErrInvalidUserData, "invalid role")
	}
	return s.memberRepo.UpdateRole(ctx, orgID, targetUserID, role)
}

func (s *OrganizationService) RemoveMember(ctx context.Context, orgID uuid.UUID, targetUserID, actorID string) error {
	_, actorRole, err := s.loadOrgWithRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if !isAdmin(actorRole) && actorID != targetUserID {
		return apperrors.ErrUnauthorizedOrganizationAccess
	}
	return s.memberRepo.Delete(ctx, orgID, targetUserID)
}

func (s *OrganizationService) EnsureMembership(ctx context.Context, orgID *uuid.UUID, userID string) (*OrganizationContext, error) {
	if orgID == nil {
		return &OrganizationContext{Role: OrgRolePersonal}, nil
	}
	member, err := s.memberRepo.Find(ctx, *orgID, userID)
	if err != nil {
		return nil, err
	}
	return &OrganizationContext{ID: &member.OrganizationID, Role: member.Role}, nil
}

func (s *OrganizationService) CreateInvite(ctx context.Context, orgID uuid.UUID, actorID string, req *models.OrganizationInviteRequest) (*models.OrganizationInviteResponse, string, error) {
	if req == nil || strings.TrimSpace(req.Email) == "" {
		return nil, "", apperrors.Wrap(apperrors.ErrInvalidUserData, "email is required")
	}
	_, actorRole, err := s.loadOrgWithRole(ctx, orgID, actorID)
	if err != nil {
		return nil, "", err
	}
	if !isAdmin(actorRole) {
		return nil, "", apperrors.ErrUnauthorizedOrganizationAccess
	}
	role := req.Role
	if role == "" {
		role = OrgRoleMember
	}
	if _, ok := allowedRoles[role]; !ok {
		return nil, "", apperrors.Wrap(apperrors.ErrInvalidUserData, "invalid role")
	}

	token := uuid.NewString()
	tokenHash := hashToken(token)

	invite := &db.OrganizationInvite{
		OrganizationID: orgID,
		Email:          strings.ToLower(strings.TrimSpace(req.Email)),
		Role:           role,
		TokenHash:      tokenHash,
		InvitedBy:      &actorID,
		Message:        req.Message,
		ExpiresAt:      time.Now().UTC().Add(7 * 24 * time.Hour),
	}
	if err := s.inviteRepo.Create(ctx, invite); err != nil {
		return nil, "", err
	}

	return toInviteResponse(invite), token, nil
}

func (s *OrganizationService) ListInvites(ctx context.Context, orgID uuid.UUID, userID string) ([]*models.OrganizationInviteResponse, error) {
	_, actorRole, err := s.loadOrgWithRole(ctx, orgID, userID)
	if err != nil {
		return nil, err
	}
	if !isAdmin(actorRole) {
		return nil, apperrors.ErrUnauthorizedOrganizationAccess
	}
	invites, err := s.inviteRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make([]*models.OrganizationInviteResponse, 0, len(invites))
	for _, inv := range invites {
		out = append(out, toInviteResponse(inv))
	}
	return out, nil
}

func (s *OrganizationService) AcceptInvite(ctx context.Context, token string, userID string) (*models.OrganizationResponse, error) {
	if token == "" {
		return nil, apperrors.ErrOrganizationInviteInvalid
	}
	hash := hashToken(token)
	invite, err := s.inviteRepo.FindValidByToken(ctx, hash)
	if err != nil {
		return nil, err
	}

	member := &db.OrganizationMember{
		OrganizationID: invite.OrganizationID,
		UserID:         userID,
		Role:           invite.Role,
		JoinedAt:       time.Now().UTC(),
		InvitedBy:      invite.InvitedBy,
		InvitedAt:      &invite.CreatedAt,
	}
	if err := s.memberRepo.Upsert(ctx, member); err != nil {
		return nil, err
	}
	if err := s.inviteRepo.MarkAccepted(ctx, invite.ID, time.Now().UTC()); err != nil {
		return nil, err
	}

	org, err := s.orgRepo.FindByID(ctx, invite.OrganizationID)
	if err != nil {
		return nil, err
	}
	return toOrgResponse(org, invite.Role), nil
}

func (s *OrganizationService) loadOrgWithRole(ctx context.Context, orgID uuid.UUID, userID string) (*db.Organization, string, error) {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, "", err
	}
	member, err := s.memberRepo.Find(ctx, orgID, userID)
	if err != nil {
		return nil, "", err
	}
	return org, member.Role, nil
}

func toOrgResponse(org *db.Organization, role string) *models.OrganizationResponse {
	return &models.OrganizationResponse{
		ID:           org.ID,
		Name:         org.Name,
		Slug:         org.Slug,
		Description:  org.Description,
		BillingEmail: org.BillingEmail,
		PlanTier:     org.PlanTier,
		Role:         role,
		CreatedBy:    org.CreatedBy,
		CreatedAt:    org.CreatedAt,
		UpdatedAt:    org.UpdatedAt,
	}
}

func toInviteResponse(inv *db.OrganizationInvite) *models.OrganizationInviteResponse {
	return &models.OrganizationInviteResponse{
		ID:             inv.ID,
		OrganizationID: inv.OrganizationID,
		Email:          inv.Email,
		Role:           inv.Role,
		ExpiresAt:      inv.ExpiresAt,
		AcceptedAt:     inv.AcceptedAt,
		CreatedAt:      inv.CreatedAt,
	}
}

func isAdmin(role string) bool {
	return role == OrgRoleOwner || role == OrgRoleAdmin
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func slugify(val string) string {
	lower := strings.ToLower(strings.TrimSpace(val))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "org-" + uuid.NewString()[:8]
	}
	return slug
}

func coalesce(str *string, fallback string) string {
	if str != nil && strings.TrimSpace(*str) != "" {
		return *str
	}
	return fallback
}
