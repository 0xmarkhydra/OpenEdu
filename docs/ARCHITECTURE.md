# ADR 0001: Modular Go core, extraction-ready

Adopt OpenRide's explicit bounded contexts, stable wire contracts, testable domain logic and service data ownership, but start with a modular monolith instead of premature microservices. Reference: https://github.com/codelocal-cloud/OpenRide/blob/main/docs/MICROSERVICES_ARCHITECTURE.md

## Core boundaries
- Identity: accounts, verified school memberships, sessions, guardian links and credentials.
- Schools: years, classes, teacher assignments, rosters, timetables.
- Library: curriculum, book metadata, chapters, lessons, copyright and publication review.
- Learning: exercises, assignments, submissions, grading and progress.
- Attendance: presence and leave requests.
- Communication: announcements, messaging, surveys, push/email.
- Media: scanned encrypted objects and authorized time-limited downloads.
- Reporting: tenant-scoped aggregate views.

## Technology
Go API, PostgreSQL, versioned OpenAPI, Next.js web, Flutter mobile. S3-compatible licensed object store; Redis for eventual cache and rate limiting. NATS JetStream with outbox/inbox and service-owned databases ONLY when measured scale and ownership require independent deployment. No cross-service database access when split.

## Critical constraints
A school is a tenant. Authenticated tokens carry verified identity/school scope; no trust in client-supplied school ID or roles. Teachers only see assigned classes. A guardian must have a verified link to a child. PostgreSQL queries AND application policy must scope private records. Consider PostgreSQL RLS when tenant-context transaction management is proven. Never expose student PII through public catalog.

Content release state: draft -> review -> approved -> published -> archived. Rights owner, source, license and distribution scope verified BEFORE publication. Do not embed copyrighted books by default.

## Scaling strategy
Stateless API horizontal replicas, tenant-aware indexes, read replicas as needed, CDN/object streaming for licensed content, async indexing/notifications, observability (structured logs and OpenTelemetry), backups, disaster recovery, load and security tests. Keep contracts stable for extraction.
