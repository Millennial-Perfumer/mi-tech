## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.

## 2026-05-22 - [Standardized Accessible Modals]
**Learning:** Global context-driven modals (like `ConfirmContext`) often lack keyboard dismissal, focus restoration, and semantic ARIA roles (`alertdialog`), making them hazardous for keyboard/screen reader users.
**Action:** Implement `Escape` key listeners, `useRef` for focus restoration, and `role="alertdialog"` in global confirmation components. For destructive actions, default focus to 'Cancel'.
