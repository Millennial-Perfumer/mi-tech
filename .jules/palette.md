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

## 2024-06-25 - Accessible Table Sorting Headers
**Learning:** For table header columns that trigger sorting and include dynamic visual indicators (like ↑ or ↓), placing a static `aria-label` on the inner button is an anti-pattern as it overrides the text for screen readers. Instead, the correct pattern is to apply the `aria-sort` attribute (`"ascending"`, `"descending"`, or `"none"`) directly to the parent `<th>` element, and ensure the inner interactive trigger explicitly sets `type="button"`.
**Action:** Always apply `aria-sort` to the `<th>` when building sortable data tables, and ensure inner trigger buttons have `type="button"` to prevent accidental form submissions.
## 2024-05-24 - Add accessibility attributes to WhatsAppChat and PurchaseOrders buttons
**Learning:** Interactive UI components (like file upload, send, toggle buttons in chats, and icon-only edit/delete actions in tables) frequently omit explicit `type="button"` and ARIA attributes (like `aria-pressed` on toggles and `aria-label` on icon-only buttons) since they are often built outside strict HTML forms. This pattern risks accidental browser refreshes if they ever get wrapped in a form, lacks screen reader clarity, and needs addressing in conversational interfaces and complex UI tables.
**Action:** Aggressively add `type="button"`, descriptive `aria-label`s, `aria-hidden="true"` for SVGs, and `aria-pressed` states for custom toggles across all interactive micro-components, even when not inside a form.
