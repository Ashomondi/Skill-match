# SkillMatch Architecture

## 1. Overview
SkillMatch is an AI-powered job search assistant that uses CockroachDB as a persistent memory layer to deliver personalized job recommendations. The system remembers user profiles, resumes, conversations, saved jobs, and application history.

## 2. Objectives
- Conversational AI job search.
- Persistent user memory.
- Resume upload and parsing.
- Semantic matching between resumes and jobs.
- Personalized recommendations.
- Production-ready, cloud-native architecture.

## 3. Technology Stack
| Layer | Technology |
|---|---|
| Frontend | React + Vite + TypeScript |
| Backend | Go (net/http) |
| AI | Amazon Bedrock |
| Database | CockroachDB |
| Storage | Amazon S3 |
| Auth | JWT |
| Vector Search | CockroachDB Distributed Vector Indexing |
| Agent Access | CockroachDB Managed MCP Server |

## 4. High-Level Flow
1. User signs up.
2. User uploads resume (stored in S3).
3. Backend extracts text and creates embeddings.
4. Embeddings + metadata stored in CockroachDB.
5. User chats with AI.
6. Agent retrieves long-term memory through MCP.
7. Agent performs semantic search using vector indexing.
8. Bedrock generates a personalized response.
9. User saves jobs and tracks applications.
10. Memory continuously improves recommendations.

## 5. Functional Modules
- Authentication
- Profile Management
- Resume Management
- AI Chat
- Job Search
- Recommendation Engine
- Saved Jobs
- Application Tracker
- Memory Service
- Admin/Monitoring

## 6. Backend Architecture
Presentation (net/http Handlers)
→ Services
→ Domain
→ Repository
→ CockroachDB / S3 / Bedrock

Request Flow:
HTTP Request
→ net/http Server
→ http.ServeMux
→ Middleware
→ Handler
→ Service
→ Repository
→ CockroachDB / Amazon S3 / Amazon Bedrock

Suggested layout:

Replace the suggested layout with:

```text
skill-match/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go
│   ├── config/
│   │   ├── config.go
│   │   └── env.go
│   ├── routes/
│   │   ├── routes.go
│   │   └── mux.go
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── users.go
│   │   ├── resume.go
│   │   ├── chat.go
│   │   ├── jobs.go
│   │   ├── applications.go
│   │   └── saved_jobs.go
│   ├── services/
│   │   ├── auth.go
│   │   ├── ai.go
│   │   ├── resume.go
│   │   ├── chat.go
│   │   ├── jobs.go
│   │   ├── recommendation.go
│   │   ├── matching.go
│   │   ├── memory.go
│   │   ├── application.go
│   │   └── saved_jobs.go
│   ├── repositories/
│   │   ├── user.go
│   │   ├── resume.go
│   │   ├── conversation.go
│   │   ├── embedding.go
│   │   ├── job.go
│   │   ├── saved_job.go
│   │   └── application.go
│   ├── models/
│   │   ├── user.go
│   │   ├── resume.go
│   │   ├── conversation.go
│   │   ├── embedding.go
│   │   ├── job.go
│   │   ├── saved_job.go
│   │   └── application.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   ├── logging.go
│   │   └── recovery.go
│   ├── clients/
│   │   ├── bedrock.go
│   │   ├── cockroach.go
│   │   ├── mcp.go
│   │   └── s3.go
│   ├── utils/
│   │   ├── errors.go
│   │   ├── jwt.go
│   │   ├── password.go
│   │   ├── response.go
│   │   └── validator.go
│   ├── migrations/
│   │   ├── 001_initial_schema.sql
│   │   ├── 002_resume.sql
│   │   ├── 003_memory.sql
│   │   ├── 004_jobs.sql
│   │   └── 005_indexes.sql
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
├── frontend/
│   ├── public/
│   │   ├── favicon.ico
│   │   └── robots.txt
│   ├── src/
│   │   ├── assets/
│   │   │   ├── logo.svg
│   │   │   └── hero.png
│   │   ├── components/
│   │   │   ├── Navbar.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   ├── ChatBox.tsx
│   │   │   ├── JobCard.tsx
│   │   │   ├── ResumeUploader.tsx
│   │   │   ├── ApplicationCard.tsx
│   │   │   └── SavedJobs.tsx
│   │   ├── hooks/
│   │   │   ├── useAuth.ts
│   │   │   ├── useChat.ts
│   │   │   ├── useJobs.ts
│   │   │   └── useUpload.ts
│   │   ├── pages/
│   │   │   ├── Home.tsx
│   │   │   ├── Login.tsx
│   │   │   ├── Register.tsx
│   │   │   ├── Dashboard.tsx
│   │   │   ├── Resume.tsx
│   │   │   ├── Chat.tsx
│   │   │   ├── Jobs.tsx
│   │   │   ├── Applications.tsx
│   │   │   └── Profile.tsx
│   │   ├── services/
│   │   │   ├── api.ts
│   │   │   ├── auth.ts
│   │   │   ├── resume.ts
│   │   │   ├── chat.ts
│   │   │   ├── jobs.ts
│   │   │   └── application.ts
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── routes.tsx
│   ├── package.json
│   └── vite.config.ts
├── docs/
│   ├── ARCHITECTURE.md
│   ├── CONTRIBUTORS.md
│   ├── MILESTONES.md
│   ├── API.md
│   └── DATABASE.md
├── README.md
├── LICENSE
└── .gitignore
```


## 7. Data Model
Core entities:
- User
- Resume
- Conversation
- Memory
- Job
- SavedJob
- Application
- Embedding

## 8. AI Memory
Persistent:
- Profile
- Skills
- Resume
- Search history
- Saved jobs
- Applications
- Conversations

Transient:
- Current prompt
- Session context

## 9. Resume Pipeline
Upload → S3 → Parse → Embed → CockroachDB → Match Jobs

## 10. Job Recommendation Pipeline
Query → Retrieve Memory → Semantic Search → Rank → Bedrock → Response

## 11. APIs
/auth
/users
/resumes
/jobs
/chat
/memory
/applications
/saved-jobs

## 12. Security
- JWT
- Password hashing
- Input validation
- Least privilege IAM
- Signed S3 URLs
- Audit logging

## 13. Observability
- Structured logs
- Metrics
- Health endpoint
- Request IDs

## 14. Backend Notes
- Uses Go's standard library `net/http`.
- Routing is handled by `http.ServeMux`.
- Middleware is implemented as `http.Handler` wrappers.
- No external web framework is required.

## 14. Deployment
React → CDN
Go API → Container
CockroachDB Cloud
Amazon Bedrock
Amazon S3

## 15. Future Enhancements
- Interview preparation
- Cover letter generation
- Notifications
- Recruiter portal
- Multi-agent workflows

## 16. End-to-End User Journey
Register → Complete profile → Upload resume → AI analyzes profile → Search jobs → Receive recommendations → Save/apply → AI remembers and improves future recommendations.
