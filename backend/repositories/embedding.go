package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"

	"skill-match/backend/models"
)

// Sentinel errors for embedding operations.
var (
	ErrEmbeddingNotFound       = errors.New("repositories: embedding not found")
	ErrInvalidEmbeddingInput   = errors.New("repositories: invalid embedding input")
	ErrEmbeddingWrongDimension = errors.New("repositories: embedding vector has wrong dimension")
)

// EmbeddingRepository provides persistence and similarity search for
// vector embeddings backed by CockroachDB's Distributed Vector Index. It
// operates on the canonical models.Embedding type; vector dimensions are
// enforced against models.EmbeddingDim.
type EmbeddingRepository struct {
	db *pgxpool.Pool
}

// NewEmbeddingRepository constructs an EmbeddingRepository.
func NewEmbeddingRepository(db *pgxpool.Pool) *EmbeddingRepository {
	return &EmbeddingRepository{db: db}
}

const embeddingColumns = `id, user_id, source_type, source_id, vector, created_at`

// Upsert inserts an embedding for (source_type, source_id), or replaces
// the existing one if that source has already been embedded — re-parsing
// a resume or re-summarizing a conversation should not accumulate stale
// vectors. This relies on the unique index on (source_type, source_id)
// defined in migrations/003_memory.sql.
func (r *EmbeddingRepository) Upsert(ctx context.Context, e *models.Embedding) (*models.Embedding, error) {
	if e == nil || e.UserID == "" || e.SourceID == "" {
		return nil, fmt.Errorf("%w: user_id and source_id are required", ErrInvalidEmbeddingInput)
	}
	if !e.SourceType.Valid() {
		return nil, fmt.Errorf("%w: source_type %q is not one of resume|conversation|job", ErrInvalidEmbeddingInput, e.SourceType)
	}
	if len(e.Vector) != models.EmbeddingDim {
		return nil, fmt.Errorf("%w: got %d dims, want %d", ErrEmbeddingWrongDimension, len(e.Vector), models.EmbeddingDim)
	}

	q := fmt.Sprintf(`
		INSERT INTO embeddings (user_id, source_type, source_id, vector)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (source_type, source_id)
		DO UPDATE SET vector = excluded.vector, user_id = excluded.user_id
		RETURNING %s`, embeddingColumns)

	row := r.db.QueryRow(ctx, q, e.UserID, e.SourceType, e.SourceID, pgvector.NewVector(e.Vector))
	return scanEmbedding(row)
}

// GetBySource fetches the embedding for a specific source row. Returns
// ErrEmbeddingNotFound if that source has not been embedded yet.
func (r *EmbeddingRepository) GetBySource(ctx context.Context, sourceType models.EmbeddingSourceType, sourceID string) (*models.Embedding, error) {
	q := fmt.Sprintf(`SELECT %s FROM embeddings WHERE source_type = $1 AND source_id = $2`, embeddingColumns)
	return scanEmbedding(r.db.QueryRow(ctx, q, sourceType, sourceID))
}

// FindSimilar performs approximate nearest-neighbor search using the
// Distributed Vector Index, returning the k closest embeddings to
// queryVector ordered nearest-first. Optionally restricted to a single
// sourceType (e.g. only match against "job" embeddings) and/or a single
// userID (e.g. only search a user's own conversation history for memory
// recall). Pass "" / empty EmbeddingSourceType to skip a filter.
//
// This is the query services/matching.go and services/memory.go depend
// on; it must stay index-backed (ORDER BY <-> ... LIMIT), never fall back
// to scanning + sorting in application code.
func (r *EmbeddingRepository) FindSimilar(ctx context.Context, queryVector []float32, sourceType models.EmbeddingSourceType, userID string, k int) ([]models.SimilarEmbedding, error) {
	if len(queryVector) != models.EmbeddingDim {
		return nil, fmt.Errorf("%w: got %d dims, want %d", ErrEmbeddingWrongDimension, len(queryVector), models.EmbeddingDim)
	}
	if k <= 0 {
		k = 10
	}

	q := fmt.Sprintf(`
		SELECT %s, vector <=> $1 AS distance
		FROM embeddings
		WHERE ($2 = '' OR source_type = $2)
		  AND ($3 = '' OR user_id = $3)
		ORDER BY vector <=> $1
		LIMIT $4`, embeddingColumns)

	qv := pgvector.NewVector(queryVector)
	rows, err := r.db.Query(ctx, q, qv, string(sourceType), userID, k)
	if err != nil {
		return nil, fmt.Errorf("repositories: find similar embeddings: %w", err)
	}
	defer rows.Close()

	var out []models.SimilarEmbedding
	for rows.Next() {
		se, err := scanSimilarEmbeddingRow(rows)
		if err != nil {
			return nil, fmt.Errorf("repositories: scan similar embedding row: %w", err)
		}
		out = append(out, *se)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repositories: find similar embeddings: %w", err)
	}
	return out, nil
}

// DeleteBySource removes the embedding for a specific source row, e.g.
// when a resume is deleted or replaced. No-op (not an error) if none
// exists.
func (r *EmbeddingRepository) DeleteBySource(ctx context.Context, sourceType models.EmbeddingSourceType, sourceID string) error {
	const q = `DELETE FROM embeddings WHERE source_type = $1 AND source_id = $2`

	if _, err := r.db.Exec(ctx, q, sourceType, sourceID); err != nil {
		return fmt.Errorf("repositories: delete embedding: %w", err)
	}
	return nil
}

type embeddingRow interface {
	Scan(dest ...any) error
}

func scanEmbedding(rw pgx.Row) (*models.Embedding, error) {
	e := &models.Embedding{}
	var sourceType string
	var vec pgvector.Vector
	err := rw.Scan(&e.ID, &e.UserID, &sourceType, &e.SourceID, &vec, &e.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmbeddingNotFound
		}
		return nil, fmt.Errorf("repositories: query embedding: %w", err)
	}
	e.SourceType = models.EmbeddingSourceType(sourceType)
	e.Vector = vec.Slice()
	return e, nil
}

func scanSimilarEmbeddingRow(rw embeddingRow) (*models.SimilarEmbedding, error) {
	se := &models.SimilarEmbedding{}
	var sourceType string
	var vec pgvector.Vector
	err := rw.Scan(&se.ID, &se.UserID, &sourceType, &se.SourceID, &vec, &se.CreatedAt, &se.Distance)
	if err != nil {
		return nil, err
	}
	se.SourceType = models.EmbeddingSourceType(sourceType)
	se.Vector = vec.Slice()
	return se, nil
}
