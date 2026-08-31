CREATE TABLE IF NOT EXISTS jobs (
    id           UUID         NOT NULL DEFAULT gen_random_uuid(),
    external_id  TEXT         NOT NULL,
    title        TEXT         NOT NULL,
    company      TEXT         NOT NULL,
    location     TEXT         NOT NULL DEFAULT '',
    description  TEXT         NOT NULL DEFAULT '',
    salary       TEXT         NOT NULL DEFAULT '',
    remote       BOOLEAN      NOT NULL DEFAULT false,
    seniority    TEXT         NOT NULL DEFAULT '',
    work_type    TEXT         NOT NULL DEFAULT '',
    source_url   TEXT         NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    PRIMARY KEY (id),
    CONSTRAINT jobs_external_id_unique UNIQUE (external_id)
);

CREATE INDEX IF NOT EXISTS idx_jobs_search ON jobs (title, company, location, remote, seniority, work_type);