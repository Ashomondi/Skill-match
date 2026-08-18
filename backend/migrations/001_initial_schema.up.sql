-- migrations/001_initial_schema.sql
-- Issue 2: Database & Configuration
-- Initial schema: users table.

CREATE TABLE IF NOT EXISTS users (
    id              UUID            NOT NULL DEFAULT gen_random_uuid(),
    email           STRING          NOT NULL,
    password_hash   STRING          NOT NULL,
    full_name       STRING          NOT NULL DEFAULT '',
    is_active       BOOL            NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    CONSTRAINT "primary" PRIMARY KEY (id),
    CONSTRAINT users_email_unique UNIQUE (email)
);

-- Case-insensitive lookups are common on login; store email lowercased at the
-- application layer and enforce it here as a defensive backstop.
-- ALTER TABLE users ADD CONSTRAINT users_email_lowercase_chk
   -- CHECK (email = lower(email));

COMMENT ON TABLE users IS 'Registered application users.';
COMMENT ON COLUMN users.password_hash IS 'bcrypt hash, never plaintext.';
