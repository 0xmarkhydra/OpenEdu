# OpenEdu — iOS Liquid Glass design system v1.1

## Mandatory frontend stack

Use **shadcn/ui-style source components + Tailwind CSS v4 + Radix UI + Lucide React**. Package versions are locked by apps/web/package-lock.json and shadcn config lives in apps/web/components.json.

- Source-owned components: apps/web/components/ui/button.tsx, card.tsx, tabs.tsx; utility lib/utils.ts cn(). Extend local components instead of mixing libraries.
- One icon family: lucide-react, individual named imports, consistent size 20px, stroke width 1.9 (small 16px, hero 24px). Decorative icons aria-hidden; interactive icon-only controls require aria-label.
- Forbidden as UI icons: emoji, Unicode pictograms and arrow characters used as icon substitutes. Do not bring Material Icons, Font Awesome, Phosphor, icon fonts, or random inline SVG for standard actions.
- Canonical mappings: brand BookOpenCheck; books LibraryBig / BookOpen; lesson NotebookPen; family UsersRound; school School; attendance ClipboardCheck; schedule CalendarDays; announcement Bell; authentication LogIn; back ArrowLeft; next ArrowRight.
- Existing translucent iOS glass CSS is maintained as compatibility while new components use shared primitives and Tailwind. Run npm run icons:check for legacy icon regression audit.

## Visual language

## Visual language

OpenEdu uses an iOS-inspired glass interface without claiming to be an Apple product or reproducing proprietary UI assets. Use the same principles on the public library, sign-in, school admin, teacher and parent panels.

- Palette: soft periwinkle #f2f7ff, frost #f8f7ff, vivid blue #3970ed, violet #7d5eea and deep ink #14223e.
- Glass surface: translucent white rgba(255,255,255,.69), 22–25px backdrop blur, 1px reflective hairline borders, inset top highlights, restrained 20px soft shadow.
- Shape: 13–15px controls, 19–29px panels; generous 44px+ touch targets.
- Typography: native -apple-system/SF Pro where installed, Inter/system-ui fallback; prominent weight + tight display letter spacing.
- Header: sticky glass navigation, no layout jump. Nav entries are real links, not decorative fake features.
- Responsive: 4 statistic tiles desktop, 2 mobile; 2-column portal collapses to 1; tab bar scrolls horizontally; search wraps on small screens.
- Motion: short 0.17–0.25 second lift and brightness transitions only. Honor prefers-reduced-motion: reduce; avoid shaking, bouncing, parallax or continuous animation.
- Accessibility: visible keyboard focus, explicit labels, live error/status messages, text contrast above pale backgrounds; graceful fallback when backdrop-filter unsupported.
- Empty states must reflect the database accurately. No fabricated books, school records or metrics.

### Implementation

The canonical tokens and material rules are in `apps/web/app/globals.css` under OpenEdu Design System.
The responsive school dashboard is `apps/web/app/dashboard/school-panel.tsx`.
The verified-parent privacy view is `apps/web/app/dashboard/parent-panel.tsx`.
Use these CSS classes consistently; do not add conflicting gradients or random icon fonts.

## Icon migration audit (2026-10-15)

- Public header, library, login, book outline/lesson reader, platform dashboard, school dashboard, parent portal and editorial navigation migrated to named Lucide components.
- All existing decorative emoji and Unicode glyph icons in web TSX replaced; navigation no longer has duplicate links to the same role-neutral destination.
- School section tabs now use accessible Radix Tabs, supporting keyboard navigation and proper tabpanels. Shared shadcn-compatible Button/Card primitives are used in public and administrative screens.
- CI/deployment enforcement: `npm run build` first runs `npm run icons:check`, rejecting legacy pictograms, raw inline SVG for commonplace icons, and mixed icon styles. `npm run verify` checks TypeScript plus build.
- Frontend package installation, TypeScript type-check, icon audit and production Next build verified locally; production browser/mobile device verification must be recorded separately before declaring visual signoff.

### Additional validation required

Visual QA on mobile Safari and desktop Safari, keyboard audits, low-power GPU sampling, accessibility contrast audit, real user tests and a real iOS Flutter design system are still outstanding.
