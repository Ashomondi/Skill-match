-- Google sign-in support.
--
-- Google accounts are identified by the stable `sub` claim from the verified
-- ID token. Password accounts keep their password_hash; Google-only accounts
-- have no password, so the column must allow NULL.

ALTER TABLE users ADD COLUMN IF NOT EXISTS google_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS users_google_id_unique
    ON users (google_id)
    WHERE google_id IS NOT NULL;

ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

COMMENT ON COLUMN users.google_id IS
    'Google account subject (sub) for sign-in with Google; NULL for password accounts.';
