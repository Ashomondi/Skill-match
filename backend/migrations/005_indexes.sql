-- migrations/005_indexes.sql
-- Issue 27: Database Optimization
-- Query-backed indexes for the hot paths introduced in 004_jobs.sql
-- (search, saved-jobs listing, and interaction tracking).

-- Keyword search is typically title + company filtering.
CREATE INDEX IF NOT EXISTS jobs_title_company_idx ON jobs (title, company);

-- Discovery lists active jobs newest-first.
CREATE INDEX IF NOT EXISTS jobs_active_posted_at_idx ON jobs (is_active, posted_at DESC);

-- Location and seniority filters used by the search UI.
CREATE INDEX IF NOT EXISTS jobs_location_idx ON jobs (location);
CREATE INDEX IF NOT EXISTS jobs_seniority_idx ON jobs (seniority);

-- Saved jobs are always read back per user, most recently saved first.
CREATE INDEX IF NOT EXISTS saved_jobs_user_saved_at_idx ON saved_jobs (user_id, saved_at DESC);

-- Lookup by job (e.g. "is this job saved?").
CREATE INDEX IF NOT EXISTS saved_jobs_job_id_idx ON saved_jobs (job_id);

-- Interaction history is read back per user for recommendation signals.
CREATE INDEX IF NOT EXISTS job_interactions_user_created_idx ON job_interactions (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS job_interactions_user_job_idx ON job_interactions (user_id, job_id);
