DROP INDEX IF EXISTS embeddings_vector_idx;
ALTER TABLE embeddings ALTER COLUMN vector TYPE VECTOR(1536);
CREATE INDEX IF NOT EXISTS embeddings_vector_idx
    ON embeddings USING hnsw (vector vector_cosine_ops);
