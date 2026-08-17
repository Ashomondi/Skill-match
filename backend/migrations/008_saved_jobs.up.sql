CREATE TABLE IF NOT EXISTS saved_jobs (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT saved_jobs_user_job_unique UNIQUE (user_id, job_id)
);
CREATE INDEX IF NOT EXISTS saved_jobs_user_saved_at_idx ON saved_jobs (user_id, saved_at DESC);
