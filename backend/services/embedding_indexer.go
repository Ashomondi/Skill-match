package services

import (
	"context"
	"log"

	"skill-match/backend/repositories"
)

// embeddingGenerator is the text→vector producer. clients.BedrockEmbedder and
// clients.LocalEmbedder both satisfy it, keeping the provider swappable.
type embeddingGenerator interface {
	Generate(ctx context.Context, text string) ([]float32, error)
}

// embeddingStore is the pgvector persistence needed by the indexer.
// repositories.EmbeddingRepository satisfies it.
type embeddingStore interface {
	Upsert(ctx context.Context, e *repositories.Embedding) (*repositories.Embedding, error)
	DeleteBySource(ctx context.Context, sourceType repositories.EmbeddingSourceType, sourceID string) error
}

// EmbeddingIndexer generates and persists embeddings for the entity types the
// matching engine queries against: resumes (user-scoped) and jobs (global
// corpus). It is optional — when no generator is wired, indexers are no-ops —
// so storage and ingestion work without vector features enabled.
type EmbeddingIndexer struct {
	generator embeddingGenerator
	store     embeddingStore
}

// NewEmbeddingIndexer builds an indexer. Pass nil generator/store to disable.
func NewEmbeddingIndexer(generator embeddingGenerator, store embeddingStore) *EmbeddingIndexer {
	return &EmbeddingIndexer{generator: generator, store: store}
}

// IndexResume embeds a resume's parsed text under its resume id, scoped to the
// owning user. Errors are logged, never returned: a failed embedding must not
// block a successful upload.
func (ix *EmbeddingIndexer) IndexResume(ctx context.Context, resumeID, userID, text string) {
	if ix == nil || ix.generator == nil || ix.store == nil {
		return
	}
	if resumeID == "" || userID == "" || text == "" {
		return
	}

	vector, err := ix.generator.Generate(ctx, text)
	if err != nil {
		log.Printf("WARNING: failed to generate resume embedding: %v", err)
		return
	}
	if _, err := ix.store.Upsert(ctx, &repositories.Embedding{
		UserID:     userID,
		SourceType: repositories.EmbeddingSourceResume,
		SourceID:   resumeID,
		Vector:     vector,
	}); err != nil {
		log.Printf("WARNING: failed to store resume embedding: %v", err)
	}
}

// IndexJob embeds a job (title + description) into the shared job corpus.
// Errors are logged, never returned.
func (ix *EmbeddingIndexer) IndexJob(ctx context.Context, jobID, title, description string) {
	if ix == nil || ix.generator == nil || ix.store == nil {
		return
	}
	if jobID == "" || title == "" {
		return
	}

	text := title
	if description != "" {
		text += "\n" + description
	}
	vector, err := ix.generator.Generate(ctx, text)
	if err != nil {
		log.Printf("WARNING: failed to generate job embedding: %v", err)
		return
	}
	if _, err := ix.store.Upsert(ctx, &repositories.Embedding{
		SourceType: repositories.EmbeddingSourceJob,
		SourceID:   jobID,
		Vector:     vector,
	}); err != nil {
		log.Printf("WARNING: failed to store job embedding: %v", err)
	}
}

// RemoveResume deletes a resume's embedding when the resume is deleted or
// replaced so no stale vector lingers.
func (ix *EmbeddingIndexer) RemoveResume(ctx context.Context, resumeID string) {
	if ix == nil || ix.store == nil || resumeID == "" {
		return
	}
	if err := ix.store.DeleteBySource(ctx, repositories.EmbeddingSourceResume, resumeID); err != nil {
		log.Printf("WARNING: failed to delete resume embedding: %v", err)
	}
}
