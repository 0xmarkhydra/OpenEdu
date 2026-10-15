# Railway — OpenEdu deployment

## Existing Railway resources (created 2026-10-14)

Workspace: LYNX AI's Projects
Project: OpenEdu
Project ID: f4589b3e-ba78-49a6-917a-08ba27b1f77b
Environment: production
Project dashboard: https://railway.com/project/f4589b3e-ba78-49a6-917a-08ba27b1f77b

Service IDs:
- Postgres: 2e97eb77-57a0-4b73-96e5-fad8163f01a7 — provisioned and healthy at initial inspection
- api: 8d072d5f-348c-4563-9a31-16bfef5e8c89 — deployed successfully 2026-10-14
- web: 984862d4-6370-486b-aa28-81af1dd183af — deployed successfully 2026-10-14

## Server environment

api variables configured:
- DATABASE_URL uses Railway private reference to Postgres.DATABASE_URL (no secret in source)
- JWT_ISSUER=freeeducation-identity
- JWT_AUDIENCE=freeeducation-api
- PORT=8080

api variable configured:
- JWT_SECRET generated with OpenSSL and set through Railway stdin. Value is not included in the repository.

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
8. For GitHub continuous deployment, connect both services to the OpenEdu repo with distinct root paths and an approved release branch. Keep main protected.

## Verified deployment, 2026-10-14

- GitHub repository: https://github.com/0xmarkhydra/OpenEdu, branch `dev`, deployed commit `9e27650`.
- `api` and `web` both deploy automatically from GitHub, using Railway service root directories `/services/api` and `/apps/web`.
- API: Docker image build succeeded, PostgreSQL bootstrap migration applied, `/health/ready` healthcheck passed and Railway deployment status SUCCESS.
- Web: Docker image build succeeded, `/` healthcheck passed, HTTPS HEAD returned HTTP 200, and browser snapshot showed the live library interface with an empty catalog rather than an upstream API error.
- Web URL: https://openedu.codelocal.cloud/
- PostgreSQL volume is healthy; no production student records or copyrighted textbook files have been seeded.
- Initial deploy failed because a SQL comment contained a semicolon. Fixed with comment-aware statement splitting and regression tests in commit `9e27650`.

**Scope:** this is a hosted foundation, NOT a fully implemented teaching, school administration, or parent communication product. Do not announce production-grade handling of real child records until account, permissions, audit and data protection work is complete.

## Verified release — 2026-10-15, version 0.5 (reader progress)

- GitHub `dev` commit: `c15e004346fb56beb94801920ded4ae89147de96`.
- Railway API and Web: both latest deployments **SUCCESS** at the matching Git commit; health checks passed.
- HTTPS custom domain: **https://openedu.codelocal.cloud/**, Railway domain list shows `ACTIVE`, target port 8080.
- Verified HTTPS HEAD: `/` returned HTTP/2 **200**, `/login` returned HTTP/2 **200**.
- The former Railway-generated `web-production-e29d5.up.railway.app` hostname now returns Railway edge fallback 404 and **must not be shared**. The verified custom domain is the public canonical address.
- New additive migration `0006_reader_progress` is applied automatically by the Go API on startup before readiness (the API started successfully). No public API domain was created; mobile production use still requires approved HTTPS API ingress.
- Disposable PostgreSQL integration verified isolated personal bookmarks/completion, other-school rejection and rights removal on book archive. Web TypeScript / production build and Flutter widget tests passed; re-run iOS Simulator build for the new Flutter reader before native release.
- No demo student accounts, personal educational records, unlicensed textbook data or default administrator were created.

## Auth rollout

Identity v0.2 adds a new append-only migration and requires explicit one-time initial administrator provisioning. See [AUTH_SETUP.md](AUTH_SETUP.md). No default account is created by deploying the code. Test child information must remain off production.

## Production warnings

- Password login exists but verified account recovery, MFA, initial approved production admin provisioning and child-data compliance are incomplete. This is not a launch-ready educational product.
- Do not upload copyrighted textbook files or student PII into a demo.
- Postgres is managed by Railway; do not reinitialize/destroy its volume to debug API startup.
- A proper migration/versioning policy and data-backup drills are required before production use.