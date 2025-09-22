package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	apperrors "github.com/yourusername/vectorchat/internal/errors"
)

type SharedKnowledgeBaseRepository struct {
	db *Database
}

func NewSharedKnowledgeBaseRepository(db *Database) *SharedKnowledgeBaseRepository {
	return &SharedKnowledgeBaseRepository{db: db}
}

// Create inserts a new shared knowledge base owned by a user.
func (r *SharedKnowledgeBaseRepository) Create(ctx context.Context, kb *SharedKnowledgeBase) error {
	if kb.ID == uuid.Nil {
		kb.ID = uuid.New()
	}
	now := time.Now().UTC()
	if kb.CreatedAt.IsZero() {
		kb.CreatedAt = now
	}
	kb.UpdatedAt = now

	query := `
        INSERT INTO shared_knowledge_bases (id, owner_id, name, description, created_at, updated_at)
        VALUES (:id, :owner_id, :name, :description, :created_at, :updated_at)
    `

	if _, err := r.db.NamedExecContext(ctx, query, kb); err != nil {
		if IsDuplicateKeyError(err) {
			return apperrors.ErrChatbotAlreadyExists
		}
		return apperrors.Wrap(err, "failed to create shared knowledge base")
	}

	return nil
}

// Update modifies the name/description of a shared knowledge base.
func (r *SharedKnowledgeBaseRepository) Update(ctx context.Context, kb *SharedKnowledgeBase) error {
	kb.UpdatedAt = time.Now().UTC()

	query := `
        UPDATE shared_knowledge_bases
        SET name = :name,
            description = :description,
            updated_at = :updated_at
        WHERE id = :id AND owner_id = :owner_id
    `

	result, err := r.db.NamedExecContext(ctx, query, kb)
	if err != nil {
		return apperrors.Wrap(err, "failed to update shared knowledge base")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}
	if rows == 0 {
		return apperrors.ErrSharedKnowledgeBaseNotFound
	}

	return nil
}

// Delete removes a shared knowledge base by ID.
func (r *SharedKnowledgeBaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM shared_knowledge_bases WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return apperrors.Wrap(err, "failed to delete shared knowledge base")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return apperrors.Wrap(err, "failed to get rows affected")
	}
	if rows == 0 {
		return apperrors.ErrSharedKnowledgeBaseNotFound
	}

	return nil
}

// FindByID retrieves a shared knowledge base by ID.
func (r *SharedKnowledgeBaseRepository) FindByID(ctx context.Context, id uuid.UUID) (*SharedKnowledgeBase, error) {
	var kb SharedKnowledgeBase
	query := `
        SELECT id, owner_id, name, description, created_at, updated_at
        FROM shared_knowledge_bases
        WHERE id = $1
    `

	if err := r.db.GetContext(ctx, &kb, query, id); err != nil {
		if IsNoRowsError(err) {
			return nil, apperrors.ErrSharedKnowledgeBaseNotFound
		}
		return nil, apperrors.Wrap(err, "failed to find shared knowledge base")
	}

	return &kb, nil
}

// ListByOwner returns all shared knowledge bases owned by a user.
func (r *SharedKnowledgeBaseRepository) ListByOwner(ctx context.Context, ownerID string) ([]*SharedKnowledgeBase, error) {
	var results []*SharedKnowledgeBase
	query := `
        SELECT id, owner_id, name, description, created_at, updated_at
        FROM shared_knowledge_bases
        WHERE owner_id = $1
        ORDER BY created_at DESC
    `

	if err := r.db.SelectContext(ctx, &results, query, ownerID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list shared knowledge bases")
	}

	return results, nil
}

// ListIDsByChatbot returns shared knowledge base IDs linked to a chatbot.
func (r *SharedKnowledgeBaseRepository) ListIDsByChatbot(ctx context.Context, chatbotID uuid.UUID) ([]uuid.UUID, error) {
	query := `
        SELECT shared_knowledge_base_id
        FROM chatbot_shared_knowledge_bases
        WHERE chatbot_id = $1
        ORDER BY created_at
    `

	var ids []uuid.UUID
	if err := r.db.SelectContext(ctx, &ids, query, chatbotID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list shared knowledge base ids for chatbot")
	}

	return ids, nil
}

// ListByChatbot returns shared knowledge base objects linked to a chatbot.
func (r *SharedKnowledgeBaseRepository) ListByChatbot(ctx context.Context, chatbotID uuid.UUID) ([]*SharedKnowledgeBase, error) {
	var results []*SharedKnowledgeBase
	query := `
        SELECT skb.id, skb.owner_id, skb.name, skb.description, skb.created_at, skb.updated_at
        FROM shared_knowledge_bases skb
        INNER JOIN chatbot_shared_knowledge_bases link ON link.shared_knowledge_base_id = skb.id
        WHERE link.chatbot_id = $1
        ORDER BY skb.name
    `

	if err := r.db.SelectContext(ctx, &results, query, chatbotID); err != nil {
		return nil, apperrors.Wrap(err, "failed to list shared knowledge bases for chatbot")
	}

	return results, nil
}

