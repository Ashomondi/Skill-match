-- Adds the seniority and work_type columns to the jobs table. Migration 004
-- was edited (commit 2703747) to create them, but that file had already been
-- applied to existing databases, so this forward migration backfills them
-- idempotently on databases that predate the edit. Safe to skip on databases
-- created after the edit.

ALTER TABLE jobs ADD COLUMN IF NOT EXISTS seniority TEXT NOT NULL DEFAULT '';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS work_type TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_jobs_search
    ON jobs (title, company, location, remote, seniority, work_type);
