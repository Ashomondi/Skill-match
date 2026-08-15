-- Query-performance indexes for the production read paths.
-- Keep these separate so they can be applied independently and measured.

CREATE INDEX IF NOT EXISTS jobs_created_at_id_idx
    ON jobs (created_at DESC, id);

CREATE INDEX IF NOT EXISTS applications_user_updated_at_idx
    ON applications (user_id, updated_at DESC, id);

CREATE INDEX IF NOT EXISTS applications_job_user_idx
    ON applications (job_id, user_id);

-- The unique (user_id, job_id, interaction_type) index already serves the
-- user-scoped interaction lookup; the older prefix index is redundant.
DROP INDEX IF EXISTS job_interactions_user_id_idx;
