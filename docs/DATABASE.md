# SkillMatch — Data Model

This document describes the persistent storage layout: the CockroachDB schema
and the S3 (MinIO) object store. Migrations live in `backend/migrations/*.sql`
and are applied automatically at server startup (see
[`backend/migrations/runner.go`](../backend/migrations/runner.go)); applied
versions are tracked in the `schema_migrations` table.

## 1. Storage split

| Concern            | Where it lives           |
| ------------------ | ------------------------ |
| Structured data    | CockroachDB (SQL)        |
| Binary file bytes  | S3 / MinIO               |
| Vector embeddings  | CockroachDB (`embeddings`) |

CockroachDB is the system of record ("agentic memory"). S3 only stores the raw
uploaded resume bytes; the `resumes` table keeps a pointer (`s3_key`) to each
object. No binary file content is ever written to CockroachDB.

## 2. CockroachDB schema

All tables are in the `public` schema of the application database. Foreign keys
cascade on user deletion unless noted.

### 2.1 `users` (migration 001)

| Column        | Type      | Notes                                 |
| ------------- | --------- | ------------------------------------- |
| `id`          | UUID      | PK, default `gen_random_uuid()`       |
| `email`       | STRING    | UNIQUE, stored lowercased (CHECK)     |
| `password_hash` | STRING  | bcrypt hash, never returned to clients |
| `full_name`   | STRING    | default `''`                          |
| `is_active`   | BOOL      | default `true`; soft-disable          |
| `created_at`  | TIMESTAMPTZ | default `now()`                     |
| `updated_at`  | TIMESTAMPTZ | default `now()`                     |

### 2.2 `resumes` (migration 002)

Resume metadata only — bytes live in S3.

| Column              | Type       | Notes                              |
| ------------------- | ---------- | ---------------------------------- |
| `id`                | UUID       | PK                                 |
| `user_id`           | UUID       | FK → `users.id` ON DELETE CASCADE  |
| `s3_key`            | STRING     | object key in the resumes bucket   |
| `original_filename` | STRING     |                                    |
| `content_type`      | STRING     |                                    |
| `file_size_bytes`   | INT8       | CHECK `> 0`                        |
| `status`            | STRING     | CHECK: `uploaded|parsing|parsed|failed` |
| `parsed_text`       | STRING     | nullable (AI parsing output)       |
| `failure_reason`    | STRING     | nullable                           |
| `created_at` / `updated_at` | TIMESTAMPTZ | |

Indexes: `(user_id)`, `(user_id, created_at DESC)`.

### 2.3 `conversations` (migration 003)

One row per chat turn (append-only).

| Column       | Type       | Notes                                    |
| ------------ | ---------- | ---------------------------------------- |
| `id`         | UUID       | PK                                       |
| `user_id`    | UUID       | FK → `users.id` ON DELETE CASCADE        |
| `role`       | STRING     | CHECK: `user|assistant|system`           |
| `content`    | STRING     | the turn text                            |
| `created_at` | TIMESTAMPTZ |                                          |

Index: `(user_id, created_at DESC)` — the hot path for loading chat history.

### 2.4 `embeddings` (migration 003)

Polymorphic vector store. One row per `(source_type, source_id)`.

| Column        | Type        | Notes                                        |
| ------------- | ----------- | -------------------------------------------- |
| `id`          | UUID        | PK                                           |
| `user_id`     | UUID        | FK → `users.id` ON DELETE CASCADE            |
| `source_type` | STRING      | CHECK: `resume|conversation|job`             |
| `source_id`   | UUID        | id of the source row (no hard FK — varies)   |
| `vector`      | VECTOR(1536)| fixed dimension (Titan Embeddings V2)        |
| `created_at`  | TIMESTAMPTZ |                                              |

Unique index `(source_type, source_id)`; user index `(user_id)`; **vector
index** `embeddings_vector_idx (vector vector_cosine_ops)` — the Distributed
Vector Index backing ANN similarity search.

### 2.5 `jobs` (migration 004)

| Column        | Type       | Notes                                          |
| ------------- | ---------- | ---------------------------------------------- |
| `id`          | UUID       | PK                                             |
| `title`       | STRING     |                                                |
| `company`     | STRING     |                                                |
| `location`    | STRING     |                                                |
| `work_type`   | STRING     | CHECK: `full-time|part-time|contract|internship` |
| `seniority`   | STRING     | CHECK: `''|entry|mid|senior|lead`              |
| `salary`      | STRING     |                                                |
| `description` | STRING     |                                                |
| `posted_at`   | TIMESTAMPTZ |                                              |
| `source`      | STRING     | default `manual`                               |
| `is_active`   | BOOL       | default `true`                                 |
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
| `interaction_type` | STRING     | CHECK: `view|save|apply|dismiss|search` |
| `created_at`       | TIMESTAMPTZ |                                         |

Indexes: `(user_id, created_at DESC)`, `(user_id, job_id)`.

### 2.8 `applications` (migration 006)

| Column       | Type       | Notes                                          |
| ------------ | ---------- | ---------------------------------------------- |
| `id`         | UUID       | PK                                             |
| `user_id`    | UUID       | FK → `users.id` ON DELETE CASCADE              |
| `job_id`     | UUID       | nullable; FK → `jobs.id` ON DELETE SET NULL    |
| `status`     | STRING     | CHECK: `applied|screening|interview|offer|rejected|withdrawn` |
| `applied_at` | TIMESTAMPTZ |                                             |
| `created_at` / `updated_at` | TIMESTAMPTZ |                             |

Indexes: `(user_id, status)`, `(user_id, updated_at DESC)`.

### 2.9 `schema_migrations` (runner-managed)

| Column       | Type       | Notes                              |
| ------------ | ---------- | ---------------------------------- |
| `version`    | STRING     | PK — the migration filename        |
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
