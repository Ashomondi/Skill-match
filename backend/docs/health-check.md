# Production Health Checks

The API exposes `GET /health` as a dependency-aware readiness check.

## Responses

- `200 OK` with `status: "healthy"` means the database and object storage responded to their checks.
- `503 Service Unavailable` with `status: "degraded"` means a dependency is unavailable or not configured.

The response includes a `dependencies` object with `ok`, `down`, or `not configured` statuses. Failure messages are intentionally generic and never include database, AWS, credentials, or connection details.

## Monitoring

Configure the load balancer, container orchestrator, or uptime monitor to poll `/health` every 10-30 seconds and alert on consecutive non-`200` responses. Use the `X-Request-ID` response header to correlate checks with structured request logs. The endpoint has a three-second timeout for dependency checks so an unavailable service cannot hold probes indefinitely.

This endpoint is intended for readiness and dependency monitoring. Keep authentication off the route so infrastructure can probe it directly, and do not use it as a substitute for application metrics or business-level alerts.
