// Package clients contains thin wrappers around external service SDKs
// (CockroachDB, S3, Bedrock, MCP). This file owns the CockroachDB
// connection pool that every repository in repositories/ is constructed
// with.
package clients
