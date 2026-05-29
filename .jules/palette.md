## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.

## 2026-05-23 - [Asynchronous Feedback for Form Submission]
**Learning:** For asynchronous operations like product creation, missing loading feedback leads to user uncertainty and potential duplicate submissions. Implementing an `isSaving` state with button label changes (e.g., "Creating...") and `disabled` attributes improves both UX and data integrity.
**Action:** Implement `isSaving` logic and visual feedback for all async create/update/delete operations.
