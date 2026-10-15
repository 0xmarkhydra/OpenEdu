# OpenEdu 0.3 — operational API guide

## Identity and school isolation

The public `/v1/books` catalog only returns published public material (or a published school book when a valid membership JWT is supplied). **Do not** trust `X-School-ID` headers; school scope comes solely from signed JWT + live DB membership verification.

All school operations use `Authorization: Bearer <token>`. Use the web server's HttpOnly cookie-backed BFF proxy, not browser-managed token storage. Admin, principal and teacher are not interchangeable: teachers must be explicitly assigned to a class for class-specific access. A student only sees their enrolled classes. A verified guardian sees only their child; enrollment and relationship records must both be active. JWTs expire after 15 minutes and account/role status is checked on each protected request.

## School administration and academic plan

| Method | Path | Required role | JSON input for writes |
| --- | --- | --- | --- |
| GET/POST | /v1/schools/{school}/years | read: school member; write: principal, school_admin | `{"name":"2026–2027","starts_on":"2026-09-01","ends_on":"2027-06-30"}` |
| GET/POST | /v1/schools/{school}/classes | read: scoped membership; write: principal, school_admin | `{"name":"6A1","academic_year_id":"UUID","grade":6}` |
| GET/POST | /v1/classes/{id}/students | read: staff assigned to class; write: principal, school_admin | `{"student_id":"UUID"}` |
| POST | /v1/classes/{id}/teachers | principal, school_admin | `{"teacher_id":"UUID","subject_code":"toan"}` |
| GET | /v1/schools/{school}/overview | active school member | No input |
| GET/POST | /v1/classes/{id}/timetable | read: enrolled/assigned/verified; write: principal, school_admin | `{"weekday":2,"period":3,"subject_code":"toan","room":"B101"}` |

Student and teacher accounts must exist and have active school memberships **before** they can be enrolled or assigned. The tenant is always derived/verified on the server. Class lists hide classes unrelated to a student, teacher, or verified guardian.

## Learning, attendance and communication

| Method | Path | Required role | JSON input for writes |
| --- | --- | --- | --- |
| GET/PUT | /v1/classes/{id}/attendance | assigned teacher, school admin/principal | `{"student_id":"UUID","date":"2026-10-15","status":"present","note":""}`; statuses: present/absent/late/excused |
| GET/POST | /v1/classes/{id}/assignments | read: enrolled/assigned/verified; write: assigned teacher or principal/school_admin | `{"title":"Ôn tập","instructions":"Bài làm","due_at":"2026-10-16T12:00:00Z"}` |
| GET/POST/PATCH | /v1/assignments/{id}/submissions | POST enrolled student; PATCH assigned teacher/admin; GET submitter or staff | POST: `{"content":"Bài làm"}`; PATCH: `{"student_id":"UUID","grade":9.0,"feedback":"Tốt"}` |
| GET/POST/PATCH | /v1/classes/{id}/leave-requests | student POST own request; assigned teacher/admin review | POST: `{"start_on":"2026-10-16","end_on":"2026-10-17","reason":"Gia đình có việc"}`; PATCH: `{"request_id":"UUID","status":"approved"}` |
| GET/POST | /v1/schools/{school}/announcements | read: active member with audience filter; write: principal/school_admin | `{"title":"Thông báo","content":"Nội dung","audience":"all"}`; audiences: all/teacher/student/guardian |

GET attendance requires `?date=YYYY-MM-DD`. No automatic scheduling or third-party messaging is implemented. Grade range is 0–10. Unsubmitted work cannot be graded.

## Verified guardian privacy flows

Platform administrators or school administrators must first create the guardian/student membership accounts using approved onboarding and check guardian consent/identity independently. School principals or admins then verify a relationship:

`POST /v1/schools/{school}/guardians` — `{"guardian_id":"UUID","student_id":"UUID"}`.

Verified guardians can use:
- `GET /v1/me/children` — verified child and class list only
- `GET /v1/me/children/{student}/grades` — only that child's submitted assignments
- `GET /v1/me/children/{student}/attendance` — only that child's last 90 days
- `POST /v1/me/children/{student}/leave` — `{"class_id":"UUID","start_on":"2026-10-16","end_on":"2026-10-17","reason":"Gia đình có việc"}`

A guardian without verified linked records receives 403 when asking for unrelated child data, even when they know valid student IDs.

## Licensed educational content

Teachers cannot bypass publication permissions. Only editor roles can create books in draft. A school_admin/content_admin may author a school-scoped book, while a platform_admin can publish public scope; school or platform reviewer approves.

`POST /v1/admin/books` takes `{"title":"Bài giảng có phép","grade":6,"subject_code":"toan","rights_owner":"Chủ sở hữu","rights_license":"Giấy phép hợp lệ","rights_source":"Nguồn quyền","visibility":"school"}`.

`GET /v1/admin/books` lists the current editor's scoped books. For draft books:
- `POST /v1/admin/books/{id}/chapters` with `{"title":"Chương 1","position":1}`
- `GET /v1/admin/books/{id}/chapters` lists chapters for ongoing editing
- `POST /v1/admin/chapters/{id}/lessons` with `{"title":"Bài 1","position":1,"content":"Nội dung văn bản có phép"}`
- Workflow transitions: `POST /v1/admin/books/{id}/submit` → `approve` → `publish` → `archive`, with separate role checks at each step.

Public readers can request `GET /v1/books/{id}/outline` and `GET /v1/lessons/{id}` only when book status is published and the JWT satisfies school scope. The HTML reader escapes source content as text (no unsafe HTML rendering). **PDF upload, signed S3 delivery, offline textbook rights, moderation evidence and content versions are not implemented.**

## Validation and rollout

Automated Go auth tests plus disposable PostgreSQL integration tests cover school operations, guardian authorization, assignment flow, attendance, submission/grade, text reader, and publication state transitions. Frontend is built with Next.js production compiler. These are necessary but not sufficient for handling real minors' data. See `docs/FEATURE_MATRIX.md`, `docs/AUTH_SETUP.md`, and `docs/RAILWAY_DEPLOY.md`.
