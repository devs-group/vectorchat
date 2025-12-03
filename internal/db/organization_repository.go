package db

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type OrganizationRepository struct {
	db *Database
}

func NewOrganizationRepository(db *Database) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *Organization) error {
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now().UTC()
	if org.CreatedAt.IsZero() {
		org.CreatedAt = now
	}
	if org.UpdatedAt.IsZero() {
		org.UpdatedAt = now
	}
	if org.PlanTier == "" {
		org.PlanTier = "free"
	}
	org.Slug = strings.ToLower(org.Slug)

	query := `
		INSERT INTO organizations (id, name, slug, description, billing_email, plan_tier, created_by, created_at, updated_at)
		VALUES (:id, :name, :slug, :description, :billing_email, :plan_tier, :created_by, :created_at, :updated_at)
	`

	if _, err := r.db.NamedExecContext(ctx, query, org); err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrOrganizationAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create organization")
	}

	return nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *Organization) error {
	org.UpdatedAt = time.Now().UTC()
	org.Slug = strings.ToLower(org.Slug)

	query := `
		UPDATE organizations
		SET name = :name,
			slug = :slug,
			description = :description,
			billing_email = :billing_email,
			plan_tier = :plan_tier,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, org)
	if err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrOrganizationAlreadyExists
		}
		return apperrors.Wrap(err, "failed to update organization")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrOrganizationNotFound
	}

	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete organization")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrOrganizationNotFound
	}
	return nil
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	var org Organization
	const query = `
		SELECT id, name, slug, description, billing_email, plan_tier, created_by, created_at, updated_at
		FROM organizations
		WHERE id = $1
	`
	if err := r.db.GetContext(ctx, &org, query, id); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find organization")
	}
	return &org, nil
}

func (r *OrganizationRepository) FindBySlug(ctx context.Context, slug string) (*Organization, error) {
	var org Organization
	const query = `
		SELECT id, name, slug, description, billing_email, plan_tier, created_by, created_at, updated_at
		FROM organizations
		WHERE slug = $1
	`
	if err := r.db.GetContext(ctx, &org, query, strings.ToLower(slug)); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find organization by slug")
	}
	return &org, nil
}

type OrganizationWithRole struct {
	Organization
	Role string `db:"role"`
}

func (r *OrganizationRepository) ListForUser(ctx context.Context, userID string) ([]OrganizationWithRole, error) {
	const query = `
		SELECT o.id, o.name, o.slug, o.description, o.billing_email, o.plan_tier, o.created_by, o.created_at, o.updated_at, m.role
		FROM organizations o
		INNER JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.created_at DESC
	`
	var result []OrganizationWithRole
	if err := r.db.SelectContext(ctx, &result, query, userID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list organizations for user")
	}
	return result, nil
}
