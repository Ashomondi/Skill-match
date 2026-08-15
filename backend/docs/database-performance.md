# Database Performance Review

The repository read paths were reviewed before production deployment. Migration `010_query_performance` adds indexes for recent job listing and user-scoped application access, and removes a redundant job-interaction index whose prefix is already covered by the unique `(user_id, job_id, interaction_type)` index.

Optimizations applied:

- Job listing is covered by `(created_at DESC, id)`.
- Application history now authorizes and reads history in one joined query instead of issuing a separate application lookup.
- Saved-job listing is bounded to 100 rows and uses the existing `(user_id, saved_at DESC)` index.
- Conversation history and resume listing retain their existing user/time composite indexes and bounded limits.
- Vector search builds SQL only for filters that are present. This avoids `($param = '' OR ...)` predicates and keeps the nearest-neighbor `ORDER BY vector <=> query LIMIT k` shape required by the vector index.
- Embedding source lookups use the unique `(source_type, source_id)` index; job recommendation search uses the vector index.

Before production rollout, run `EXPLAIN ANALYZE` on representative job, application, saved-job, conversation, and vector-search requests against production-sized data. Confirm the expected indexes are selected, watch query latency and contention, and schedule index creation during an appropriate maintenance window if the cluster requires it.
