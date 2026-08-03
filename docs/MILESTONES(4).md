# MILESTONES.md

# Sprint 1 — Foundation & Authentication

## Issue 1: Initialize Backend API
**Assignee:** Evans

**Description**

Initialize the backend application by configuring the Gin HTTP server, registering global middleware, defining the base routing structure, exposing a health check endpoint, and establishing the application's startup sequence so all future modules plug into a consistent foundation.

**Files**

```text
backend/
├── cmd/api/main.go (Evans)
├── routes/routes.go (Evans)
├── middleware/auth.go (Evans)
├── middleware/cors.go (Evans)
├── middleware/logging.go (Evans)
└── middleware/recovery.go (Evans)
```

## Issue 2: Database & Configuration
**Assignee:** Sospeter

**Description**

Design and implement the persistence layer by creating the initial CockroachDB schema, user repository, and database migration required to store and retrieve user data. Environment and application configuration are handled separately.

**Files**

```text
backend/
├── config/config.go (Evans)
├── config/env.go (Evans)
├── repositories/user.go (Sospeter)
└── migrations/001_initial_schema.sql (Sospeter)
```

## Issue 3: Authentication
**Assignee:** Ashley

**Description**

Develop the complete authentication workflow, including user registration, login, password hashing with bcrypt, JWT generation and validation, and the business logic required to securely authenticate users.

**Files**

```text
backend/
├── handlers/auth.go (Ashley)
├── services/auth.go (Ashley)
├── models/user.go (Ashley)
├── utils/jwt.go (Ashley)
└── utils/password.go (Ashley)
```

## Issue 4: Authentication UI
**Assignee:** Ian

**Description**

Create the authentication user interface, including responsive login and registration pages, client-side validation, navigation, and routing to protected areas of the application.

**Files**

```text
frontend/src/
├── pages/Login.tsx (Ian)
├── pages/Register.tsx (Ian)
└── routes.tsx (Ian)
```

## Issue 5: Auth Integration
**Assignee:** Emma

**Description**

Connect the frontend authentication workflow to the backend REST API, manage authentication state, persist user sessions, and provide navigation updates based on login status.

**Files**

```text
frontend/src/
├── services/auth.ts (Emma)
├── hooks/useAuth.ts (Emma)
├── components/Navbar.tsx (Emma)
└── pages/Dashboard.tsx (Emma)
```

---

# Sprint 2 — Resume Management

## Issue 6: Resume API
**Assignee:** Ashley

**Description**

Create the backend endpoints and business logic required to upload resumes, validate supported file formats, process uploaded files, and prepare them for storage and later AI analysis.

**Files**

```text
backend/
├── handlers/resume.go (Ashley)
├── services/resume.go (Ashley)
└── models/resume.go (Ashley)
```

## Issue 7: Resume Repository
**Assignee:** Sospeter

**Description**

Design the database schema for resumes, implement the repository layer for storing and retrieving resume metadata, and create the required database migrations.

**Files**

```text
backend/
├── repositories/resume.go (Sospeter)
└── migrations/002_resume.sql (Sospeter)
```

## Issue 8: Amazon S3 Integration
**Assignee:** Evans

**Description**

Integrate Amazon S3 as the document storage service, implement secure upload functionality, and provide reusable utilities for storing and retrieving uploaded resumes.

**Files**

```text
backend/
├── clients/s3.go (Evans)
└── utils/file.go (Evans)
```

## Issue 9: Resume Upload UI
**Assignee:** Ian

**Description**

Develop an intuitive interface that allows users to upload, replace, view, and manage their resumes while displaying upload progress and validation feedback.

**Files**

```text
frontend/src/
├── pages/Resume.tsx (Ian)
└── components/ResumeUploader.tsx (Ian)
```

## Issue 10: Resume Integration
**Assignee:** Emma

**Description**

Connect the frontend resume management interface with the backend APIs, handle upload requests, display progress, and manage success and error states.

**Files**

```text
frontend/src/
├── services/resume.ts (Emma)
└── hooks/useUpload.ts (Emma)
```

---

# Sprint 3 — AI Memory & Chat

## Issue 11: Chat Backend
**Assignee:** Ashley

**Description**

Build the conversational backend by implementing chat endpoints, coordinating requests to Amazon Bedrock, and preparing prompts using the user's stored context and memory.

**Files**

```text
backend/
├── handlers/chat.go (Ashley)
├── services/chat.go (Ashley)
└── services/ai.go (Ashley)
```

## Issue 12: Memory Layer
**Assignee:** Sospeter

**Description**

Design the database schema and repository layer for conversation history and embeddings, ensuring efficient persistence and retrieval of long-term AI memory.

**Files**

```text
backend/
├── repositories/conversation.go (Sospeter)
├── repositories/embedding.go (Sospeter)
└── migrations/003_memory.sql (Sospeter)
```

## Issue 13: Bedrock Integration
**Assignee:** Evans

**Description**

Develop the Amazon Bedrock client, integrate the CockroachDB MCP client, and implement the memory service that combines historical context with current user prompts.

**Files**

```text
backend/
├── clients/bedrock.go (Evans)
├── clients/mcp.go (Evans)
└── services/memory.go (Evans)
```

