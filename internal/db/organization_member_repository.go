package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type OrganizationMemberRepository struct {
	db *Database
}

func NewOrganizationMemberRepository(db *Database) *OrganizationMemberRepository {
	return &OrganizationMemberRepository{db: db}
}

func (r *OrganizationMemberRepository) Upsert(ctx context.Context, member *OrganizationMember) error {
	if member.ID == uuid.Nil {
		member.ID = uuid.New()
	}
	now := time.Now().UTC()
	if member.JoinedAt.IsZero() {
		member.JoinedAt = now
	}
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, invited_by, invited_at, joined_at, last_active_at)
		VALUES (:id, :organization_id, :user_id, :role, :invited_by, :invited_at, :joined_at, :last_active_at)
		ON CONFLICT (organization_id, user_id)
		DO UPDATE SET role = EXCLUDED.role,
			invited_by = COALESCE(EXCLUDED.invited_by, organization_members.invited_by),
			invited_at = COALESCE(EXCLUDED.invited_at, organization_members.invited_at),
			last_active_at = COALESCE(EXCLUDED.last_active_at, organization_members.last_active_at)
	`
	if _, err := r.db.NamedExecContext(ctx, query, member); err != nil {
		return apperrors.Wrap(err, "failed to upsert organization member")
	}
	return nil
}

func (r *OrganizationMemberRepository) Find(ctx context.Context, orgID uuid.UUID, userID string) (*OrganizationMember, error) {
	const query = `
		SELECT id, organization_id, user_id, role, invited_by, invited_at, joined_at, last_active_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`
	var member OrganizationMember
	if err := r.db.GetContext(ctx, &member, query, orgID, userID); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrUnauthorizedOrganizationAccess
		}
		return nil, apperrors.Wrap(err, "failed to load membership")
	}
	return &member, nil
}

func (r *OrganizationMemberRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationMember, error) {
	const query = `
		SELECT id, organization_id, user_id, role, invited_by, invited_at, joined_at, last_active_at
		FROM organization_members
		WHERE organization_id = $1
		ORDER BY joined_at ASC
	`
	var members []*OrganizationMember
	if err := r.db.SelectContext(ctx, &members, query, orgID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list organization members")
	}
	return members, nil
}

func (r *OrganizationMemberRepository) UpdateRole(ctx context.Context, orgID uuid.UUID, userID, role string) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE organization_members
		SET role = $1
		WHERE organization_id = $2 AND user_id = $3
	`, role, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to update member role")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrUnauthorizedOrganizationAccess
	}
	return nil
}

func (r *OrganizationMemberRepository) Delete(ctx context.Context, orgID uuid.UUID, userID string) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete organization member")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrUnauthorizedOrganizationAccess
	}
	return nil
}

func (r *OrganizationMemberRepository) TouchLastActive(ctx context.Context, orgID uuid.UUID, userID string, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE organization_members
		SET last_active_at = $1
		WHERE organization_id = $2 AND user_id = $3
	`, at, orgID, userID)
	if err != nil {
		return apperrors.Wrap(err, "failed to update last active")
	}
	return nil
}