// ReplaceChatbotLinks replaces the set of shared knowledge bases attached to a chatbot.
func (r *SharedKnowledgeBaseRepository) ReplaceChatbotLinks(ctx context.Context, chatbotID uuid.UUID, kbIDs []uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return apperrors.Wrap(err, "failed to start transaction")
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM chatbot_shared_knowledge_bases WHERE chatbot_id = $1`, chatbotID); err != nil {
		return apperrors.Wrap(err, "failed to clear existing knowledge base links")
	}

	if len(kbIDs) > 0 {
		insertQuery := `
            INSERT INTO chatbot_shared_knowledge_bases (chatbot_id, shared_knowledge_base_id)
            VALUES ($1, $2)
            ON CONFLICT (chatbot_id, shared_knowledge_base_id) DO NOTHING
        `
		for _, id := range kbIDs {
			if _, err := tx.ExecContext(ctx, insertQuery, chatbotID, id); err != nil {
				if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
					switch pqErr.Constraint {
					case "chatbot_shared_knowledge_bases_chatbot_id_fkey":
						return apperrors.ErrChatbotNotFound
					case "chatbot_shared_knowledge_bases_shared_knowledge_base_id_fkey":
						return apperrors.ErrSharedKnowledgeBaseNotFound
					}
				}
				return apperrors.Wrap(err, "failed to insert shared knowledge base link")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return apperrors.Wrap(err, "failed to commit knowledge base link changes")
	}

	return nil
}

// DetachChatbotLink removes a single shared knowledge base link from a chatbot.
func (r *SharedKnowledgeBaseRepository) DetachChatbotLink(ctx context.Context, chatbotID, kbID uuid.UUID) error {
	query := `
        DELETE FROM chatbot_shared_knowledge_bases
        WHERE chatbot_id = $1 AND shared_knowledge_base_id = $2
    `

	if _, err := r.db.ExecContext(ctx, query, chatbotID, kbID); err != nil {
		return apperrors.Wrap(err, "failed to detach shared knowledge base link")
	}
	return nil
}

// AttachChatbotLink links a single shared knowledge base to a chatbot.
func (r *SharedKnowledgeBaseRepository) AttachChatbotLink(ctx context.Context, chatbotID, kbID uuid.UUID) error {
	query := `
        INSERT INTO chatbot_shared_knowledge_bases (chatbot_id, shared_knowledge_base_id)
        VALUES ($1, $2)
        ON CONFLICT (chatbot_id, shared_knowledge_base_id) DO NOTHING
    `

	if _, err := r.db.ExecContext(ctx, query, chatbotID, kbID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			switch pqErr.Constraint {
			case "chatbot_shared_knowledge_bases_chatbot_id_fkey":
				return apperrors.ErrChatbotNotFound
			case "chatbot_shared_knowledge_bases_shared_knowledge_base_id_fkey":
				return apperrors.ErrSharedKnowledgeBaseNotFound
			}
		}
		return apperrors.Wrap(err, "failed to attach shared knowledge base")
	}

	return nil
}

// DetachAllOwned removes all links for knowledge bases owned by a user (used on delete).
func (r *SharedKnowledgeBaseRepository) DetachAllOwned(ctx context.Context, ownerID string) error {
	query := `
        DELETE FROM chatbot_shared_knowledge_bases
        WHERE shared_knowledge_base_id IN (
            SELECT id FROM shared_knowledge_bases WHERE owner_id = $1
        )
    `

	if _, err := r.db.ExecContext(ctx, query, ownerID); err != nil {
		return apperrors.Wrap(err, "failed to detach knowledge base links for owner")
	}
	return nil
}

// ListOwnersByKnowledgeBaseIDs returns owner IDs for a set of shared KBs.
func (r *SharedKnowledgeBaseRepository) ListOwnersByKnowledgeBaseIDs(ctx context.Context, kbIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	result := make(map[uuid.UUID]string)
	if len(kbIDs) == 0 {
		return result, nil
	}

	query := `
        SELECT id, owner_id
        FROM shared_knowledge_bases
        WHERE id = ANY($1)
    `

	type row struct {
		ID      uuid.UUID `db:"id"`
		OwnerID string    `db:"owner_id"`
	}

	var rows []row
	if err := r.db.SelectContext(ctx, &rows, query, pq.Array(kbIDs)); err != nil {
		return nil, apperrors.Wrap(err, "failed to lookup knowledge base owners")
	}

	for _, r := range rows {
		result[r.ID] = r.OwnerID
	}

	return result, nil
}
