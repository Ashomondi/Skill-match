# SkillMatch API — AI & Chat

## POST /chat

Sends a user message to the AI agent and returns a response generated via
Amazon Bedrock, with context assembled from CockroachDB (conversation
history and, optionally, parsed resume text).

### Request

```json
{
  "user_id": "string, required",
  "message": "string, required, max 4000 characters",
  "resume_id": "string, optional"
}
```

| Field | Type | Required | Notes |
|---|---|---|---|
| user_id | string | yes | Authenticated user's ID |
| message | string | yes | 1–4000 characters |
| resume_id | string | no | Includes parsed resume text as context if provided; rejected if the resume does not belong to user_id |

### Response — success (200)

```json
{
  "message": "string"
}
```

### Response — error

```json
{
  "error": "string, safe for display to the user",
  "code": "validation | storage | database | auth | internal"
}
```

| Code | Meaning | Typical status |
|---|---|---|
| validation | Empty/oversized message, missing fields | 400 |
| database | CockroachDB read/write failure | 500 |
| internal | Bedrock failure (throttling, timeout, access, or unclassified) | 502 |
| auth | Resume ownership mismatch or missing authentication | 401/403 |

## Error handling model

All AI/chat errors flow through the shared `AppError` type
(`backend/utils/errors.go`), which separates two things:

- **UserMsg** — safe, non-technical text returned in the `error` field above.
- **Err** — the real underlying error (e.g. raw Bedrock/AWS error), logged
  server-side with context (user ID, category) but never included in the
  HTTP response.

This means every endpoint using `utils.WriteError` returns errors in this
same shape — the frontend can rely on one consistent error format across
the whole API, not just `/chat`.

## Bedrock error classification

Bedrock-specific failures are mapped to safe messages before reaching the
client (`backend/clients/bedrock.go`, `ClassifyBedrockError`):

| Bedrock error | User-facing message |
|---|---|
| ThrottlingException | "The AI service is busy right now. Please try again in a moment." |
| ModelTimeoutException | "The AI took too long to respond. Please try again." |
| ValidationException | "There was a problem with the request format." |
| AccessDeniedException | "AI service access is not configured correctly." |
| (unclassified) | "The AI service encountered an error. Please try again." |