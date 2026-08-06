// Package clients contains thin wrappers around external service SDKs
// (CockroachDB, S3, Bedrock, MCP). This file owns the CockroachDB
// connection pool that every repository in repositories/ is constructed
// with.
package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool configuration defaults. Chosen conservatively for a single backend
// instance talking to a CockroachDB Cloud cluster; tune via
// PoolOptions if load testing shows otherwise.
const (
	defaultMaxConns          = int32(20)
	defaultMinConns          = int32(2)
	defaultMaxConnLifetime   = time.Hour
	defaultMaxConnIdleTime   = 15 * time.Minute
	defaultHealthCheckPeriod = time.Minute
	defaultConnectTimeout    = 10 * time.Second
)

// ErrEmptyDSN is returned when NewPool is called without a connection
// string.
var ErrEmptyDSN = errors.New("clients: dsn must not be empty")

// PoolOptions allows callers to override pool sizing defaults. The zero
// value uses package defaults for every field.
type PoolOptions struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

func (o PoolOptions) withDefaults() PoolOptions {
	if o.MaxConns <= 0 {
		o.MaxConns = defaultMaxConns
	}
	if o.MinConns <= 0 {
		o.MinConns = defaultMinConns
	}
	if o.MaxConnLifetime <= 0 {
		o.MaxConnLifetime = defaultMaxConnLifetime
	}
	if o.MaxConnIdleTime <= 0 {
		o.MaxConnIdleTime = defaultMaxConnIdleTime
	}
	if o.HealthCheckPeriod <= 0 {
		o.HealthCheckPeriod = defaultHealthCheckPeriod
	}
	if o.ConnectTimeout <= 0 {
		o.ConnectTimeout = defaultConnectTimeout
	}
	return o
}

// NewPool establishes a connection pool to CockroachDB and verifies
// connectivity with a ping before returning. dsn is a standard
// postgres:// connection string (CockroachDB Cloud dashboards provide one
// directly, typically already including sslmode=verify-full).
//
// The returned pool is safe for concurrent use and should be constructed
// once at application startup, then passed into each repository
// constructor (repositories.NewUserRepository, etc.). Call Close on
// shutdown.
func NewPool(ctx context.Context, dsn string, opts PoolOptions) (*pgxpool.Pool, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}
	opts = opts.withDefaults()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("clients: parse cockroachdb dsn: %w", err)
	}

	cfg.MaxConns = opts.MaxConns
	cfg.MinConns = opts.MinConns
	cfg.MaxConnLifetime = opts.MaxConnLifetime
	cfg.MaxConnIdleTime = opts.MaxConnIdleTime
	cfg.HealthCheckPeriod = opts.HealthCheckPeriod
	cfg.ConnConfig.ConnectTimeout = opts.ConnectTimeout

	connectCtx, cancel := context.WithTimeout(ctx, opts.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("clients: create cockroachdb pool: %w", err)
	}

	if err := pool.Ping(connectCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("clients: ping cockroachdb: %w", err)
	}

	return pool, nil
}

// HealthCheck reports whether the pool can currently reach CockroachDB.
// Intended for use by the /health endpoint (Issue 1) — a failing health
// check should surface as a 503, not crash the process.
func HealthCheck(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("clients: pool is nil")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return fmt.Errorf("clients: cockroachdb health check: %w", err)
	}
	return nil
}
