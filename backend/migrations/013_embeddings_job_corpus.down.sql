-- Reverts 013_embeddings_job_corpus: job rows with NULL user_id must be
-- cleaned up first, then user_id is restored to NOT NULL and the guard
-- constraint removed.

DO $$
BEGIN
    ALTER TABLE embeddings DROP CONSTRAINT IF EXISTS embeddings_user_id_required_for_user_sources;

    DELETE FROM embeddings WHERE user_id IS NULL;

    ALTER TABLE embeddings ALTER COLUMN user_id SET NOT NULL;
END $$;
