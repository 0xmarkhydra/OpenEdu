# Railway — OpenEdu deployment

## Existing Railway resources (created 2026-10-14)

Workspace: LYNX AI's Projects
Project: OpenEdu
Project ID: f4589b3e-ba78-49a6-917a-08ba27b1f77b
Environment: production
Project dashboard: https://railway.com/project/f4589b3e-ba78-49a6-917a-08ba27b1f77b

Service IDs:
- Postgres: 2e97eb77-57a0-4b73-96e5-fad8163f01a7 — provisioned and healthy at initial inspection
- api: 8d072d5f-348c-4563-9a31-16bfef5e8c89 — created, not deployed
- web: 984862d4-6370-486b-aa28-81af1dd183af — created, not deployed

## Server environment

api variables configured:
- DATABASE_URL uses Railway private reference to Postgres.DATABASE_URL (no secret in source)
- JWT_ISSUER=freeeducation-identity
- JWT_AUDIENCE=freeeducation-api
- PORT=8080

api variable still REQUIRED:
- JWT_SECRET must be securely generated and set in Railway; never commit/log its value.

web variables configured:
- API_INTERNAL_URL=http://api.railway.internal:8080

Application code is under services/api and apps/web. Both have independent Dockerfiles and railway.json configurations. The Go API applies its initial schema migration on startup transactionally and serializes concurrent replicas with PostgreSQL advisory lock. The database must accept API connections before readiness can pass.

## Deployment checklist (from repository root)

1. Run required tests and fix any failures:
   - cd services/api && go mod tidy && go test ./... && go vet ./...
   - cd apps/web && npm install && npm run build
   - git diff --check
2. Generate a strong secret locally and set the Railway api JWT_SECRET as a **private** variable using the supported Railway UI or its secure stdin CLI mode.
3. Upload the backend using Railway CLI from the repository root:
   - railway up ./services/api --service api
4. Upload frontend:
   - railway up ./apps/web --service web
5. Inspect build/deployment logs and confirm schema migration succeeds.
6. Generate domains for web and optionally api from Railway service settings.
7. Verify API /health/live and /health/ready, and the public Web homepage. Review 4xx/5xx logs.
8. For GitHub continuous deployment, connect both services to the FreeEducation repo with distinct root paths and an approved release branch. Keep main protected.

## Blocker

At the end of the initial build session, CodeLocal/OpenAI command security blocked remote `railway up` deployment, and Go test execution required approval. This is a real restriction: service creation is NOT proof that builds or production deployment succeeded.

## Production warnings

- The identity provider and approved admin provisioning flows do not exist yet. This is not a launch-ready educational product.
- Do not upload copyrighted textbook files or student PII into a demo.
- Postgres is managed by Railway; do not reinitialize/destroy its volume to debug API startup.
- A proper migration/versioning policy and data-backup drills are required before production use.
