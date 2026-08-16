package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/models"
)

// Sentinel errors for conversation operations.
var (
	ErrConversationNotFound     = errors.New("repositories: conversation turn not found")
	ErrInvalidConversationInput = errors.New("repositories: invalid conversation input")
)

// ConversationRepository provides persistence operations for chat history
// backed by CockroachDB. It operates on the canonical models.Conversation
// type and models.ConversationRole.
type ConversationRepository struct {
	db *pgxpool.Pool
}

// NewConversationRepository constructs a ConversationRepository.
func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{db: db}
}

const conversationColumns = `id, user_id, role, content, created_at`

// Create appends a single turn to a user's conversation history. Chat
// turns are append-only; there is no Update.
func (r *ConversationRepository) Create(ctx context.Context, c *models.Conversation) (*models.Conversation, error) {
	if c == nil || c.UserID == "" || c.Content == "" {
		return nil, fmt.Errorf("%w: user_id and content are required", ErrInvalidConversationInput)
	}
	if !c.Role.Valid() {
		return nil, fmt.Errorf("%w: role %q is not one of user|assistant|system", ErrInvalidConversationInput, c.Role)
	}

	q := fmt.Sprintf(`
		INSERT INTO conversations (user_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING %s`, conversationColumns)

	row := r.db.QueryRow(ctx, q, c.UserID, c.Role, c.Content)
	return scanConversation(row)
}

// CreateBatch inserts multiple turns in a single round trip, e.g. a user
// prompt and the assistant's reply written together after a chat
// completion. All rows are inserted in one transaction — either all
// succeed or none do.
func (r *ConversationRepository) CreateBatch(ctx context.Context, turns []*models.Conversation) ([]*models.Conversation, error) {
	if len(turns) == 0 {
		return nil, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repositories: begin conversation batch: %w", err)
	}
	defer tx.Rollback(ctx) // no-op if committed

	q := fmt.Sprintf(`
		INSERT INTO conversations (user_id, role, content)
		VALUES ($1, $2, $3)
		RETURNING %s`, conversationColumns)

	out := make([]*models.Conversation, 0, len(turns))
	for _, t := range turns {
		if t == nil || t.UserID == "" || t.Content == "" || !t.Role.Valid() {
			return nil, fmt.Errorf("%w: all turns require user_id, valid role, and content", ErrInvalidConversationInput)
		}
		row := tx.QueryRow(ctx, q, t.UserID, t.Role, t.Content)
		c, err := scanConversation(row)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repositories: commit conversation batch: %w", err)
	}
	return out, nil
}

// ListRecentByUserID returns a user's most recent conversation turns in
// chronological order (oldest first), suitable for feeding directly into
// a Bedrock prompt as message history. limit <= 0 defaults to 20.
func (r *ConversationRepository) ListRecentByUserID(ctx context.Context, userID string, limit int) ([]*models.Conversation, error) {
	if limit <= 0 {
		limit = 20
	}

	// Fetch newest-first (index order), then reverse in Go so callers get
	// chronological order without CockroachDB needing to sort twice.
	q := fmt.Sprintf(`
		SELECT %s FROM conversations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, conversationColumns)

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("repositories: list conversations: %w", err)
	}
	defer rows.Close()

	var out []*models.Conversation
	for rows.Next() {
		c, err := scanConversationRow(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan conversation row: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: list conversations: %w", err)
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

// DeleteByUserID removes all conversation history for a user. Intended
// for account-deletion / GDPR erasure flows, not routine use.
func (r *ConversationRepository) DeleteByUserID(ctx context.Context, userID string) error {
	const q = `DELETE FROM conversations WHERE user_id = $1`

	if _, err := r.db.Exec(ctx, q, userID); err != nil {
		return fmt.Errorf("repositories: delete conversations: %w", err)
	}
	return nil
}

type conversationRow interface {
	Scan(dest ...any) error
}

func scanConversation(rw pgx.Row) (*models.Conversation, error) {
	c, err := scanConversationRow(rw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("repositories: query conversation: %w", err)
	}
	return c, nil
}

func scanConversationRow(rw conversationRow) (*models.Conversation, error) {
	c := &models.Conversation{}
	var role string
	if err := rw.Scan(&c.ID, &c.UserID, &role, &c.Content, &c.CreatedAt); err != nil {
		return nil, err
	}
	c.Role = models.ConversationRole(role)
	return c, nil
}
