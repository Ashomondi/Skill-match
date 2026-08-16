package models

import "time"

// EmbeddingDim is the fixed vector dimension enforced by the VECTOR(1536)
// column in migrations/003_memory.sql (Titan Embeddings V2). Changing the
// embedding model requires a new migration and backfill.
const EmbeddingDim = 1536

// EmbeddingSourceType mirrors the CHECK constraint in migrations/003_memory.sql.
type EmbeddingSourceType string

const (
	EmbeddingSourceResume       EmbeddingSourceType = "resume"
	EmbeddingSourceConversation EmbeddingSourceType = "conversation"
	EmbeddingSourceJob          EmbeddingSourceType = "job"
)

func (s EmbeddingSourceType) Valid() bool {
	switch s {
	case EmbeddingSourceResume, EmbeddingSourceConversation, EmbeddingSourceJob:
		return true
	default:
		return false
	}
}

// Embedding is a vector representation of some source entity (a resume,
// a conversation turn, or a job description).
type Embedding struct {
	ID         string
	UserID     string
	SourceType EmbeddingSourceType
	SourceID   string
	Vector     []float32
	CreatedAt  time.Time
}

// SimilarEmbedding is a search result: the matched embedding plus its
// distance from the query vector. Lower Distance means more similar.
type SimilarEmbedding struct {
	Embedding
	Distance float64
}
