-- Corrects embeddings.vector from VECTOR(1536) to VECTOR(1024) to match
-- Titan Text Embeddings V2's actual max output size (1536 was mistakenly
-- copied from the older Titan G1 model). Table is empty — safe to drop
-- and recreate the column rather than attempt an in-place type change.

DROP INDEX IF EXISTS embeddings_vector_idx;
ALTER TABLE embeddings DROP COLUMN vector;
ALTER TABLE embeddings ADD COLUMN vector VECTOR(1024) NOT NULL;
CREATE INDEX IF NOT EXISTS embeddings_vector_idx
    ON embeddings USING hnsw (vector vector_cosine_ops);
