# OpenEdu Core

Scalable education platform: licensed digital library, assessment, student/teacher portals and school-family communication.

**Status:** early implementation. Scope: [Notion proposal for client review](https://app.notion.com/p/3f37dabb721d81549ec8d5670a5d5ac7). Proposed items are NOT customer-approved commitments.

## OpenEdu interface and agent guidelines

The frontend uses **local shadcn/ui-style components + Tailwind CSS v4 + Radix UI + Lucide React** with an iOS-inspired Liquid Glass theme. Icon rules and screen guidelines are mandatory for every UI agent: read `AGENTS.md`, `CLAUDE.md`, `skills/openedu-ui/SKILL.md` and `docs/DESIGN_SYSTEM.md`.

`cd apps/web && npm ci && npm run verify` runs TypeScript, a full icon glyph audit and production build. The build fails when legacy emoji/pictogram icons are introduced in TSX. Use named Lucide imports and reusable local UI components.

## Hosted foundation

- Public web: https://openedu.codelocal.cloud/
- Railway project: https://railway.com/project/f4589b3e-ba78-49a6-917a-08ba27b1f77b
- GitHub deployments track branch `dev`; `api` is private inside the Railway project. See `docs/RAILWAY_DEPLOY.md` for verified status.
- The catalog intentionally starts empty until legally licensed resources are approved and published. Web login, school operations, parent portal and text lesson reader are implemented in code; **operator provisioning, privacy review and a verified new production release** are required before real student use.

## Architecture
Go 1.26 HTTP API, Railway PostgreSQL 18, Next.js/TypeScript web, Flutter iOS/Android app foundation, and planned S3-compatible licensed media storage. Modular-monolith first with extraction-ready contexts. See docs/ARCHITECTURE.md and docs/PROJECT_STATUS.md.

## Quickstart
- docker compose up -d postgres
- On API startup, the PostgreSQL baseline migration runs transactionally under an advisory lock. Configure DATABASE_URL to a supported PostgreSQL instance (example in .env.example).
- Set DATABASE_URL and a JWT_SECRET of at least 32 bytes; cd services/api && go mod tidy && go run ./cmd/api
- Visit http://localhost:8080/health/live and /health/ready
- Tests: make api-test
- Web: cd apps/web && npm install && npm run dev (http://localhost:3000). The web app renders real published catalog API data, never fabricated textbook records.
- API contract: packages/contracts/openapi.yaml
- Identity + school management: /login and /dashboard, see docs/AUTH_SETUP.md. No default admin exists; it requires explicit one-time operator provisioning.
- Docker images: services/api/Dockerfile and apps/web/Dockerfile; Next.js production build and lockfile exist. See docs/FEATURE_MATRIX.md and docs/DESIGN_SYSTEM.md.

The public catalog serves published licensed public books. School-scoped books are accessible only with verified school membership. Identity login issues short-lived tokens with live role revocation. No account is auto-created in production.

## Safety
No default passwords. Do not commit student information, textbook PDFs, license keys or private env vars. School authorization must be enforced at query and application layers.

Development branch: dev. Test before merging into main. License for project not yet decided; no third-party textbook redistribution is permitted by default.

## 0.4 progress — communications and native app foundation

The current development branch adds tenant-authorized teacher/guardian messaging with read receipts, school aggregate reporting, authenticated password rotation with immediate JWT revocation, and a native Flutter app foundation (iOS Simulator previously compiled; Android SDK not configured). Read `docs/API_OPERATIONS.md`, `apps/mobile/README.md`, and **`docs/RELEASE_GATES.md`**. This is not yet a fully launched school platform.
