CREATE TABLE IF NOT EXISTS applications (
    id UUID NOT NULL DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE, status TEXT NOT NULL DEFAULT 'saved',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT applications_primary PRIMARY KEY (id), CONSTRAINT applications_user_job_unique UNIQUE (user_id, job_id),
    CONSTRAINT applications_status_chk CHECK (status IN ('saved','applied','screening','interview','offer','rejected','withdrawn'))
);
CREATE TABLE IF NOT EXISTS application_status_history (
    id UUID NOT NULL DEFAULT gen_random_uuid(), application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    status TEXT NOT NULL, changed_at TIMESTAMPTZ NOT NULL DEFAULT now(), CONSTRAINT application_status_history_primary PRIMARY KEY (id),
    CONSTRAINT application_status_history_status_chk CHECK (status IN ('saved','applied','screening','interview','offer','rejected','withdrawn'))
);
CREATE INDEX IF NOT EXISTS application_status_history_application_idx ON application_status_history(application_id, changed_at);
