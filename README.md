# FreeEducation Core

Scalable education platform: licensed digital library, assessment, student/teacher portals and school-family communication.

**Status:** early implementation. Scope: [Notion proposal for client review](https://app.notion.com/p/3f37dabb721d81549ec8d5670a5d5ac7). Proposed items are NOT customer-approved commitments.

## Hosted foundation

- Public web: https://web-production-e29d5.up.railway.app/
- Railway project: https://railway.com/project/f4589b3e-ba78-49a6-917a-08ba27b1f77b
- GitHub deployments track branch `dev`; `api` is private inside the Railway project. See `docs/RAILWAY_DEPLOY.md` for verified status.
- The catalog intentionally starts empty until legally licensed resources have been approved and published. Student, school admin and parent flows are not launched yet.

## Architecture
Go 1.26 HTTP API, PostgreSQL 16, Next.js/TypeScript web, planned Flutter mobile, S3 object storage for licensed files. Modular-monolith first with extraction-ready contexts. See docs/ARCHITECTURE.md and docs/PROJECT_STATUS.md.

## Quickstart
- docker compose up -d postgres
- On API startup, the PostgreSQL baseline migration runs transactionally under an advisory lock. Configure DATABASE_URL to a PostgreSQL 16 instance (example in .env.example).
- Set DATABASE_URL and a JWT_SECRET of at least 32 bytes; cd services/api && go mod tidy && go run ./cmd/api
- Visit http://localhost:8080/health/live and /health/ready
- Tests: make api-test
- Web: cd apps/web && npm install && npm run dev (http://localhost:3000). The web app renders real published catalog API data, never fabricated textbook records.
- API contract: packages/contracts/openapi.yaml
- Identity + school management: /login and /dashboard, see docs/AUTH_SETUP.md. No default admin exists; it requires explicit one-time operator provisioning.
- Docker images: services/api/Dockerfile and apps/web/Dockerfile; web dependency lockfile and CI build still pending.

The public catalog only serves published public books. Privileged operations require a verified JWT from a trusted issuer. Identity login / issuing production tokens is NOT implemented in this initial slice.

## Safety
No default passwords. Do not commit student information, textbook PDFs, license keys or private env vars. School authorization must be enforced at query and application layers.

Development branch: dev. Test before merging into main. License for project not yet decided; no third-party textbook redistribution is permitted by default.
