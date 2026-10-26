-- migrations/015_cv_documents.down.sql
-- cv_sections must be dropped first because it references cv_documents.
DROP TABLE IF EXISTS cv_sections;
DROP TABLE IF EXISTS cv_documents;
