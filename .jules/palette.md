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
## 2026-06-25 - [Admin Note Accessibility and Modal Buttons]
**Learning:** Found an admin note modal close button containing only an SVG without an `aria-label` or `type="button"`. Other "Cancel" buttons in modals also lacked `type="button"`. This pattern risks accidental form submissions and poor screen reader experiences.
**Action:** Always verify `type="button"` and `aria-label` when creating modal controls, particularly for icon-only close buttons.
