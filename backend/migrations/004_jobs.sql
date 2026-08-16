-- migrations/004_jobs.sql
-- Issue 17: Job Repository + Issue 56: Job Interaction Tracking
-- Schema for job listings, saved jobs, and per-user job interactions
-- (the interaction data powers personalized recommendations).

CREATE TABLE IF NOT EXISTS jobs (
    id          UUID            NOT NULL DEFAULT gen_random_uuid(),
    title       STRING          NOT NULL,
    company     STRING          NOT NULL,
    location    STRING          NOT NULL DEFAULT '',
    work_type   STRING          NOT NULL DEFAULT 'full-time',
    seniority   STRING          NOT NULL DEFAULT '',
    salary      STRING          NOT NULL DEFAULT '',
    description STRING          NOT NULL DEFAULT '',
    posted_at   TIMESTAMPTZ     NOT NULL DEFAULT now(),
    source      STRING          NOT NULL DEFAULT 'manual',
    is_active   BOOL            NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT jobs_work_type_chk CHECK (work_type IN ('full-time', 'part-time', 'contract', 'internship')),
    CONSTRAINT jobs_seniority_chk CHECK (seniority IN ('', 'entry', 'mid', 'senior', 'lead'))
);

CREATE TABLE IF NOT EXISTS saved_jobs (
    id        UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id   UUID        NOT NULL,
    job_id    UUID        NOT NULL,
    saved_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT saved_jobs_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT saved_jobs_job_fk  FOREIGN KEY (job_id)  REFERENCES jobs (id)  ON DELETE CASCADE,
    CONSTRAINT saved_jobs_user_job_unique UNIQUE (user_id, job_id)
);

CREATE TABLE IF NOT EXISTS job_interactions (
    id               UUID        NOT NULL DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL,
    job_id           UUID        NOT NULL,
    interaction_type STRING      NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT job_interactions_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT job_interactions_job_fk  FOREIGN KEY (job_id)  REFERENCES jobs (id)  ON DELETE CASCADE,
    CONSTRAINT job_interactions_type_chk CHECK (interaction_type IN ('view', 'save', 'apply', 'dismiss', 'search'))
);

COMMENT ON TABLE jobs IS 'Job listings shown in discovery and search.';
COMMENT ON COLUMN jobs.work_type IS 'full-time | part-time | contract | internship';
COMMENT ON COLUMN jobs.seniority IS 'entry | mid | senior | lead (empty when unspecified)';
COMMENT ON TABLE saved_jobs IS 'Jobs a user has saved for later. One row per (user, job).';
COMMENT ON TABLE job_interactions IS 'Per-user interactions with jobs, used to inform recommendations.';
