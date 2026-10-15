# OpenEdu release gates — verified scope and remaining 100% blockers

## Achieved during 0.4 implementation
- Private classroom messaging between verified guardian and assigned teacher, available on web; message contacts, send/list and read receipts. Revoking guardian link immediately hides historical messages.
- Authenticated password rotation, bcrypt update, incrementing DB auth_version in the same transaction, immediate rejection of old JWT sessions and a minimal account security event.
- Aggregate school report for leadership only (attendance records, class sizes, assignments, submissions, pending leave). No export of student personal information.
- Flutter iOS/Android project with permitted-library browsing, lesson text, per-account bookmark/completion, JWT login via secure device storage, class/announcement viewing, private-message compose and reading; Flutter widget tests and unsigned iOS Simulator build previously passed; revalidate iOS compilation for each reader change.
- Go integration tests run against a disposable PostgreSQL test database; Next.js TypeScript, icon scan and production build checked.

## Still blocking full completion
1. **Legal/privacy/child safety:** consent, independently verified guardians, data subject rights, deletion and retention workflows, privacy review. No real student onboarding before approval.
2. **Security:** email-verified password recovery, MFA/2FA, device/session inventory, admin action audit coverage, pentest, IP/device throttling, tenant boundary fuzzing, backup/restore drills.
3. **Content & media:** private licensed S3/PDF flow, signing/encryption and download rights, preview, chapter/file versioning, licensed-content verification and takedown process.
4. **Learning:** single-attempt timed quizzes and server-graded single-choice questions now pass PostgreSQL tests; still missing reusable bank, attachments, essay rubric/customization, exam accommodations/retakes, progress analytics, attendance automation and schedule conflicts.
5. **Communication:** push/email delivery, urgent alerts, events, read receipts for announcements, moderation/escalation workflow, chat retention policy.
6. **Native:** Android SDK missing on dev Mac, native Android build not verified, mobile API public URL intentionally not enabled, real-device QA and signing/store submissions incomplete.
7. **Production onboarding:** no default admin or school data. Customer/operator must authorize first-admin email/password and setup; verify live auth/school flows without storing children’s information.
8. **Operations:** CDN/object media policy, on-call incident runbook, horizontal load tests, monitoring/alerting and rollbacks, production database backup recovery exercise.

Full product signoff requires explicit evidence for each gate and acceptance from the school/business owner. **A successful Docker deployment, passing local tests or a UI mock is not equivalent to a feature-complete or compliance-ready product.**

## Safe rollout
Keep migration order 0004_messaging.sql followed by 0005_auth_version.sql; **do not rewrite an already applied migration**. The auth version rollout invalidates previously signed JWTs that lack an auth_version claim, requiring re-login. This is expected.

After pushing to dev, check Railway API and Web deployment SHA/status, migration startup logs and read-only page smoke tests. Do not provision live accounts or add private school records without explicit authorization.

### Quiz module evidence (v0.6)
- Class-scoped exam creation, question authoring, publish, single student attempt, server-enforced deadline and single-choice grading have disposable PostgreSQL integration tests.
- Negative tests cover draft answer secrecy, incomplete answers, second submission, restart deadline, cross-school student, guardian access and expired attempt. Verified guardians can read only submitted child scores, and revocation immediately denies access; school leadership has aggregate (non-PII) quiz reporting.
- Web has role-aware iOS-style dashboard Quiz tab. Flutter adds a class-to-exams student UI; native widget tests and iOS Simulator build pass for this change. Private mobile API domain, Android build and real-device acceptance remain rollout prerequisites. Successful local tests do not authorize storing real student exam data.
