CREATE TABLE IF NOT EXISTS job_interactions (
    id               UUID         NOT NULL DEFAULT gen_random_uuid(),
    user_id          UUID         NOT NULL,
    job_id           UUID         NOT NULL,
    interaction_type STRING       NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT job_interactions_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT job_interactions_job_fk FOREIGN KEY (job_id) REFERENCES jobs (id) ON DELETE CASCADE,
    CONSTRAINT job_interactions_type_chk CHECK (interaction_type IN ('viewed', 'saved', 'applied', 'dismissed'))
);

CREATE INDEX IF NOT EXISTS job_interactions_user_id_idx ON job_interactions (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS job_interactions_user_job_type_unique_idx
    ON job_interactions (user_id, job_id, interaction_type);
