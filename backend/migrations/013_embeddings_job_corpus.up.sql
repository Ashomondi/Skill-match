-- Relaxes embeddings.user_id so that the global job corpus can be stored
-- without an owner.
--
-- Embeddings are polymorphic via (source_type, source_id):
--   * resume / conversation rows belong to a single user (user_id NOT NULL)
--   * job rows form a SHARED corpus that is matched against any user's
--     query vector (see EmbeddingRepository.FindSimilarJobs), so they have
--     no meaningful owner.
--
-- Making the column nullable lets job rows store NULL while user-owned rows
-- keep the FK. The per-source unique index already guarantees one embedding
-- per job, so the shared corpus stays deduplicated.
--
-- Idempotent: safe to re-run if a previous attempt failed partway.

DO $$
BEGIN
    -- Drop NOT NULL if still set (idempotent no-op once relaxed).
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'embeddings'
          AND column_name = 'user_id'
          AND is_nullable = 'NO'
    ) THEN
        ALTER TABLE embeddings ALTER COLUMN user_id DROP NOT NULL;
    END IF;

    -- Drop the guard constraint if present, then re-add only when missing.
    ALTER TABLE embeddings DROP CONSTRAINT IF EXISTS embeddings_user_id_required_for_user_sources;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'embeddings_user_id_required_for_user_sources'
    ) THEN
        ALTER TABLE embeddings ADD CONSTRAINT embeddings_user_id_required_for_user_sources
            CHECK (user_id IS NOT NULL OR source_type = 'job');
    END IF;
END $$;
