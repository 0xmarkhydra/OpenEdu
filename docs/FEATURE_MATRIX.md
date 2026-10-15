# OpenEdu — functional delivery matrix (2026-10-15)

This is a development checkpoint, not a statement that the full proposal is finished.
Working means automated code checks **and** (where indicated) disposable PostgreSQL integration tests. Production deployment requires separate live build and smoke test evidence.

| Area | Implemented in code | Validated | Still missing |
| --- | --- | --- | --- |
| Identity & RBAC | Login, bcrypt, short-lived JWT, DB role revalidation, platform admin provisioning, limited failed login attempts | Unit + Postgres | Password reset, 2FA, verified invitations, audit and consent |
| Multi-school | Add/list schools, members, academic years, classes, student enrolment and teacher assignment | Postgres | Bulk import, transfers, semester administration, advanced teacher scheduling |
| Attendance | Teacher records attendance by day, reads class history | Postgres | Mass entry, analytics, guardian alerts |
| Assignment | Teacher posts assignment, student submits, teacher grades and provides feedback | Postgres | Files, rubrics, timed quizzes, anti-cheat |
| Leave requests | Student or verified guardian submits, teacher approves/rejects | Postgres | Notification to teachers and calendar linkage |
| Timetable | Add/update period by weekday and view per authorized class | Postgres | Time collisions, substitute teacher, school-wide schedule |
| Announcements | Target all/teacher/student/guardian within school | Postgres | Push notifications, read receipts, messaging |
| Parent portal | Admin-verified guardian links, child scores/attendance, request leave | Postgres | Two-way chat, multi-guardian audits, consent and recovery |
| Licensed library | Catalog, chapter/lesson editor, text reader, publication lifecycle | PostgreSQL integration + local Next.js build | PDF/media upload, bookmark/progress, offline rights, legal license verification |
| Web UX | Public catalog, login, platform admin, school dashboard, iOS-style Liquid Glass responsive UI | Next.js build | Safari device visual QA, accessibility audit |
| Mobile | Responsive web | Web build | Native Flutter iOS/Android application |
| Infrastructure | Go API, PostgreSQL migrations, Railway services and branch-based deployments | Verified baseline | Backups drill, monitoring alerts, signed media URLs, production load/security review |

Live student/guardian information must not be entered until privacy and compliance review is complete. Do not seed unlicensed textbooks. This file must be updated after each milestone, including any failing tests or deployment issues.

## Rollout order

1. Local PostgreSQL integration + web build.
2. Commit and push to dev. Railway GitHub source must point at 0xmarkhydra/OpenEdu and the correct root directories.
3. Confirm both API and Web latest deployment SUCCESS and health checks pass.
4. Browser smoke test: login renders, dashboard redirects without session, library public route works, no server exceptions.
5. Production admin account must be provisioned out-of-band with separate secret approval (see AUTH_SETUP.md); no default credentials.
6. Build reader/CMS, richer assessment, complete guardian features, mobile, privacy and scalability next.
