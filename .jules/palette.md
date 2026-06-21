## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.
## 2026-06-15 - [Icon-only Button Accessibility and Form Submission Prevention]
**Learning:** Many interactive buttons in complex components like `AIAnalysis.tsx` missed `type="button"`, which could cause accidental form submissions if placed near forms. Additionally, utility buttons with only icons (like the send button or clear search) lacked `aria-label` attributes, impacting screen reader users.
**Action:** Always add `type="button"` defensively to non-submit buttons and ensure explicit `aria-label` attributes on any icon-only interactive element.
## 2026-06-25 - [Accessibility for Modals and Search Filters]
**Learning:** Found multiple instances where utility 'close' or 'clear' icon buttons in modals and search inputs (e.g., `&times;` or `✕`) lacked `type="button"` and `aria-label`. Without these attributes, screen readers cannot communicate the button's function, and they may inadvertently trigger parent form submissions.
**Action:** When creating or maintaining modals, popovers, and search bars, aggressively add `type="button"` and descriptive `aria-label` attributes to any icon-only interactive controls.

## 2024-05-18 - Missing Type and ARIA labels on utility icon buttons
**Learning:** Found a recurring pattern in React components where icon-only buttons (like those enclosing `svg` icons without text, e.g., in `App.tsx`, `MarketingDashboard.tsx`, `Planner.tsx`, `WhatsAppChat.tsx`, and `SettingsTab.tsx`) lacked explicit `type="button"` and semantic `aria-label`s. This lack of explicit type can cause unintentional form submissions in React context, and missing aria-labels hinder screen reader accessibility for vital actions.
**Action:** Always verify that every `<button>` element containing only icon/svg elements explicitly declares `type="button"` (if not a submit button) and provides a clear `aria-label` (or dynamically generated label derived from the `title` attribute if suitable). Consider using a script or regex to audit `<button><svg/></button>` instances continuously to maintain a11y standards.
