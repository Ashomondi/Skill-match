-- migrations/006_applications.sql
-- Issue 66: Application Tracking Backend
-- Schema for job applications and their status transitions.

CREATE TABLE IF NOT EXISTS applications (
    id          UUID            NOT NULL DEFAULT gen_random_uuid(),
    user_id     UUID            NOT NULL,
    job_id      UUID,
    status      STRING          NOT NULL DEFAULT 'applied',
    applied_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT applications_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT applications_job_fk  FOREIGN KEY (job_id)  REFERENCES jobs (id)  ON DELETE SET NULL,
    CONSTRAINT applications_status_chk CHECK (status IN ('applied', 'screening', 'interview', 'offer', 'rejected', 'withdrawn'))
);

-- Application tracking pages filter by user and group by status.
CREATE INDEX IF NOT EXISTS applications_user_status_idx ON applications (user_id, status);

-- Dashboard shows a user's most recently updated applications.
CREATE INDEX IF NOT EXISTS applications_user_updated_idx ON applications (user_id, updated_at DESC);

COMMENT ON TABLE applications IS 'Job applications tracked by a user, with status transitions.';
COMMENT ON COLUMN applications.status IS 'applied | screening | interview | offer | rejected | withdrawn';
