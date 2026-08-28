# SkillMatch — Data Model

This document describes the persistent storage layout: the PostgreSQL schema
and the S3 (MinIO) object store. Migrations live in `backend/migrations/*.sql`
and are applied automatically at server startup (see
[`backend/migrations/runner.go`](../backend/migrations/runner.go)); applied
versions are tracked in the `schema_migrations` table.

## 1. Storage split

| Concern            | Where it lives           |
| ------------------ | ------------------------ |
| Structured data    | PostgreSQL (SQL)         |
| Binary file bytes  | S3 / MinIO               |
| Vector embeddings  | PostgreSQL (`embeddings`, pgvector) |

PostgreSQL is the system of record ("agentic memory"). S3 only stores the raw
uploaded resume bytes; the `resumes` table keeps a pointer (`s3_key`) to each
object. No binary file content is ever written to PostgreSQL.

## 2. PostgreSQL schema

All tables are in the `public` schema of the application database. Foreign keys
cascade on user deletion unless noted.

### 2.1 `users` (migration 001)

| Column        | Type      | Notes                                 |
| ------------- | --------- | ------------------------------------- |
| `id`          | UUID      | PK, default `gen_random_uuid()`       |
| `email`       | TEXT    | UNIQUE, stored lowercased (CHECK)     |
| `password_hash` | TEXT  | bcrypt hash, never returned to clients |
| `full_name`   | TEXT    | default `''`                          |
| `is_active`   | BOOLEAN      | default `true`; soft-disable          |
| `created_at`  | TIMESTAMPTZ | default `now()`                     |
| `updated_at`  | TIMESTAMPTZ | default `now()`                     |

### 2.2 `resumes` (migration 002)

Resume metadata only — bytes live in S3.

| Column              | Type       | Notes                              |
| ------------------- | ---------- | ---------------------------------- |
| `id`                | UUID       | PK                                 |
| `user_id`           | UUID       | FK → `users.id` ON DELETE CASCADE  |
| `s3_key`            | TEXT     | object key in the resumes bucket   |
| `original_filename` | TEXT     |                                    |
| `content_type`      | TEXT     |                                    |
| `file_size_bytes`   | BIGINT       | CHECK `> 0`                        |
| `status`            | TEXT     | CHECK: `uploaded|parsing|parsed|failed` |
| `parsed_text`       | TEXT     | nullable (AI parsing output)       |
| `failure_reason`    | TEXT     | nullable                           |
| `created_at` / `updated_at` | TIMESTAMPTZ | |

Indexes: `(user_id)`, `(user_id, created_at DESC)`.

### 2.3 `conversations` (migration 003)

One row per chat turn (append-only).

| Column       | Type       | Notes                                    |
| ------------ | ---------- | ---------------------------------------- |
| `id`         | UUID       | PK                                       |
| `user_id`    | UUID       | FK → `users.id` ON DELETE CASCADE        |
| `role`       | TEXT     | CHECK: `user|assistant|system`           |
| `content`    | TEXT     | the turn text                            |
| `created_at` | TIMESTAMPTZ |                                          |

Index: `(user_id, created_at DESC)` — the hot path for loading chat history.

### 2.4 `embeddings` (migration 003)

Polymorphic vector store. One row per `(source_type, source_id)`.

| Column        | Type        | Notes                                        |
| ------------- | ----------- | -------------------------------------------- |
| `id`          | UUID        | PK                                           |
| `user_id`     | UUID        | FK → `users.id` ON DELETE CASCADE            |
| `source_type` | TEXT      | CHECK: `resume|conversation|job`             |
| `source_id`   | UUID        | id of the source row (no hard FK — varies)   |
| `vector`      | VECTOR(1536)| fixed dimension (Titan Embeddings V2)        |
| `created_at`  | TIMESTAMPTZ |                                              |

Unique index `(source_type, source_id)`; user index `(user_id)`; **vector
index** `embeddings_vector_idx (vector vector_cosine_ops)` — a pgvector
HNSW index backing ANN similarity search (requires the `vector` extension,
enabled by migration 003).

### 2.5 `jobs` (migration 004)

