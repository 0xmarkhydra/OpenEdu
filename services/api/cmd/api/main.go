package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/0xmarkhydra/FreeEducation/services/api/internal/auth"
	"github.com/0xmarkhydra/FreeEducation/services/api/internal/catalog"
	"github.com/0xmarkhydra/FreeEducation/services/api/internal/httpapi"
	"github.com/0xmarkhydra/FreeEducation/services/api/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func env(key, fallback string) string {
	if s := os.Getenv(key); s != "" {
		return s
	}
	return fallback
}
func main() {
	v, err := auth.NewVerifier(os.Getenv("JWT_SECRET"), env("JWT_ISSUER", "freeeducation-identity"), env("JWT_AUDIENCE", "freeeducation-api"))
	if err != nil {
		log.Fatal(err)
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL must be configured")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := migrations.Apply(context.Background(), db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)
	port := env("PORT", "8080")
	server := &http.Server{Addr: ":" + port, Handler: httpapi.New(catalog.NewPostgresStore(db), v, db), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("FreeEducation API listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}
