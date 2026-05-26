## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.

## 2026-05-26 - [Standardized Non-blocking Notifications]
**Learning:** Replacing native browser `alert()` calls with a themed toast system (via `useToast`) provides a more professional, integrated, and non-blocking user experience. Native alerts interrupt the user flow and cannot be styled to match the app's design system.
**Action:** Use `toastSuccess`, `toastError`, etc., from the `useToast` hook for all user notifications and feedback.
