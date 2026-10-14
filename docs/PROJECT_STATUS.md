# OpenEdu status — 2026-10-15

Status legend: IMPLEMENTED = executable and tested; FOUNDATION = meaningful code with missing full integration; PLANNED = absent.

| Capability | Status | Notes |
| --- | --- | --- |
| Go API and PostgreSQL schema | FOUNDATION | Railway Docker build succeeded; startup migration applied and /health/ready succeeded; Go tests and vet passed |
| Identity/login | FOUNDATION | Go login + bcrypt, JWT issuance, DB role revocation and HTTP-only web cookie tested locally; first production admin not provisioned, recovery/2FA pending |
| Tenant-safe library catalog | FOUNDATION | published listing and basic mutation, more tests needed |
| School management and RBAC | FOUNDATION | Platform school CRUD/list and member provisioning implemented; PostgreSQL integration tests pass; production rollout pending |
| Books/PDF reader | PLANNED | rights and object delivery not implemented |
| Learning/assignments/grading | PLANNED | not implemented |
| Attendance and leave | PLANNED | not implemented |
| Communication and notifications | PLANNED | not implemented |
| Web library foundation | FOUNDATION | Next.js production build succeeded locally and on Railway; public HTTPS returns 200; catalog shows correct empty state (no licensed books loaded) |
| Flutter mobile | PLANNED | not implemented |
| Production security and load testing | PLANNED | not validated |

Do not infer production readiness from repository structure alone.
