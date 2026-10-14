# FreeEducation implementation roadmap

The linked Notion document is a proposal pending customer approval.

## Stage 0 — executable foundation
- [x] Backend scaffold and architecture decision
- [x] Baseline school + curriculum schema
- [x] Published catalog API and privileged book authoring/publishing foundation
- [x] Trusted JWT verification boundary
- [ ] Trusted login/token issuance, account provisioning and password reset
- [ ] DB-backed integration testing with real PostgreSQL
- [x] Next.js responsive public library pages scaffolded (build pending)
- [ ] Web identity/login UI, account flows, mobile Flutter
## Stage 1 — education flows
- [ ] School/year/semester/class CRUD, enrollment, teacher assignments
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
