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
// backed by PostgreSQL.
type ConversationRepository struct {
	db *pgxpool.Pool
}

// NewConversationRepository constructs a ConversationRepository.
func NewConversationRepository(db *pgxpool.Pool) *ConversationRepository {
	return &ConversationRepository{
		db: db,
	}
}

const conversationColumns = `
	id,
	user_id,
	role,
	content,
	created_at
`

// Create appends a single turn to a user's conversation history.
func (r *ConversationRepository) Create(
	ctx context.Context,
	c *models.Conversation,
) (*models.Conversation, error) {

	if c == nil {
		return nil, fmt.Errorf(
			"%w: conversation cannot be nil",
			ErrInvalidConversationInput,
		)
	}

	if c.UserID == "" {
		return nil, fmt.Errorf(
			"%w: user_id is required",
			ErrInvalidConversationInput,
		)
	}

	if c.Content == "" {
		return nil, fmt.Errorf(
			"%w: content is required",
			ErrInvalidConversationInput,
		)
	}

	if !c.Role.Valid() {
		return nil, fmt.Errorf(
			"%w: role %q is not one of user|assistant|system",
			ErrInvalidConversationInput,
			c.Role,
		)
	}

	q := fmt.Sprintf(`
		INSERT INTO conversations (
			user_id,
			role,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING %s
	`, conversationColumns)

	row := r.db.QueryRow(
		ctx,
		q,
		c.UserID,
		c.Role,
		c.Content,
	)

	return scanConversation(row)
}

// CreateBatch inserts multiple conversation turns in one transaction.
//
// This is useful when saving both a user's message and the assistant's
// response after an AI request.
func (r *ConversationRepository) CreateBatch(
	ctx context.Context,
	turns []*models.Conversation,
) ([]*models.Conversation, error) {

	if len(turns) == 0 {
		return nil, nil
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"repositories: begin conversation batch: %w",
			err,
		)
	}

	defer tx.Rollback(ctx)

	q := fmt.Sprintf(`
		INSERT INTO conversations (
			user_id,
			role,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING %s
	`, conversationColumns)

	out := make(
		[]*models.Conversation,
		0,
		len(turns),
	)

	for _, turn := range turns {
		if turn == nil {
			return nil, fmt.Errorf(
				"%w: conversation turn cannot be nil",
				ErrInvalidConversationInput,
			)
		}

		if turn.UserID == "" {
			return nil, fmt.Errorf(
				"%w: user_id is required",
				ErrInvalidConversationInput,
			)
		}

		if turn.Content == "" {
			return nil, fmt.Errorf(
				"%w: content is required",
				ErrInvalidConversationInput,
			)
		}

		if !turn.Role.Valid() {
			return nil, fmt.Errorf(
				"%w: role %q is invalid",
				ErrInvalidConversationInput,
				turn.Role,
			)
		}

		row := tx.QueryRow(
			ctx,
			q,
			turn.UserID,
			turn.Role,
			turn.Content,
		)

		conversation, err := scanConversation(row)
		if err != nil {
			return nil, err
		}

		out = append(out, conversation)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf(
			"repositories: commit conversation batch: %w",
			err,
		)
	}

	return out, nil
}

// ListRecentByUserID returns a user's most recent conversation turns.
//
// Results are returned in chronological order, oldest first.
func (r *ConversationRepository) ListRecentByUserID(
	ctx context.Context,
	userID string,
	limit int,
) ([]*models.Conversation, error) {

	if userID == "" {
		return nil, fmt.Errorf(
			"%w: user_id is required",
			ErrInvalidConversationInput,
		)
	}

	if limit <= 0 {
		limit = 20
	}

	q := fmt.Sprintf(`
		SELECT %s
		FROM conversations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, conversationColumns)

	rows, err := r.db.Query(
		ctx,
		q,
		userID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"repositories: list conversations: %w",
			err,
		)
	}

	defer rows.Close()

	var conversations []*models.Conversation

	for rows.Next() {
		conversation, err := scanConversationRow(rows)
		if err != nil {
			return nil, fmt.Errorf(
				"repositories: scan conversation row: %w",
				err,
			)
		}

		conversations = append(
			conversations,
			conversation,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"repositories: list conversations: %w",
			err,
		)
	}

	// The database returns newest first.
	// Reverse so the service receives chronological history.
	for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
		conversations[i], conversations[j] =
			conversations[j], conversations[i]
	}

	return conversations, nil
}

// DeleteByUserID removes all conversation history for a user.
//
// This can be used during account deletion or GDPR erasure.
func (r *ConversationRepository) DeleteByUserID(
	ctx context.Context,
	userID string,
) error {

	if userID == "" {
		return fmt.Errorf(
			"%w: user_id is required",
			ErrInvalidConversationInput,
		)
	}

	const q = `
		DELETE FROM conversations
		WHERE user_id = $1
	`

	if _, err := r.db.Exec(
		ctx,
		q,
		userID,
	); err != nil {
		return fmt.Errorf(
			"repositories: delete conversations: %w",
			err,
		)
	}

	return nil
}

type conversationRow interface {
	Scan(dest ...any) error
}

func scanConversation(
	row pgx.Row,
) (*models.Conversation, error) {

	conversation, err := scanConversationRow(row)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConversationNotFound
		}

		return nil, fmt.Errorf(
			"repositories: query conversation: %w",
			err,
		)
	}

	return conversation, nil
}

func scanConversationRow(
	row conversationRow,
) (*models.Conversation, error) {

	conversation := &models.Conversation{}

	var role string

	if err := row.Scan(
		&conversation.ID,
		&conversation.UserID,
		&role,
		&conversation.Content,
		&conversation.CreatedAt,
	); err != nil {
		return nil, err
	}

	conversation.Role = models.ConversationRole(role)

	return conversation, nil
}
