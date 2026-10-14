# FreeEducation implementation roadmap

The linked Notion document is a proposal pending customer approval.

## Stage 0 — executable foundation
- [x] Backend scaffold and architecture decision
- [x] Baseline school + curriculum schema
- [x] Published catalog API and privileged book authoring/publishing foundation
- [x] Trusted JWT verification boundary
- [x] Password-based login, short-lived JWT token issuance and membership revalidation (Go + PostgreSQL)
- [x] Platform admin school and member provisioning APIs
- [ ] Password reset, verified invitations, two-factor auth and audited account recovery
- [ ] DB-backed integration testing with real PostgreSQL
- [x] Next.js responsive public library pages scaffolded (build pending)
- [x] Web login and platform administration dashboard (bootstrap account setup still required)
- [ ] Mobile Flutter, teacher/student/parent-specific dashboards
## Stage 1 — education flows
- [x] Platform admin school creation/list and school member account provisioning
- [ ] Academic year/semester/class CRUD, enrollment and teacher assignments
- [ ] RBAC permissions matrix, verified parent-child relationships
- [ ] Library reader, bookmarks, progress, offline rights
- [ ] Assignments, submission, grading and feedback
- [ ] Attendance and parent leave request approval
## Stage 2 — full portal
- [ ] Student/teacher/parent dashboards and timetables
- [ ] Notifications, messaging, surveys, events, menus
- [ ] CMS moderation and versioned publishing
- [ ] Flutter client and push, reporting/analytics

Each feature needs tenant/role integration tests, OpenAPI contract, audit/PII review, accessibility checks, backup and staged deployment evidence prior to launch.
