-- Corrects embeddings.vector from VECTOR(1536) to VECTOR(1024) to match
-- Titan Text Embeddings V2's actual max output size (1536 was mistakenly
-- copied from the older Titan G1 model). Table is empty — safe to drop
-- and recreate the column rather than attempt an in-place type change.
-- This migration is idempotent: safe to re-run if a previous attempt
-- failed partway (e.g. crash after DROP COLUMN before version recorded).

DO $$
BEGIN
    -- Drop the HNSW index if it exists (idempotent).
    DROP INDEX IF EXISTS embeddings_vector_idx;

    -- Drop the vector column if it exists (idempotent).
    ALTER TABLE embeddings DROP COLUMN IF EXISTS vector;

    -- Add the column with the corrected dimension (idempotent: IF NOT EXISTS
    -- ensures it is added only when missing; if already present with the
    -- correct type it is a no-op, if present with the old type it is dropped
    -- and recreated below).
    ALTER TABLE embeddings ADD COLUMN IF NOT EXISTS vector VECTOR(1024) NOT NULL;

    -- Re-create the cosine similarity HNSW index (idempotent).
    CREATE INDEX IF NOT EXISTS embeddings_vector_idx
        ON embeddings USING hnsw (vector vector_cosine_ops);
END $$;