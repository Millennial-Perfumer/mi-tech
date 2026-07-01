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

## 2026-06-27 - [Table Header Sorting Accessibility]
**Learning:** Found table header columns (`<th>`) that allow sorting but only use visual indicators (like ↑ or ↓) for state. When an interactive trigger is used inside the `<th>`, screen readers miss the sort state if only visual characters are used.
**Action:** Always apply the `aria-sort` attribute (`"ascending"`, `"descending"`, or `"none"`) to the parent `<th>` element of sortable columns, and ensure any internal interactive button explicitly uses `type="button"` to avoid form submission side-effects and improve screen reader announcements.