| Column        | Type       | Notes                                          |
| ------------- | ---------- | ---------------------------------------------- |
| `id`          | UUID       | PK                                             |
| `title`       | TEXT     |                                                |
| `company`     | TEXT     |                                                |
| `location`    | TEXT     |                                                |
| `work_type`   | TEXT     | CHECK: `full-time|part-time|contract|internship` |
| `seniority`   | TEXT     | CHECK: `''|entry|mid|senior|lead`              |
| `salary`      | TEXT     |                                                |
| `description` | TEXT     |                                                |
| `posted_at`   | TIMESTAMPTZ |                                              |
| `source`      | TEXT     | default `manual`                               |
| `is_active`   | BOOLEAN       | default `true`                                 |
| `created_at` / `updated_at` | TIMESTAMPTZ |                             |

### 2.6 `saved_jobs` (migration 004)

| Column     | Type   | Notes                                    |
| ---------- | ------ | ---------------------------------------- |
| `id`       | UUID   | PK                                       |
| `user_id`  | UUID   | FK → `users.id` ON DELETE CASCADE        |
| `job_id`   | UUID   | FK → `jobs.id` ON DELETE CASCADE         |
| `saved_at` | TIMESTAMPTZ |                                    |

UNIQUE `(user_id, job_id)`; index `(user_id, saved_at DESC)`.

### 2.7 `job_interactions` (migration 004)

Behavioral signal for recommendations.

| Column             | Type       | Notes                                   |
| ------------------ | ---------- | --------------------------------------- |
| `id`               | UUID       | PK                                      |
| `user_id`          | UUID       | FK → `users.id` ON DELETE CASCADE       |
| `job_id`           | UUID       | FK → `jobs.id` ON DELETE CASCADE        |
| `interaction_type` | TEXT     | CHECK: `view|save|apply|dismiss|search` |
| `created_at`       | TIMESTAMPTZ |                                         |

Indexes: `(user_id, created_at DESC)`, `(user_id, job_id)`.

### 2.8 `applications` (migration 006)

| Column       | Type       | Notes                                          |
| ------------ | ---------- | ---------------------------------------------- |
| `id`         | UUID       | PK                                             |
| `user_id`    | UUID       | FK → `users.id` ON DELETE CASCADE              |
| `job_id`     | UUID       | nullable; FK → `jobs.id` ON DELETE SET NULL    |
| `status`     | TEXT     | CHECK: `applied|screening|interview|offer|rejected|withdrawn` |
| `applied_at` | TIMESTAMPTZ |                                             |
| `created_at` / `updated_at` | TIMESTAMPTZ |                             |

Indexes: `(user_id, status)`, `(user_id, updated_at DESC)`.

### 2.9 `schema_migrations` (runner-managed)

| Column       | Type       | Notes                              |
| ------------ | ---------- | ---------------------------------- |
| `version`    | TEXT     | PK — the migration filename        |
| `applied_at` | TIMESTAMPTZ |                                    |

## 3. S3 / MinIO object store

Bucket: configured via `S3_BUCKET_NAME` (e.g. `initone`).

Key layout:

```
resumes/{userID}/{fileID}
```

- `userID` — the owning user's UUID.
- `fileID` — a hex-encoded random id plus the original extension
  (`utils.GenerateFileID`), e.g. `af5e098b75f1db93cc20dea2122b6478.pdf`.

Examples:

```
resumes/e147a448-88d5-4d58-8cac-1a780ab04757/af5e098b75f1db93cc20dea2122b6478.pdf
```

Access is via short-lived presigned URLs (uploads/downloads) generated by
`clients/s3.go`; objects are never made publicly readable. Deleting a resume
removes both the DB row and the object at its `s3_key`.

## 4. Relationship summary

```
users 1───N resumes  (bytes in S3, pointer via resumes.s3_key)
users 1───N conversations
users 1───N embeddings   (polymorphic: resume | conversation | job)
users 1───N saved_jobs N───1 jobs
users 1───N job_interactions N───1 jobs
users 1───N applications N───0..1 jobs (SET NULL on job delete)
```
