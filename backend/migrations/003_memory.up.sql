-- migrations/003_memory.sql
-- Issue 12: Memory Layer
-- Schema for conversation history and vector embeddings powering AI memory
-- (Sprint 3) and semantic job matching (Sprint 4, Issue 18).

-- Vector columns rely on the pgvector extension (available on PostgreSQL 13+
-- via `CREATE EXTENSION vector`). For local dev use the
-- pgvector/pgvector:pg16 container image.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_available_extensions WHERE name = 'vector') THEN
        CREATE EXTENSION IF NOT EXISTS vector;

        CREATE TABLE IF NOT EXISTS conversations (
            id          UUID            NOT NULL DEFAULT gen_random_uuid(),
            user_id     UUID            NOT NULL,
            role        TEXT            NOT NULL,
            content     TEXT            NOT NULL,
            created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),

            PRIMARY KEY (id),
            CONSTRAINT conversations_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
            CONSTRAINT conversations_role_chk CHECK (role IN ('user', 'assistant', 'system'))
        );

        -- Chat history is always read back per-user, in order. This is the hot
        -- path for "load conversation history" on every chat request.
        CREATE INDEX IF NOT EXISTS conversations_user_id_created_at_idx
            ON conversations (user_id, created_at DESC);

        COMMENT ON TABLE conversations IS 'Chat turn history for AI memory context.';
        COMMENT ON COLUMN conversations.role IS 'user | assistant | system, mirrors Bedrock message roles.';

        -- Embeddings are polymorphic: a vector can represent a resume, a
        -- conversation turn, or (later) a job description. source_type/source_id
        -- identify the origin row without a hard FK, since the source table
        -- varies.
        --
        -- Vector dimension is fixed at 1536 to match Amazon Titan Embeddings V2's
        -- default output size. If services/ai.go (Issue 11, Ashley) uses a
        -- different Bedrock embedding model, this column definition must change
        -- to match before any rows are inserted — pgvector columns are
        -- fixed-dimension.
        CREATE TABLE IF NOT EXISTS embeddings (
            id          UUID            NOT NULL DEFAULT gen_random_uuid(),
            user_id     UUID            NOT NULL,
            source_type TEXT            NOT NULL,
            source_id   UUID            NOT NULL,
            vector      VECTOR(1536)    NOT NULL,
            created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),

            PRIMARY KEY (id),
            CONSTRAINT embeddings_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
            CONSTRAINT embeddings_source_type_chk CHECK (source_type IN ('resume', 'conversation', 'job'))
        );

        -- One embedding per source row is the expected shape (re-embedding
        -- replaces, it doesn't accumulate). Enforced here rather than left to
        -- application discipline.
        CREATE UNIQUE INDEX IF NOT EXISTS embeddings_source_unique_idx
            ON embeddings (source_type, source_id);

        CREATE INDEX IF NOT EXISTS embeddings_user_id_idx ON embeddings (user_id);

        -- pgvector HNSW index for approximate nearest-neighbor search. This is
        -- what services/matching.go (Evans, Issue 18) and services/memory.go
        -- (Evans, Issue 13) query against. cosine distance matches Titan
        -- embeddings' intended similarity metric.
        CREATE INDEX IF NOT EXISTS embeddings_vector_idx
            ON embeddings USING hnsw (vector vector_cosine_ops);

        COMMENT ON TABLE embeddings IS 'Vector embeddings for resumes, conversations, and jobs; polymorphic via source_type/source_id.';
        COMMENT ON COLUMN embeddings.vector IS 'Fixed at 1536 dims (Titan Embeddings V2). Change requires a new migration + backfill if the model changes.';
    ELSE
        RAISE NOTICE 'pgvector extension is not installed; skipping memory schema migration';
    END IF;
END $$;
