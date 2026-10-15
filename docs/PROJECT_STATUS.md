# OpenEdu status — 2026-10-15

Status legend: IMPLEMENTED = executable and tested; FOUNDATION = meaningful code with missing full integration; PLANNED = absent.

| Capability | Status | Notes |
| --- | --- | --- |
| Go API and PostgreSQL schema | FOUNDATION | Railway Docker build succeeded; startup migration applied and /health/ready succeeded; Go tests and vet passed |
| Identity/login | FOUNDATION | Go login + bcrypt, JWT issuance, DB role revocation and HTTP-only web cookie tested locally; first production admin not provisioned, recovery/2FA pending |
| Tenant-safe library catalog | FOUNDATION | published listing and basic mutation, more tests needed |
| School management and RBAC | FOUNDATION | school/year/class/roster/teacher/timetable APIs and dashboards, guardian verification and scoped access tested; bulk import and audit pending |
| Text lesson reader and editorial CMS | FOUNDATION | chapter/lesson authoring, moderation and safe JSON-backed text reader pass PostgreSQL integration; PDF/object delivery and licensing review remain pending |
| Learning/assignments/grading | FOUNDATION | teacher assignments, submissions, scores and feedback pass disposable PostgreSQL tests; file attachments, exams/rubrics missing |
| Attendance and leave | FOUNDATION | teacher attendance, student and verified guardian leave request, approval/rejection pass PostgreSQL integration |
| Communication and notifications | FOUNDATION | school announcements and authorized teacher-guardian messages with read receipts pass PostgreSQL integration; push, surveys and real-time delivery missing |
| Web portal / iOS glass design | FOUNDATION | responsive public library, lesson reader, login, admin/CMS, teacher/student/guardian pages build locally; Safari device QA and production deployment pending |
| Flutter mobile | FOUNDATION | Flutter project built for iOS Simulator; shared UI for library, lessons, login, classes, announcements and messages; Android SDK and mobile API hostname pending |
| Production security and load testing | PLANNED | JWT auth version and password-change revocation tested on disposable PostgreSQL; formal pen-test, 2FA, backups and load tests not validated |

Do not infer production readiness from repository structure alone.
