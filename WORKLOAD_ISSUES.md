# SkillMatch — Remaining Workload Issues

Goal: everything below is the work still needed to take SkillMatch from its current
"partially wired, partially stubbed" state to a working product. Each issue is an
assignable unit of work (bug fix or feature) targeting production-quality code.

Assignee buckets:

| Area    | People       |
| ------- | ------------ |
| Backend | @sos, @ashley, @evans |
| Frontend| @emma, @ian  |

Conventions

- Follow existing style: `net/http` handlers → services → repositories, `AppError`
  for errors (`utils/errors.go`), never log credentials/query text, always run
  `gofmt` and `go test ./...` before finishing a backend issue.
- Frontend: run `npm run build` (runs `tsc`) before finishing.
- If an issue touches a migration, write both `.up.sql` and `.down.sql` and keep the
  runner's idempotency rules in mind (`backend/migrations/runner.go`).

---

## Backend — @sos

### Issue #1 — Implement the Applications repository (currently a stub)

**Type:** bug / feature · **Priority:** high

**Description:**
`repositories/application.go` is a stub: every method returns placeholder values and
never touches the database (e.g. `Create` returns `&models.Application{...}, nil`,
`ListByUserID` returns `nil, nil`). This means the applications API is non-functional
end-to-end — list/create/update all return empty or fake data, and nothing is persisted.

**Why it matters**
- The integration test `repositories/application_test.go` (build-tagged `integration`)
  expects real persistence and will FAIL against a live database.
- The dashboard, applications page, and tailoring "submit application" flow all depend
  on this data.

**Files**
- `backend/repositories/application.go`
- `backend/models/application.go`
- `backend/services/application.go`
- `backend/handlers/applications.go`
- `backend/repositories/application_test.go`
- `backend/migrations/009_applications.up.sql`

**Acceptance criteria**
- [ ] `Create` inserts into `applications` with default status `saved`; maps unique
      violation `(user_id, job_id)` to a conflict error and FK violation to a
      not-found error.
- [ ] `GetByID` is user-scoped (`WHERE id = $1 AND user_id = $2`) so users can't read
      another user's application.
- [ ] `UpdateStatus` updates the row, bumps `updated_at`, and records the change in
      `application_status_history`.
- [ ] `History` returns `application_status_history` rows ordered by `changed_at`.
- [ ] `ListByUserID` joins `jobs` so each application carries its `Job` details, ordered
      by most recent; returns an empty slice (not nil) when none exist.
- [ ] `Delete` is user-scoped; `ErrApplicationNotFound` when no row matched.
- [ ] Integration tests pass with `go test -tags integration ./repositories -run TestApplication`.

---

### Issue #2 — Align application status values across the stack

**Type:** bug · **Priority:** medium

**Description:**
`services/application.go` `isValidApplicationStatus` accepts `"interviewing"` and
`"accepted"`, which are NOT valid `models.ApplicationStatus` values (`interview`,
`offer`). The DB CHECK constraint in `009_applications` also only allows
`saved|applied|screening|interview|offer|rejected|withdrawn`. A status update with an
invalid value today silently passes validation in the service layer and then either
fails in the DB or stores a value the UI can't render.

**Files**
- `backend/services/application.go`
- `backend/models/application.go`
- `backend/handlers/applications.go`

**Acceptance criteria**
- [ ] Validation uses the same set as the model's `Valid()` method and the DB CHECK.
- [ ] Invalid status → `400` with a clear message (no raw DB error).

---

## Backend — @ashley

### Issue #3 — Implement real job search (filters are currently ignored)

**Type:** bug · **Priority:** high

**Description:**
`repositories/jobs.go` `Search()` just calls `List()` and throws away every filter
(query, location, company, remote, offset). The `GET /api/jobs/search` endpoint returns
all jobs regardless of query params, and the `pagination.total` is the page size, not
the real total. Also, the frontend sends `seniority` and `work_type` params that the
backend doesn't even model — the `jobs` table (`004_jobs.up.sql`) lacks those columns
while `docs/DATABASE.md` claims they exist.

