-- Roll back Google sign-in support.
--
-- Google-only accounts (no password) cannot satisfy the original NOT NULL
-- constraint, so remove them before restoring it. Password accounts that were
-- linked to Google keep their password and remain usable.

DELETE FROM users WHERE google_id IS NOT NULL AND password_hash IS NULL;

DROP INDEX IF EXISTS users_google_id_unique;

ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;

ALTER TABLE users DROP COLUMN IF EXISTS google_id;
