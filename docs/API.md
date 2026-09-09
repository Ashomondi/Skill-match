# SkillMatch API

Base URL: `{VITE_API_BASE_URL}` / the backend's `PORT`. All responses are JSON
unless noted. Auth-protected endpoints require an `Authorization: Bearer
<token>` header obtained from login/register.

## Health

### `GET /health`

Pings PostgreSQL and local file storage.

| Status | Body |
| ------ | ---- |
| 200    | `{"status":"healthy","dependencies":{"database":{"status":"ok"},"storage":{"status":"ok"}}}` |
| 503    | `{"status":"degraded","dependencies":{...}}` when a dependency is down or unconfigured |

## Authentication

### `POST /api/auth/register`

Create an account (auto-login: returns a token).

Request:
```json
{ "email": "jane@example.com", "password": "password123", "fullName": "Jane Doe" }
```

| Status | Meaning |
| ------ | ------- |
| 201    | `{ "message", "user": { id, email, fullName, is_active, created_at, updated_at }, "token" }` |
| 400    | invalid email / password < 8 chars / bad body |
| 409    | email already registered |

### `POST /api/auth/login`

Request:
```json
{ "email": "jane@example.com", "password": "password123" }
```

| Status | Meaning |
| ------ | ------- |
| 200    | `{ "message", "user", "token" }` |
| 401    | invalid email or password |

`user` never contains a password hash.

## Resumes (auth required)

The `resumes` table stores metadata; the file bytes live on the backend's
local filesystem and are served from `/storage`.

### `GET /api/resumes`

List the authenticated user's resumes, most recent first.

| Status | Body |
| ------ | ---- |
| 200    | `{ "resumes": [ { id, name, filename, size, status, uploadedAt } ] }` |
| 401    | missing/invalid token |

`status` is one of `uploaded | parsing | parsed | failed`.

### `POST /api/resumes`

Upload a resume (`multipart/form-data`). Fields:

- `resume` — the file part (`application/pdf`, `application/msword`,
  `application/vnd.openxmlformats-officedocument.wordprocessingml.document`,
  `text/plain`; max 5 MB).
- `replaceId` (optional) — id of an existing resume owned by the user; that
  resume and its object are replaced.

| Status | Meaning |
| ------ | ------- |
| 201    | the created resume object (same shape as list items) |
| 400    | missing/invalid file, unsupported type, file too large |
| 401    | missing/invalid token |
| 403    | `replaceId` belongs to another user |
| 404    | `replaceId` does not exist |

### `GET /api/resumes/{id}`

Fetch one of the user's resumes plus a download URL for the stored file.

| Status | Body |
| ------ | ---- |
| 200    | resume object with `url` (local file URL under `/storage`) |
| 401    | missing/invalid token |
| 403    | resume belongs to another user |
| 404    | resume not found |

### `DELETE /api/resumes/{id}`

Delete the user's resume: removes the stored file and the DB row.

| Status | Meaning |
| ------ | ------- |
| 204    | deleted |
| 401    | missing/invalid token |
| 403    | resume belongs to another user |
| 404    | resume not found |

## Chat (auth required)

### `POST /api/chat`

Send a message to the AI assistant. Only registered when `GEMINI_API_KEY` is
set.

Request:
```json
{ "message": "How should I frame my backend experience?" }
```

| Status | Body |
| ------ | ---- |
| 200    | `{ "message": "<assistant reply>" }` |
| 400    | missing/invalid message |
| 401    | missing/invalid token |
| 503    | AI/upstream unavailable |

## CV tailoring (auth required)

### `POST /api/tailor`

Generate a CV tailored to a job from the user's resume. Only registered when
`GEMINI_API_KEY` is set.

Request:
```json
{
  "resume_id": "00000000-0000-0000-0000-000000000001",
  "job_title": "Backend Engineer",
  "company": "Acme Corp",
  "job_description": "…",
  "current_content": ""
}
```

| Status | Body |
| ------ | ---- |
| 200    | `{ "content": "<tailored CV>" }` |
| 400    | missing `resume_id` / `job_title` |
| 401    | missing/invalid token |
| 404    | resume not found |
| 503    | AI/upstream unavailable |

## Errors

Error responses are `{ "error": "<message>" }`. Handlers return 400 for
validation problems and 500 for unexpected failures.

## Implemented endpoints

Health, auth (register/login), resume management (list/upload/get/delete),
job search, recommendations, saved jobs, applications (list/create/update),
chat, and CV tailoring are implemented and exposed.

