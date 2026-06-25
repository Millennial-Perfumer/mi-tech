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
## 2024-05-18 - Accessibility: Modal Close Buttons and Table Sort Headers
**Learning:** Icon-only buttons (like modal close buttons) and text buttons in table headers that trigger sorting actions often lack `type="button"`. This can cause unintended form submissions if they are placed inside a form. Icon-only buttons also require an `aria-label` attribute so that screen readers can convey their purpose correctly.
**Action:** Always add `type="button"` and `aria-label` to icon-only buttons (e.g. `aria-label="Close modal"` for a close button). Add `type="button"` to table header sort buttons as well.
## 2024-05-18 - Accessibility: Table Header Sorting
**Learning:** Adding a static `aria-label` to table header buttons that trigger sorting hides the dynamic sort direction indicator (`↑` or `↓`) from screen readers.
**Action:** Use the `aria-sort` attribute on the `<th>` element instead of a static `aria-label` on the button to properly convey the sort state to screen readers.
