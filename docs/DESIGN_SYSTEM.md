# OpenEdu — iOS Liquid Glass design system v1

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

### Additional validation required

Visual QA on mobile Safari and desktop Safari, keyboard audits, low-power GPU sampling, accessibility contrast audit, real user tests and a real iOS Flutter design system are still outstanding.