## Issue 14: Chat Interface
**Assignee:** Ian

**Description**

Create a responsive chat experience that supports real-time conversations, conversation history, and AI-generated responses in an intuitive interface.

**Files**

```text
frontend/src/
├── pages/Chat.tsx (Ian)
└── components/ChatBox.tsx (Ian)
```

## Issue 15: Chat Integration
**Assignee:** Emma

**Description**

Integrate the chat interface with backend APIs, manage message state, send user prompts, receive AI responses, and maintain conversation history.

**Files**

```text
frontend/src/
├── services/chat.ts (Emma)
└── hooks/useChat.ts (Emma)
```

---

# Sprint 4 — Job Search & Matching

## Issue 16: Job Search API
**Assignee:** Ashley

**Description**

Develop APIs that search available jobs, process user preferences, and expose recommendation endpoints that deliver personalized results.

**Files**

```text
backend/
├── handlers/jobs.go (Ashley)
├── services/jobs.go (Ashley)
└── services/recommendation.go (Ashley)
```

## Issue 17: Job Repository
**Assignee:** Sospeter

**Description**

Implement database support for jobs and saved jobs, including repositories, migrations, and efficient data access patterns.

**Files**

```text
backend/
├── repositories/job.go (Sospeter)
├── repositories/saved_job.go (Sospeter)
└── migrations/004_jobs.sql (Sospeter)
```

## Issue 18: Matching Engine
**Assignee:** Evans

**Description**

Develop the semantic matching engine using CockroachDB Distributed Vector Indexing to compare resumes, user preferences, and job descriptions for intelligent recommendations.

**Files**

```text
backend/
├── services/matching.go (Evans)
└── clients/cockroach.go (Evans)
```

## Issue 19: Jobs UI
**Assignee:** Ian

**Description**

Design and implement the job discovery interface, including searchable job listings, recommendation cards, and filtering capabilities.

**Files**

```text
frontend/src/
├── pages/Jobs.tsx (Ian)
└── components/JobCard.tsx (Ian)
```

## Issue 20: Jobs Integration
**Assignee:** Emma

**Description**

Connect the frontend job pages to backend APIs, manage search queries, filters, recommendations, and loading states.

**Files**

```text
frontend/src/
├── services/jobs.ts (Emma)
└── hooks/useJobs.ts (Emma)
```

---

# Sprint 5 — Saved Jobs & Applications

## Issue 21: Applications API
**Assignee:** Ashley

**Description**

Develop backend APIs that allow users to save jobs, submit applications, update application status, and retrieve application history.

**Files**

```text
backend/
├── handlers/applications.go (Ashley)
└── services/applications.go (Ashley)
```

## Issue 22: Application Repository
**Assignee:** Sospeter

**Description**

Implement the repository layer responsible for storing applications, tracking status changes, and retrieving application records.

**Files**

```text
backend/
└── repositories/application.go (Sospeter)
```

## Issue 23: Protected Routes
**Assignee:** Evans

**Description**

Complete route protection and authorization middleware to ensure only authenticated users can access protected resources.

**Files**

```text
backend/
├── routes/routes.go (Evans)
└── middleware/auth.go (Evans)
```

## Issue 24: Applications UI
**Assignee:** Ian

**Description**

Create pages that allow users to monitor saved jobs, application progress, and status updates from a centralized dashboard.

**Files**

```text
frontend/src/
├── pages/Applications.tsx (Ian)
└── components/ApplicationCard.tsx (Ian)
```

## Issue 25: Saved Jobs Integration
**Assignee:** Emma

**Description**

Integrate the frontend application tracker with backend services for saved jobs, application updates, and synchronized user data.

**Files**

```text
frontend/src/
├── services/application.ts (Emma)
└── components/SavedJobs.tsx (Emma)
```

---

# Sprint 6 — Polish & Deployment

## Issue 26: Backend Polish
**Assignee:** Ashley

**Description**

Perform final backend stabilization by resolving defects, refining API behavior, improving code quality, and completing technical documentation.

**Files**

```text
backend/
├── handlers/ (Ashley)
└── services/ (Ashley)
```

## Issue 27: Database Optimization
**Assignee:** Sospeter

**Description**

Improve database performance by optimizing queries, creating indexes, and validating migration efficiency.

**Files**

```text
backend/
├── migrations/005_indexes.sql (Sospeter)
└── repositories/ (Sospeter)
```

## Issue 28: Deployment
**Assignee:** Evans

**Description**

Containerize the backend, configure automated CI/CD workflows, and prepare the deployment pipeline for a production-ready release.

**Files**

```text
backend/
├── Dockerfile (Evans)
└── .github/workflows/ci.yml (Evans)
```

## Issue 29: UI Polish
**Assignee:** Ian

**Description**

Polish the user interface by improving responsiveness, consistency, animations, and overall user experience across supported devices.

**Files**

```text
frontend/src/
├── pages/ (Ian)
└── components/ (Ian)
```

## Issue 30: Frontend QA
**Assignee:** Emma

**Description**

Enhance accessibility, refine loading and error handling, and perform final frontend quality assurance before release.

**Files**

```text
frontend/src/
├── services/ (Emma)
└── hooks/ (Emma)
```

---