**Files**
- `backend/repositories/jobs.go`
- `backend/services/jobs.go`
- `backend/handlers/jobs.go`
- `backend/migrations/004_jobs.up.sql` (+ new migration for `seniority`/`work_type`
  if you add them)

**Acceptance criteria**
- [ ] `Search` filters by title/company/description (ILIKE), location, company, and
      remote flag; supports `limit`/`offset`; returns a real `Total` count via a
      `COUNT(*)` query.
- [ ] Empty/nil filter returns all jobs (or an explicit "no filters" path) without an
      unbounded scan.
- [ ] Decide on `seniority`/`work_type`: add columns via a new migration and include
      them in the search filter, OR reject them explicitly. Schema must match
      `docs/DATABASE.md`.
- [ ] Unit tests for the filter SQL (via a fake pool or `pgxmock`-style approach that
      matches the repo's existing test style).

---

### Issue #4 — Replace hardcoded job matching with real scoring

**Type:** bug / feature · **Priority:** high

**Description:**
`repositories/jobs.go` `MatchJobs()` returns every job with a hardcoded `0.85` score —
it ignores `SemanticMatchFilter.UserSkills` and `MinScore`. `services/matching.go` is an
empty file. Recommendations are therefore meaningless, and the UI shows a made-up
"85% match".

**Files**
- `backend/services/matching.go` (empty — implement it)
- `backend/repositories/jobs.go`
- `backend/services/recommendation.go`
- `backend/handlers/jobs.go` (`Match` handler is also currently unregistered — see #10)

**Acceptance criteria**
- [ ] A matching service computes a real similarity score: at minimum a skills/keyword
      overlap against job title + description; optionally embedding-based similarity
      via `repositories.EmbeddingRepository.FindSimilarJobs` (pgvector, see `003_memory`).
- [ ] `MinScore` filters results; results are sorted best-first.
- [ ] `MatchJobs` no longer fabricates scores; empty skills → defined behavior (no
      matches, not all jobs at 0.85).
- [ ] Recommendations endpoint returns meaningful scores/labels the frontend can render.

---

### Issue #5 — Wire real job sources and make ingestion robust

**Type:** feature · **Priority:** medium

**Description:**
`main.go` only uses `SeedJobSource` (3 hardcoded jobs). The real fetchers
(`services/external_job_source.go` for Arbeitnow + Remotive + RemoteOK, and
`services/arbeitnow.go`) are never instantiated anywhere. Job ingestion runs at
startup with no retry/backoff and no way to re-run on demand.

**Files**
- `backend/services/external_job_source.go`
- `backend/services/arbeitnow.go`
- `backend/services/jobs.go`
- `backend/cmd/api/main.go`

**Acceptance criteria**
- [ ] `JobService` is constructed with a configurable source chain (external sources
      with a seed/fallback), driven by env config.
- [ ] Ingestion failures are retried (or at least logged and re-runnable) instead of
      only logging a warning at boot.
- [ ] Dedup by `external_id` keeps working across sources (`arbeitnow:...`,
      `remotive:...`, `remoteok:...` prefixes).
- [ ] No unused/duplicate fetchers: either `arbeitnow.go` is used or removed.

---

## Backend — @evans

### Issue #6 — Fix chat route: auth middleware missing

**Type:** bug · **Priority:** high

**Description:**
`RegisterChat` (`routes/routes.go`) registers `POST /api/chat` WITHOUT the Auth
middleware, but `ChatHandler.Chat` requires `middleware.GetUserID(r)` from context.
Since the auth middleware is the only thing that injects the user ID, every chat
request returns `401 authentication required` — even with a valid bearer token.
Chat is currently unusable.

**Files**
- `backend/routes/routes.go`
- `backend/handlers/chat.go`
- `backend/cmd/api/main.go`

**Acceptance criteria**
- [ ] `RegisterChat` wraps the handler with `middleware.Auth(jwt)` like the other
      protected routes.
- [ ] With a valid token, `POST /api/chat` reaches the service; without one it returns
      `401`.

---

### Issue #7 — Fix chat enablement: env var mismatch

**Type:** bug · **Priority:** high

**Description:**
`main.go` enables chat only when `cfg.BedrockModelID` (env `BEDROCK_MODEL_ID`) is set,
but `backend/.env.example` only documents `BEDROCK_CHAT_MODEL_ID`, and `config.go`
already loads `BedrockChatModelID` (env `BEDROCK_CHAT_MODEL_ID`) which is never used.
Following the docs never enables chat.

**Files**
- `backend/cmd/api/main.go`
- `backend/config/config.go`
- `backend/config/env.go`
- `backend/.env.example`

**Acceptance criteria**
- [ ] Chat gate uses the documented env var (`BEDROCK_CHAT_MODEL_ID` / a model the user
      actually sets), or the env example is corrected to match `main.go`.
- [ ] `BedrockChatModelID` / `BedrockEmbedModelID` are either actually wired in or
      removed (see #12).

---

### Issue #8 — Register the CV Tailor endpoint

**Type:** bug · **Priority:** high

**Description:**
`handlers/tailor.go` defines `TailorHandler.Generate` (`POST /api/tailor`) but it is
never registered in `routes/routes.go` or constructed in `main.go`. The frontend
`tailoringService.generate()` posts to `/api/tailor` and gets a 404 — the whole
"Tailor my CV" feature is dead on arrival.

**Files**
- `backend/handlers/tailor.go`
- `backend/routes/routes.go`
- `backend/cmd/api/main.go`

**Acceptance criteria**
- [ ] `POST /api/tailor` is registered, auth-protected, and constructs `TailorHandler`
      from the same `AIService` used by chat.
- [ ] The handler validates `resume_id` + `job_title` and returns a 502-style friendly
      error when the AI is unavailable (already partially handled).

---

### Issue #9 — Fix saved-jobs routing bugs

**Type:** bug · **Priority:** high

**Description:**
Two routing problems make save/delete of saved jobs non-functional:
1. `routes.go` maps `/api/saved-jobs` to `HandleSavedJobs`, which unconditionally calls
   `List`. A `POST` (save) therefore returns the list and saves nothing — the real
   `Save` handler is never invoked.
2. `HandleDeleteSavedJob` → `Remove` reads `r.PathValue("job_id")`, but the route
   pattern is `/api/saved-jobs/` with no `{job_id}` segment, so the path value is always
   empty and every delete returns 400.

**Files**
- `backend/routes/routes.go`
- `backend/handlers/saved_jobs.go`

**Acceptance criteria**
- [ ] `POST /api/saved-jobs` → `Save` (201 on success, 409 duplicate, 404 unknown job).
- [ ] `DELETE /api/saved-jobs/{job_id}` → `Remove` using the path segment (204 on
      success, 404 when not saved).
- [ ] `GET /api/saved-jobs` → `List` unchanged.

---

### Issue #10 — Clean up migrations (empty file + non-idempotent destructive step)

**Type:** bug · **Priority:** medium

**Description:**
- `005_indexes.up.sql` is an empty file. The runner skips empty migrations but never
  records them, so it re-processes the filename every boot (noisy, and it will never
  be marked applied).
- `006_fix_embedding_dimension.up.sql` drops and re-adds the `vector` column. Migrations
  are NOT wrapped in transactions (`runner.go` `applyOne`), so a crash after the `DROP
  COLUMN` but before the version is recorded leaves the schema broken and 006
  unrecorded — the next boot fails on `DROP COLUMN` (not idempotent).

**Files**
- `backend/migrations/005_indexes.up.sql`
- `backend/migrations/006_fix_embedding_dimension.up.sql`
- `backend/migrations/runner.go`

**Acceptance criteria**
- [ ] Empty `005` is removed or given real content (e.g. the intended index additions).
- [ ] `006` is made idempotent (e.g. `ALTER TABLE ... DROP COLUMN IF EXISTS vector`,
      and only re-add when missing) or wrapped in an advisory-lock/transaction pattern
      the runner supports, so a re-run after partial application succeeds.

---

### Issue #11 — Remove or wire dead config/clients (AgentRouter, MCP, embeddings)

**Type:** bug · **Priority:** low

**Description:**
Large amounts of config and client code are never used in production:
- `clients/agentrouter.go` (`AgentRouterClient`) — never instantiated; only tests.
- `clients/mcp.go` (`MCPClient`) — never instantiated; only tests.
- Config fields `AgentRouterAPIKey/BaseURL/Model`, `MCPEndpoint/MCPAPIKey/MCPClusterID`,
  `BedrockEmbedModelID` — loaded but unused.
- `clients/bedrock.go` `GenerateEmbedding` imports `repositories` for `EmbeddingDim` —
  a layering inversion (client should not depend on the repository layer).

**Files**
- `backend/clients/agentrouter.go`
- `backend/clients/mcp.go`
- `backend/config/config.go`, `backend/config/env.go`
- `backend/clients/bedrock.go`

**Acceptance criteria**
- [ ] Decide per item: wire it into the service layer with tests, or delete it.
- [ ] If `AgentRouter` becomes the default chat provider, `AIService` should depend on
      the `BedrockGenerator` interface and get a real default from config.
- [ ] `EmbeddingDim` moves to a shared location (e.g. `models`) so `clients` no longer
      imports `repositories`.

---

### Issue #12 — Resume processing pipeline (status never leaves "uploaded")

**Type:** feature · **Priority:** high

**Description:**
There is no resume parser. After upload the status stays `uploaded` forever, `parsed_text`
is never populated, and the frontend renders every resume as "processing". The whole
AI/memory feature depends on `parsed_text` (see `services/ai.go`), and embeddings for
resumes were designed in `003_memory`.

**Files**
- `backend/services/resume.go`
- `backend/services/ai.go` (consumes `ParsedText`)
- `backend/models/resume.go`
- New: `backend/services/resume_parser.go`
- `backend/repositories/embedding.go` (resume embedding via `Upsert`)

**Acceptance criteria**
- [ ] A parse step extracts plain text from PDF/DOC/DOCX/TXT after upload (respect the
      existing magic-byte validation) and transitions status
      `uploaded → parsing → parsed | failed` (see `models.ResumeStatus`).
- [ ] `failure_reason` is recorded on failure and surfaced to the user (not a raw error).
- [ ] On `parsed`, upsert a resume embedding (source_type `resume`) so resume-based
      similarity matching can work.

---

### Issue #13 — Consolidate the duplicate resume validators

**Type:** bug · **Priority:** low

**Description:**
There are two parallel validators with different rules: `utils/validator.go`
(`ValidateResume` — PDF/DOCX only, used only by tests) and `utils/file.go`
(`ValidateResumeFile` — PDF/DOC/DOCX/TXT with magic bytes, used by the handler and
service). `README.md` says `.pdf/.doc/.docx/.txt` are accepted, so the code and docs
must agree.

**Files**
- `backend/utils/validator.go`
- `backend/utils/file.go`
- `backend/services/resume.go`
- `backend/handlers/resume.go`
- `backend/utils/validator_test.go`

**Acceptance criteria**
- [ ] One validator remains; the other is deleted and its tests merged.
- [ ] Allowed extensions match `README.md` and `docs/API.md`.

---

## Frontend — @emma

### Issue #14 — Build the Profile page against a real API

**Type:** feature · **Priority:** high

**Description:**
`frontend/src/pages/Profile.tsx` is 100% hardcoded mock data ("Jane Doe", static work
history, static skills). The backend already has a `user_profiles` table
(`011_user_profiles.up.sql`) and `repositories/profile.go`, but there is no profile
handler or route, and the frontend doesn't call anything. The recommendations engine
uses `profile.Skills` + `profile.Experience`, so the profile must be editable.

> Backend dependency: this needs a `GET/PUT /api/profile` endpoint. Coordinate with
> @sos/@ashley to add the route (auth-protected) or this issue should block until one
> of the backend issues lands. Frontend work is what's scoped here.

**Files**
- `frontend/src/pages/Profile.tsx`
- `frontend/src/services/profile.ts` (new)
- `frontend/src/hooks/useProfile.ts` (new, optional)

**Acceptance criteria**
- [ ] Profile loads the authenticated user's data (bio, skills, experience, resume
      link) from the backend.
- [ ] Skills and experience are editable and persisted.
- [ ] Removing/adding skills updates what recommendations will use.
- [ ] Handles the "no profile yet" state (empty state + prompt to create).

---

### Issue #15 — Finish the Tailor (CV tailoring) workflow

**Type:** feature · **Priority:** medium

**Description:**
The Tailor flow (`pages/Tailor.tsx`, `services/tailoring.ts`) depends on `POST /api/tailor`
which 404s until backend Issue #8 lands. Beyond that, the page should be polished:
job picker fallback when no `jobId` is in the URL, resume-required state, regeneration,
and a sane experience when the AI is down. Also the "Submit application" button assumes
applications work (backend Issue #1).

**Files**
- `frontend/src/pages/Tailor.tsx`
- `frontend/src/services/tailoring.ts`
- `frontend/src/services/application.ts` (as needed)

**Acceptance criteria**
- [ ] Works end-to-end once backend #8 lands (pick job → generate → edit → submit).
- [ ] Clear inline states: no resume uploaded, AI unavailable, generation in progress,
      submission success.
- [ ] Result is not lost if the user navigates away and back within the same session
      (or a documented limitation).

---

### Issue #16 — Applications page: complete status management and empty states

**Type:** feature · **Priority:** medium

**Description:**
`pages/Applications.tsx` renders from `applicationService.list()`, which today returns
nothing because the backend is a stub (Issue #1). Once the backend is real:
- Verify status updates propagate (PUT `/api/applications/{id}`).
- Ensure the filter tabs match the canonical status set (`applied/screening/interview/
  offer/rejected/withdrawn`).
- Add a link/CTA to submit an application from the page (today you can only apply from
  the Tailor flow).

**Files**
- `frontend/src/pages/Applications.tsx`
- `frontend/src/components/ApplicationCard.tsx`
- `frontend/src/services/application.ts`

**Acceptance criteria**
- [ ] Status change persists and the card reflects the new state immediately.
- [ ] Filters/search match the backend's status vocabulary.
- [ ] Remove the unused `statusStyle` map in `ApplicationCard.tsx` (dead code).

---

## Frontend — @ian

### Issue #17 — Store chat history in the backend, not just localStorage

**Type:** bug / feature · **Priority:** high

**Description:**
`services/chat.ts` persists conversations only in localStorage
(`skillmatch-conversations`) and sends a `history` array to `POST /api/chat` that the
backend ignores (its `chatRequest` only accepts `message` + `resumeId`). The backend
already persists conversation turns per user (`003_memory`, `repositories/conversation.go`)
and builds history server-side. Today history is device-local, silently dropped on the
server, and mixed across users on a shared machine.

**Files**
- `frontend/src/services/chat.ts`
- `frontend/src/hooks/useChat.ts`
- `frontend/src/pages/Chat.tsx`
- `frontend/src/components/ChatBox.tsx`

**Acceptance criteria**
- [ ] Conversation history is loaded from the backend per authenticated user (new
      `GET` endpoint as needed — coordinate with @evans).
- [ ] New turns are persisted server-side; localStorage becomes a cache/offline layer,
      not the source of truth.
- [ ] The client stops sending the redundant `history` array once the server is the
      source of truth.

---

### Issue #18 — Fix the job-search params the backend can't understand

**Type:** bug · **Priority:** medium

**Description:**
`useJobs`/`jobsService.search` send `seniority` and `work_type`, which the backend
doesn't model (see backend Issue #3). Filters therefore appear to work in the UI but
do nothing. Also `formatJobDescription` and `normalizeJob` make many assumptions about
field shapes that should be locked down once the API shape is stable.

**Files**
- `frontend/src/services/jobs.ts`
- `frontend/src/hooks/useJobs.ts`
- `frontend/src/pages/Jobs.tsx`

**Acceptance criteria**
- [ ] The UI either uses the filters the backend supports or clearly marks the
      unsupported ones once backend #3 lands.
- [ ] Search results, pagination total, and job-detail loading are consistent with the
      API response shape.

---

### Issue #19 — Resume lifecycle states in the UI (no more "processing" forever)

**Type:** bug · **Priority:** medium

**Description:**
`services/resume.ts` `mapStatus` maps only `parsed → ready` and `failed → failed`;
everything else (including the only status the backend ever sets today, `uploaded`)
shows as "processing" indefinitely. This depends on backend Issue #12 (parsing
pipeline) landing, but the frontend should render the real status vocabulary and handle
the in-between `parsing` state with a live indicator.

**Files**
- `frontend/src/services/resume.ts`
- `frontend/src/components/ResumeList.tsx`
- `frontend/src/pages/Resume.tsx`

**Acceptance criteria**
- [ ] Statuses map cleanly to `uploaded | parsing | parsed | failed` and render
      distinctly (e.g. parsing shows a spinner, failed shows the failure reason).
- [ ] Resume list reflects parsing completion without a manual refresh (poll once the
      backend supports it).

---

### Issue #20 — Frontend cleanup & consistency pass

**Type:** bug · **Priority:** low

**Description:**
Small fixes to tighten the frontend before it's considered done:
- `services/api.ts`, `pages/Home.tsx` are empty files — remove or implement.
- `MatchRing` breaks for values outside 0–100 (CSS `--match` conic gradient) — clamp
  the value.
- The Google "Continue with Google" buttons on Login/Register are non-functional
  placeholders — wire to real OAuth, or disable them so it's not a broken affordance.
- Demo-auth (`VITE_DEMO_AUTH_*`) should be dev-only and clearly flagged (it already is
  gated on `import.meta.env.DEV`, verify that's respected everywhere).
- Error-message parsing (`errorMessage`/`requestError`) is duplicated across services —
  consolidate into a single helper.

**Files**
- `frontend/src/services/api.ts`
- `frontend/src/pages/Home.tsx`
- `frontend/src/components/MatchRing.tsx`
- `frontend/src/pages/Login.tsx`, `frontend/src/pages/Register.tsx`
- `frontend/src/services/*.ts`

**Acceptance criteria**
- [ ] No empty source files remain in `src/`.
- [ ] `MatchRing` clamps its value.
- [ ] One shared error-message helper is used by all services.
- [ ] Google auth either works or is disabled with a clear reason.

---

## Cross-cutting

### Issue #21 — Update documentation to match reality (README, API.md, openapi.yml, DATABASE.md)

**Type:** bug · **Priority:** low · **Assignee:** @evans (with contributions from whoever
owns the related feature)

**Description:**
Docs are stale and misleading:
- `README.md` and `docs/API.md` claim chat/jobs/recommendations/saved-jobs/applications
  are "not implemented" — they ARE routed (some broken, some stubbed). Reconcile once the
  issues above land.
- `docs/DATABASE.md` documents a `jobs` schema (work_type, seniority, posted_at, source,
  is_active) and an `applications` schema that don't match the actual migrations.
- `docs/openapi.yml` only covers resumes and documents a `PUT /api/resumes/{id}` that
  doesn't exist (replacement is done via `POST /api/resumes` with `replaceId`).
- Delete committed `coverage` artifacts (repo root `coverage`, `backend/coverage`).

**Files**
- `README.md`
- `docs/API.md`
- `docs/DATABASE.md`
- `docs/openapi.yml`
- `coverage`, `backend/coverage`

**Acceptance criteria**
- [ ] README reflects what's actually implemented after the fixes.
- [ ] API docs (API.md/openapi.yml) match real routes, methods, and response shapes.
- [ ] DATABASE.md matches the applied migrations.
- [ ] `coverage` files removed from the repo.
