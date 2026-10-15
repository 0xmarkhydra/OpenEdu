---
name: openedu-ui
description: OpenEdu design system standard. Mandatory when an agent adds, edits or reviews any UI, icon, component or mobile layout.
---

# OpenEdu UI agent skill

## Read before editing
- `AGENTS.md` and `docs/DESIGN_SYSTEM.md`.
- `apps/web/components.json`, `apps/web/components/ui/*`, `apps/web/lib/utils.ts`.
- Existing relevant page component and `apps/web/app/globals.css`.

## Fixed technology decisions
1. React/Next.js + Tailwind CSS v4 for styling, local **shadcn/ui-style components**, Radix UI primitives for accessible interactions, and **Lucide React as the only icon family**.
2. Import icons by name, for example `import { BookOpen, ArrowRight } from "lucide-react"`. Default `size={20} strokeWidth={1.9}`, small 16, large 24. Use icons that communicate their action and keep consistent visual weight across pages.
3. No emoji, standalone Unicode symbols, arrows in place of SVG icons, font-awesome, Material Icons, mixed packs or ad hoc inline SVG for commonplace UI actions. No external fonts or icon web requests.
4. Prefer `Button`, `Card`, `Tabs`, `cn()` from the shared source-owned UI. If a new Radix primitive is needed, add a source-owned wrapper in `components/ui/`.
5. iOS-inspired Liquid Glass = translucent frost white panels, thin reflective borders, soft indigo/blue-violet gradients, 13–29px radius, subtle shadows, controlled `backdrop-filter`. It must be legible without filter support.
6. Use real functionality; do not add fake links, blank menu routes, inactive icon buttons or mock counters as live data.
7. Mobile navigation must not overflow. Touch targets at least 44px. Explicit input labels, focus states, keyboard-friendly tabs and reduced-motion behavior are mandatory.
8. Privacy guardrails: never embed student records, private tenant details or copyrighted textbooks in screenshots or UI fixture data.

## Checklist before committing a UI change
- Inspect route and surrounding component library.
- Reuse shared icon names and sizes, no emoji glyphs.
- Check small phones (320px), 390px, tablet, desktop at 1280px. The top bar and tab bar must remain usable.
- `cd apps/web && npm ci && npm run verify`; the build runs icon audit by default.
- `cd services/api && go test ./... && go vet ./...` if API/contracts changed.
- Record changed screens and evidence in `docs/DESIGN_SYSTEM.md` and task PR/commit.
- Push to `dev`, then verify Railway Web deployment SHA = pushed SHA and status SUCCESS; inspect public landing and login in browser.
- Do not claim 100% functional product completeness from visual audits; call out missing product features separately.

## Standard icon mapping
| Surface | Lucide |
|---|---|
| Brand | BookOpenCheck |
| Library/catalog | LibraryBig / BookOpen |
| Classroom | School |
| Students / parents | UsersRound |
| Teachers | GraduationCap |
| Assignments | NotebookPen |
| Attendance | ClipboardCheck |
| Timetable | CalendarDays |
| Announcements | Bell |
| Leave requests | CalendarClock |
| Login | LogIn |
| Logout | LogOut |
| Navigate back / forward | ArrowLeft / ArrowRight |
| Success / reject | CircleCheck / CircleX |

## Existing implementation
- `apps/web/app/page.tsx`: public catalog iconized cards.
- `apps/web/app/layout.tsx`: consistent Lucide branding and responsive global navigation.
- `apps/web/app/login/page.tsx`: login with Button and Lucide.
- `apps/web/app/dashboard/school-panel.tsx`: Radix accessible tabs and role-scoped operations with Lucide.
- `apps/web/app/dashboard/parent-panel.tsx`: parent portal.
- `apps/web/app/dashboard/content`: content studio and editorial routes.
- `apps/web/scripts/check-ui-icons.mjs`: automatic icon regression check.
