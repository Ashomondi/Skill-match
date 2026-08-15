CREATE TABLE IF NOT EXISTS jobs (
    id           UUID         NOT NULL DEFAULT gen_random_uuid(),
    external_id  STRING       NOT NULL,
    title        STRING       NOT NULL,
    company      STRING       NOT NULL,
    location     STRING       NOT NULL DEFAULT '',
    description  STRING       NOT NULL DEFAULT '',
    salary       STRING       NOT NULL DEFAULT '',
    remote       BOOL         NOT NULL DEFAULT false,
    source_url   STRING       NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT jobs_external_id_unique UNIQUE (external_id)
);