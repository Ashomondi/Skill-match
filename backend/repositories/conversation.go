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

	out := make([]*models.Conversation, 0, len(turns))
	for _, t := range turns {
		if t == nil || t.UserID == "" || t.Content == "" || !t.Role.Valid() {
			return nil, fmt.Errorf("%w: all turns require user_id, valid role, and content", ErrInvalidConversationInput)

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

<<<<<<< HEAD
// ListRecentByUserID returns a user's most recent conversation turns in
// chronological order (oldest first), suitable for feeding directly into
// a Bedrock prompt as message history. limit <= 0 defaults to 20.
func (r *ConversationRepository) ListRecentByUserID(ctx context.Context, userID string, limit int) ([]*models.Conversation, error) {
=======
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

>>>>>>> dev
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

<<<<<<< HEAD
	var out []*models.Conversation
=======
	var conversations []*models.Conversation

>>>>>>> dev
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

<<<<<<< HEAD
func scanConversation(rw pgx.Row) (*models.Conversation, error) {
	c, err := scanConversationRow(rw)
=======
func scanConversation(
	row pgx.Row,
) (*models.Conversation, error) {

	conversation, err := scanConversationRow(row)

>>>>>>> dev
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

<<<<<<< HEAD
func scanConversationRow(rw conversationRow) (*models.Conversation, error) {
	c := &models.Conversation{}
=======
func scanConversationRow(
	row conversationRow,
) (*models.Conversation, error) {

	conversation := &models.Conversation{}

>>>>>>> dev
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
<<<<<<< HEAD
	c.Role = models.ConversationRole(role)
	return c, nil
=======

	conversation.Role = models.ConversationRole(role)

	return conversation, nil
>>>>>>> dev
}
