-- migrations/015_cv_documents.up.sql
-- SM-01: structured CV documents.
--
-- Resumes used to be opaque uploaded files with a single parsed_text blob.
-- Structured CVs let the tailoring flow render templates, reorder sections,
-- and run targeted ATS checks without re-parsing the original file.

CREATE TABLE IF NOT EXISTS cv_documents (
    id              UUID            NOT NULL DEFAULT gen_random_uuid(),
    user_id         UUID            NOT NULL,
    title           TEXT            NOT NULL,
    target_role     TEXT            NOT NULL DEFAULT '',
    template_id     TEXT            NOT NULL DEFAULT '',
    theme_config    JSONB           NOT NULL DEFAULT '{}'::jsonb,
    is_primary      BOOLEAN         NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    PRIMARY KEY (id),
    CONSTRAINT cv_documents_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT cv_documents_title_not_blank_chk CHECK (btrim(title) <> '')
);

CREATE TABLE IF NOT EXISTS cv_sections (
    id              UUID            NOT NULL DEFAULT gen_random_uuid(),
    cv_id           UUID            NOT NULL,
    section_type    TEXT            NOT NULL,
    title           TEXT            NOT NULL,
    sort_order      INTEGER         NOT NULL DEFAULT 0,
    content         JSONB           NOT NULL DEFAULT '{}'::jsonb,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),

    PRIMARY KEY (id),
    CONSTRAINT cv_sections_cv_fk FOREIGN KEY (cv_id) REFERENCES cv_documents (id) ON DELETE CASCADE,
    CONSTRAINT cv_sections_type_chk CHECK (section_type IN ('summary', 'experience', 'education', 'skills', 'projects', 'custom')),
    CONSTRAINT cv_sections_title_not_blank_chk CHECK (btrim(title) <> '')
);

-- Document lists are always scoped to a user.
CREATE INDEX IF NOT EXISTS cv_documents_user_id_idx ON cv_documents (user_id);

-- Supports "most recently edited CV first" listings.
CREATE INDEX IF NOT EXISTS cv_documents_user_updated_at_idx
    ON cv_documents (user_id, updated_at DESC);

-- At most one primary CV per user. The repository clears the previous primary
-- in the same transaction, so normal writes never hit this constraint.
CREATE UNIQUE INDEX IF NOT EXISTS cv_documents_one_primary_per_user_idx
    ON cv_documents (user_id) WHERE is_primary;

-- Sections are rendered in order for a single CV.
CREATE INDEX IF NOT EXISTS cv_sections_cv_sort_order_idx
    ON cv_sections (cv_id, sort_order, created_at);

COMMENT ON TABLE cv_documents IS 'Structured CV documents owned by a user.';
COMMENT ON COLUMN cv_documents.theme_config IS 'Template/theme options rendered by the frontend.';
COMMENT ON COLUMN cv_documents.is_primary IS 'At most one primary CV per user.';
COMMENT ON TABLE cv_sections IS 'Ordered, typed sections belonging to a structured CV.';
COMMENT ON COLUMN cv_sections.content IS 'Section payload; shape depends on section_type.';
