# OpenEdu implementation roadmap

The linked Notion document is a proposal pending customer approval.

## Stage 0 — executable foundation
- [x] Backend scaffold and architecture decision
- [x] Baseline school + curriculum schema
- [x] Published catalog API and privileged book authoring/publishing foundation
- [x] Trusted JWT verification boundary
- [x] Password-based login, short-lived JWT token issuance and membership revalidation (Go + PostgreSQL)
- [x] Platform admin school and member provisioning APIs
- [x] Authenticated password change with immediate JWT revocation and security event
- [ ] Password reset, verified invitations, two-factor auth and audited account recovery
- [x] DB-backed integration testing with disposable PostgreSQL (identity, school operations, guardian isolation, editorial lifecycle)
- [x] Next.js responsive public library pages scaffolded (build pending)
- [x] Web login and platform administration dashboard (bootstrap account setup still required)
- [x] Flutter iOS/Android shared UI foundation (widget tests + iOS Simulator build)
- [ ] Role-complete native dashboards, Android SDK build, app store release
## Stage 1 — education flows
- [x] Platform admin school creation/list and school member account provisioning
- [x] Academic year/class creation, student enrolment and teacher assignments
- [ ] Semester edit/archive, transfer and bulk import
- [x] Tenant-scoped role checks and verified parent-child relationships
- [ ] Full RBAC role-matrix audit and consent verification
- [x] Text chapter/lesson reader with school-publication scope, safe text rendering
- [x] Per-user bookmark and completion progress for licensed published text lessons (web + Flutter)
- [ ] PDF reader, offline rights, cross-device progress analytics
- [x] Assignments, text submission, grading and feedback
- [x] Timed, one-attempt multiple-choice tests with server-side grading (teacher drafting, publishing, student submissions, scores)
- [ ] Reusable question bank, file attachments, exam rubrics, retakes and exam accommodations
- [x] Class attendance and student/verified-parent leave request approval
## Stage 2 — full portal
- [x] Responsive shared student/teacher/parent portal and class timetables
- [x] Tenant-scoped school aggregate reports (enrollment, attendance, assignments and leave)
- [ ] Detailed analytics, timetable conflict rules
- [x] Role-targeted school announcements
- [x] Verified teacher–guardian messages with read receipts and immediate relationship revocation
- [ ] Surveys, calendar events, menus and push notifications
- [x] CMS draft chapters and lessons plus approve/publish workflow
- [ ] Full content version history, media storage and license verification
- [x] Flutter iOS Simulator-tested client foundation; school-level aggregate web reports
- [ ] Native Android build, push and role-complete analytics

Each feature needs tenant/role integration tests, OpenAPI contract, audit/PII review, accessibility checks, backup and staged deployment evidence prior to launch.
