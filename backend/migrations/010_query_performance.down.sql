DROP INDEX IF EXISTS applications_job_user_idx;
DROP INDEX IF EXISTS applications_user_updated_at_idx;
DROP INDEX IF EXISTS jobs_created_at_id_idx;

CREATE INDEX IF NOT EXISTS job_interactions_user_id_idx
    ON job_interactions (user_id);
