# OpenEdu development instructions

1. Client-review Notion document is proposed scope; never claim client acceptance without confirmation.
2. Work on dev. Commit only verified changes; never force push, override others, or merge untested code into main.
3. Enforce school/class/guardian authorization server side. Never accept role or tenant identity from untrusted request headers.
4. Treat child data as sensitive; do not log PII, copy copyrighted textbooks, or commit credentials.
5. Version API contracts, append database migrations, write tests for tenant isolation, role checks, publication workflow and concurrency.
6. Follow modular boundaries: pure domain + application + adapters. Avoid shared mutable persistence when extracting services.
7. Report implemented vs foundation vs planned status honestly. Documentation alone is not a tested capability.
8. Frontend is shadcn/ui-style local source components + Tailwind CSS 4 + Radix UI + Lucide React, with iOS Liquid Glass styling (docs/DESIGN_SYSTEM.md). Do not add a second icon pack, emoji icons, Unicode pictograms, or ad hoc icon fonts.
9. Import named icons from lucide-react only; standard size 20px, strokeWidth 1.9, small 16px, large 24px. Decorative SVG icons must be aria-hidden; icon-only controls must have accessible names.
10. Reuse apps/web/components/ui/* (shadcn-style source files), lib/utils.ts cn(), and Radix primitives before inventing new styled elements. Do not ship non-functional icon controls.
11. Touch targets >=44px, no horizontal page overflow, focus-visible keyboard outlines, reduced motion and readable glass fallback.
12. Verify apps/web: npm run lint && npm run icons:check && npm run build. Verify services/api: go test ./... && go vet ./.... Then git diff --check. Commit to dev and confirm Railway latest SHA SUCCESS before claiming deployment.
13. **Mandatory skill for any frontend change:** read `skills/openedu-ui/SKILL.md`. Keep `CLAUDE.md` consistent with `AGENTS.md`.
