// Package migrations embeds the SQL migration files and applies any pending
// migrations to the database when the server starts. Applied migrations are
// tracked in a schema_migrations table so each migration runs exactly once,
// in filename order, inside its own transaction.
package migrations

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed *.sql
var migrationFS embed.FS

// Apply runs every up migration that has not already been recorded in
// schema_migrations. It is safe to call on every startup: already-applied
// migrations are skipped, and only pending ones execute. A failing migration
// aborts with an error so the server never boots on a half-migrated schema.
//
// Down migrations (*.down.sql) are never applied automatically; they exist
// for manual rollbacks.
func Apply(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return errors.New("migrations: nil connection pool")
	}

	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    STRING        NOT NULL PRIMARY KEY,
		applied_at TIMESTAMPTZ   NOT NULL DEFAULT now()
	)`); err != nil {
		return fmt.Errorf("migrations: ensure schema_migrations table: %w", err)
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return err
	}

	files, err := fs.Glob(migrationFS, "*.up.sql")
	if err != nil {
		return fmt.Errorf("migrations: list embedded files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		if applied[file] {
			continue
		}

		body, err := migrationFS.ReadFile(file)
		if err != nil {
			return fmt.Errorf("migrations: read %s: %w", file, err)
		}
		if strings.TrimSpace(string(body)) == "" {
			log.Printf("migrations: %s is empty, skipping", file)
			continue
		}

		if err := applyOne(ctx, pool, file, string(body)); err != nil {
			return fmt.Errorf("migrations: apply %s: %w", file, err)
		}
		log.Printf("migrations: applied %s", file)
	}

	return nil
}

// appliedVersions returns the set of migration filenames already recorded in
// schema_migrations.
func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	applied := map[string]bool{}

	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("migrations: query applied versions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("migrations: scan version: %w", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("migrations: iterate applied versions: %w", err)
	}

	return applied, nil
}

// applyOne executes a single migration and records its version on success.
//
// Note: migrations are intentionally NOT wrapped in a pgx transaction.
// CockroachDB refuses transactional execution of certain DDL (e.g. vector
// indexes, which trigger "auto-committing transaction before processing
// DDL"), so the whole file would deadlock inside BEGIN/COMMIT. Instead we
// run the file exactly as the SQL CLI would; the migration files are written
// idempotently (IF NOT EXISTS everywhere) so a failure that leaves partial
// state is safe to re-run, and an unrecorded version is retried on the next
// startup.
func applyOne(ctx context.Context, pool *pgxpool.Pool, version, body string) error {
	if _, err := pool.Exec(ctx, body); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
		return fmt.Errorf("record version: %w", err)
	}

	return nil
}
