# OpenEdu — functional delivery matrix (2026-10-15)

This is a development checkpoint, not a statement that the full proposal is finished.
Working means automated code checks **and** (where indicated) disposable PostgreSQL integration tests. Production deployment requires separate live build and smoke test evidence.

| Area | Implemented in code | Validated | Still missing |
| --- | --- | --- | --- |
| Identity & RBAC | Login, bcrypt, short-lived JWT, DB role revalidation, self-service password change with immediate JWT revocation, platform admin provisioning, throttling | Unit + Postgres integration | Password reset, 2FA, verified invitations, consent, full audit |
| Multi-school | Add/list schools, members, academic years, classes, student enrolment and teacher assignment | Postgres | Bulk import, transfers, semester administration, advanced teacher scheduling |
| Attendance | Teacher records attendance by day, reads class history | Postgres | Mass entry, analytics, guardian alerts |
| Assignment & quiz | Text assignments, teacher grading and feedback; server-timed single-choice quizzes with protected answer keys, one attempt, automatic grade and teacher results | Disposable PostgreSQL + web build | Reusable question bank, file attachments, essay rubrics, approved retake/accommodations, proctoring |
| Leave requests | Student or verified guardian submits, teacher approves/rejects | Postgres | Notification to teachers and calendar linkage |
| Timetable | Add/update period by weekday and view per authorized class | Postgres | Time collisions, substitute teacher, school-wide schedule |
| Announcements | Target all/teacher/student/guardian within school | Postgres | Push notifications, per-announcement read receipts |
| Parent portal | Verified guardian links, child assignment grades and submitted quiz scores, attendance, leave and secure teacher messaging/read receipts | Postgres integration | Multi-guardian audit and consent/recovery procedures, mobile parent detail workflows |
| Licensed library | Catalog, chapter/lesson editor, text reader, publication lifecycle, per-account bookmark/completion progress on web and Flutter | PostgreSQL integration + local Next.js/Flutter tests | PDF/media upload, offline rights, legal license verification and reader progress analytics |
| Web UX | Public catalog, login, platform admin, school dashboard, iOS-style Liquid Glass responsive UI | Next.js build | Safari device visual QA, accessibility audit |
| Mobile | Flutter native library/search, lesson text reader/progress, login with secure storage, classes, announcements, messaging and student timed quiz attempt | Flutter widget tests + iOS Simulator compile | Mobile HTTPS API publication, Android SDK build, App Store/Play review, complete role features and native teacher quiz editor |
| Infrastructure | Go API, PostgreSQL append-only migrations, versioned JWT revocation, Railway GitHub dev deployments | Verified baseline + local Postgres tests | Backups drill, monitoring alerts, signed media URLs, production load/security review |

Live student/guardian information must not be entered until privacy and compliance review is complete. Do not seed unlicensed textbooks. This file must be updated after each milestone, including any failing tests or deployment issues.

## Rollout order

1. Local PostgreSQL integration + web build.
2. Commit and push to dev. Railway GitHub source must point at 0xmarkhydra/OpenEdu and the correct root directories.
3. Confirm both API and Web latest deployment SUCCESS and health checks pass.
4. Browser smoke test: login renders, dashboard redirects without session, library public route works, no server exceptions.
5. Production admin account must be provisioned out-of-band with separate secret approval (see AUTH_SETUP.md); no default credentials.
6. Build reader/CMS, richer assessment, complete guardian features, mobile, privacy and scalability next.
