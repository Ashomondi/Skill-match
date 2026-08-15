-- migrations/002_resume.sql
-- Issue 7: Resume Repository
-- Schema for resume metadata. Binary content lives in S3 (Issue 8); this
-- table stores pointers and processing state only.

CREATE TABLE IF NOT EXISTS resumes (
    id                  UUID            NOT NULL DEFAULT gen_random_uuid(),
    user_id             UUID            NOT NULL,
    s3_key              STRING          NOT NULL,
    original_filename   STRING          NOT NULL,
    content_type        STRING          NOT NULL,
    file_size_bytes      INT8            NOT NULL,
    status              STRING          NOT NULL DEFAULT 'uploaded',
    parsed_text         STRING,
    failure_reason       STRING,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT resumes_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT resumes_status_chk CHECK (status IN ('uploaded', 'parsing', 'parsed', 'failed')),
    CONSTRAINT resumes_file_size_positive_chk CHECK (file_size_bytes > 0)
);

-- Resume list/detail views are always scoped to a user; index the FK
-- explicitly since CockroachDB does not auto-index foreign keys the way
-- some engines do for the referencing column alone in composite cases.
CREATE INDEX IF NOT EXISTS resumes_user_id_idx ON resumes (user_id);

-- Supports "most recent resume per user" without a full table scan.
CREATE INDEX IF NOT EXISTS resumes_user_id_created_at_idx
    ON resumes (user_id, created_at DESC);

COMMENT ON TABLE resumes IS 'Uploaded resume metadata; file bytes live in S3.';
COMMENT ON COLUMN resumes.s3_key IS 'Object key in the resumes S3 bucket, not a full URL.';
COMMENT ON COLUMN resumes.status IS 'uploaded -> parsing -> parsed | failed';