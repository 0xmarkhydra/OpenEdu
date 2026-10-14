package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"
)

//go:embed 0001_initial.sql
var scripts embed.FS

// Apply serializes bootstrap across horizontally scaled API replicas.
// Migrations are executed transactionally and recorded exactly once.
func Apply(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(462021)"); err != nil {
		return fmt.Errorf("migration lock: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		return fmt.Errorf("migration ledger: %w", err)
	}
	var done bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)", "0001_initial").Scan(&done); err != nil {
		return err
	}
	if !done {
		sqlText, err := scripts.ReadFile("0001_initial.sql")
		if err != nil {
			return fmt.Errorf("read initial migration: %w", err)
		}
		// First migration is plain DDL without stored procedures/quoted semicolons.
		for _, statement := range splitStatements(string(sqlText)) {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				return fmt.Errorf("migration 0001_initial failed: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES ($1)", "0001_initial"); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// splitStatements is intentionally scoped to simple DDL. Strip SQL line comments
// before splitting so semicolons in comments cannot be executed as SQL.
// Use a dedicated SQL migration parser if future migrations include procedural
// blocks, dollar-quoted strings or semicolons within string literals.
func splitStatements(script string) []string {
	var lines []string
	for _, line := range strings.Split(script, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		lines = append(lines, line)
	}
	var statements []string
	for _, raw := range strings.Split(strings.Join(lines, "\n"), ";") {
		if statement := strings.TrimSpace(raw); statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}
