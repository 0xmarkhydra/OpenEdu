package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/0xmarkhydra/FreeEducation/services/api/internal/auth"
	"github.com/0xmarkhydra/FreeEducation/services/api/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Run intentionally with operator-provided credentials, never on API startup.
// It is single-use when no active platform admin exists.
func main() {
	email := strings.ToLower(strings.TrimSpace(os.Getenv("BOOTSTRAP_ADMIN_EMAIL")))
	password := os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	dsn := os.Getenv("DATABASE_URL")
	if !strings.Contains(email, "@") || len(email) > 254 || dsn == "" {
		log.Fatal("configure BOOTSTRAP_ADMIN_EMAIL and DATABASE_URL")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal("BOOTSTRAP_ADMIN_PASSWORD must be 12-72 bytes")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := migrations.Apply(ctx, db); err != nil {
		log.Fatal("database migration: ", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(462022)"); err != nil {
		log.Fatal(err)
	}
	var exists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM platform_roles WHERE role='platform_admin')").Scan(&exists)
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		log.Fatal("platform administrator already provisioned; bootstrap is disabled")
	}
	var userID string
	err = tx.QueryRowContext(ctx, `INSERT INTO users(email,display_name) VALUES($1,'OpenEdu administrator') RETURNING id::text`, email).Scan(&userID)
	if err != nil {
		log.Fatal("admin creation failed: ", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_credentials(user_id,password_hash) VALUES($1::uuid,$2)`, userID, hash); err != nil {
		log.Fatal(err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO platform_roles(user_id,role) VALUES($1::uuid,'platform_admin')`, userID); err != nil {
		log.Fatal(err)
	}
	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Initial platform admin provisioned successfully; remove bootstrap environment variables.")
}
