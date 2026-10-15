# OpenEdu identity and school onboarding (v0.2 foundation)

## Implemented

- Passwords are salted bcrypt (cost 12), 12-72 bytes, never stored in plaintext.
- Active user, verified JWT signature and active DB role/school membership are checked server side on every privileged request.
- 15-minute HttpOnly, Secure and SameSite=Strict web session cookie.
- Shared PostgreSQL login lockout after repeated failed attempts (per normalized email hash).
- Auth API: POST /v1/auth/login, GET /v1/auth/me.
- Platform management: GET/POST /v1/admin/schools, POST /v1/admin/schools/:id/members.
- Responsive /login and /dashboard; no public signup for minors or guardians.
- 0002_identity migration is append-only and applied automatically by API at startup.
- Real disposable PostgreSQL integration suite validates login, isolated school catalog, immediate role revocation and platform vs school-admin permissions.

## First platform administrator

There are intentionally no default accounts, passwords, or student data. This deployment has no interactive administrator until an operator supplies a real email and a strong private password.

The bootstrap binary is compiled into the API image but is NEVER run automatically. Before creating a user, the operator must set in the Railway **api** service private environment variables:

- BOOTSTRAP_ADMIN_EMAIL = desired admin email.
- BOOTSTRAP_ADMIN_PASSWORD = a new random password (12–72 bytes).
- DATABASE_URL is already a private Railway PostgreSQL reference.

Then run from a trusted authenticated Railway CLI terminal:

```sh
railway ssh --service api /openedu-bootstrap
```

It refuses to run if any platform administrator already exists. It uses a transaction and PostgreSQL advisory lock. **Immediately delete** both BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD from Railway service settings after successful provisioning; do not paste them into Git, issue trackers, or shell history. Sign in at `/login`, then create a school and its school-admin account at `/dashboard`.

A second bootstrap is prohibited; an existing platform administrator must handle additional identities through an approved administrative flow.

## Scope and security limitations

- Do not onboard actual students/guardians until consent, account recovery, 2FA for administrators, guardian verification, event audit, retention policies, and school-specific privacy tests are complete.
- For multi-school accounts, login requires a school UUID to select the correct scope.
- Password reset/change and session revocation lists are not yet implemented; browser logout clears the local cookie, but issued JWTs expire after at most 15 minutes. Account/role revocation is checked directly in the database for every request.
- School member creation is restricted to platform admins for now; it should transition to verified invitation flows rather than direct password issuance.
- Configure strict TLS, distributed IP-based login rate limiting, automated credential auditing and secrets rotation before broad public use.
- Do not copy test users or unlicensed textbook content into production.

## Integration tests (disposable Postgres only)

The test suite rejects arbitrary DB names and requires the name `openedu_test` in the DSN:

```sh
cd services/api
OPENEDU_TEST_DSN='postgres://openedu_test:local_test_only@127.0.0.1:55489/openedu_test?sslmode=disable' go test ./internal/httpapi -run TestIdentityAndSchoolIsolationPostgres -count=1 -v
```

The disposable container used for verification must be stopped after tests; never change the live Railway Postgres to run fixture tests.
