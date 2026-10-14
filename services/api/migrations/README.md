# Database migrations

0001_initial.sql creates schools, users, tenant memberships, academic years, class rosters, guardian verification references, books, chapters, lessons and audit history.

On a fresh PostgreSQL 16 database, apply once:

    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f services/api/migrations/0001_initial.sql

For local Docker PostgreSQL when psql is unavailable on the host:

    docker compose exec -T postgres psql -U freeeducation -d freeeducation -v ON_ERROR_STOP=1 < services/api/migrations/0001_initial.sql

The API automatically runs the initial migration once on startup, guarded by an advisory transaction lock and the schema_migrations tracking table. Manual application is an emergency/local troubleshooting option, not a production deployment requirement. Do NOT rerun an already-applied migration. Production deployment needs a dedicated migration runner with a tracking table and advisory lock. Future schema changes are strictly append-only migration files, not edits to an already-installed baseline.

An active trusted identity service must provision matching users and school memberships before privileged book authoring is possible. No built-in admin accounts/passwords are included.

Do not seed identifiable student/guardian data into development or demos.
