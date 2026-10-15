# OpenEdu development instructions

1. Client-review Notion document is proposed scope; never claim client acceptance without confirmation.
2. Work on dev. Commit only verified changes; never force push, override others, or merge untested code into main.
3. Enforce school/class/guardian authorization server side. Never accept role or tenant identity from untrusted request headers.
4. Treat child data as sensitive; do not log PII, copy copyrighted textbooks, or commit credentials.
5. Version API contracts, append database migrations, write tests for tenant isolation, role checks, publication workflow and concurrency.
6. Follow modular boundaries: pure domain + application + adapters. Avoid shared mutable persistence when extracting services.
7. Report implemented vs foundation vs planned status honestly. Documentation alone is not a tested capability.
8. Verification: cd services/api && go test ./... ; git diff --check.
