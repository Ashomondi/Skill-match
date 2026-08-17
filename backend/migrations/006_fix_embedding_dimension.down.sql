DROP INDEX IF EXISTS embeddings_vector_idx;
ALTER TABLE embeddings ALTER COLUMN vector TYPE VECTOR(1536);
CREATE VECTOR INDEX IF NOT EXISTS embeddings_vector_idx ON embeddings (vector vector_cosine_ops);
