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

## 2026-06-22 - [Accessibility for Complex Modals and Interactive Panels]
**Learning:** Newly introduced, complex interactive components (like Planner task/sprint boards) frequently omit baseline accessibility attributes (like `type="button"` and `aria-label`) on utility controls (Dismiss, Close, Delete). These omissions pose risks for unintended form submissions and severe usability drops for screen reader users.
**Action:** When implementing or modifying complex panels and modals, systematically apply defensive `type="button"` attributes and ensure every icon-only or utility button possesses a descriptive `aria-label`.

## 2026-07-03 - [Table Sort Buttons Accessibility]
**Learning:** Found table headers in `Feedback.tsx` using `aria-label` dynamically on inner buttons to indicate sorting (e.g., 'Customer ↑'). This approach visually works but can override or confuse screen reader announcements.
**Action:** Always apply `aria-sort="ascending|descending|none"` to the parent `<th>` element instead of relying on text/aria-labels on the inner button, and guarantee the inner trigger button maintains `type="button"`.
