package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type OrganizationInviteRepository struct {
	db *Database
}

func NewOrganizationInviteRepository(db *Database) *OrganizationInviteRepository {
	return &OrganizationInviteRepository{db: db}
}

func (r *OrganizationInviteRepository) Create(ctx context.Context, invite *OrganizationInvite) error {
	if invite.ID == uuid.Nil {
		invite.ID = uuid.New()
	}
	now := time.Now().UTC()
	if invite.CreatedAt.IsZero() {
		invite.CreatedAt = now
	}
	if invite.ExpiresAt.IsZero() {
		invite.ExpiresAt = now.Add(7 * 24 * time.Hour)
	}
	query := `
		INSERT INTO organization_invites (id, organization_id, email, role, token_hash, invited_by, message, expires_at, created_at, accepted_at)
		VALUES (:id, :organization_id, :email, :role, :token_hash, :invited_by, :message, :expires_at, :created_at, :accepted_at)
		ON CONFLICT (organization_id, email) DO UPDATE
		SET role = EXCLUDED.role,
			token_hash = EXCLUDED.token_hash,
			invited_by = EXCLUDED.invited_by,
			message = EXCLUDED.message,
			expires_at = EXCLUDED.expires_at,
			accepted_at = EXCLUDED.accepted_at,
			created_at = EXCLUDED.created_at
	`
	if _, err := r.db.NamedExecContext(ctx, query, invite); err != nil {
		return apperrors.Wrap(err, "failed to upsert invite")
	}
	return nil
}

func (r *OrganizationInviteRepository) FindValidByToken(ctx context.Context, tokenHash string) (*OrganizationInvite, error) {
	const query = `
		SELECT id, organization_id, email, role, token_hash, invited_by, message, expires_at, created_at, accepted_at
		FROM organization_invites
		WHERE token_hash = $1 AND expires_at > NOW() AND accepted_at IS NULL
	`
	var invite OrganizationInvite
	if err := r.db.GetContext(ctx, &invite, query, tokenHash); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrOrganizationInviteInvalid
		}
		return nil, apperrors.Wrap(err, "failed to find invite by token")
	}
	return &invite, nil
}

func (r *OrganizationInviteRepository) MarkAccepted(ctx context.Context, id uuid.UUID, acceptedAt time.Time) error {
	res, err := r.db.ExecContext(ctx, `
		UPDATE organization_invites
		SET accepted_at = $1
		WHERE id = $2
	`, acceptedAt, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to mark invite accepted")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to read rows affected")
	}
	if rows == 0 {
		return apperrors.ErrOrganizationInviteInvalid
	}
	return nil
}

func (r *OrganizationInviteRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*OrganizationInvite, error) {
	const query = `
		SELECT id, organization_id, email, role, token_hash, invited_by, message, expires_at, created_at, accepted_at
		FROM organization_invites
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`
	var invites []*OrganizationInvite
	if err := r.db.SelectContext(ctx, &invites, query, orgID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list organization invites")
	}
	return invites, nil
}

func (r *OrganizationInviteRepository) DeleteByOrgAndEmail(ctx context.Context, orgID uuid.UUID, email string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM organization_invites WHERE organization_id = $1 AND email = $2
	`, orgID, email)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete organization invite")
	}
	return nil
}
