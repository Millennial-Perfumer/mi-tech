## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.

## 2026-05-30 - [Global Modal Accessibility & Keyboard Support]
**Learning:** Global context-provided modals (like `ConfirmContext`) often overlook keyboard navigation and ARIA attributes. Implementing an `Escape` key listener and linking the container to its content via `aria-labelledby` and `aria-describedby` provides a significant accessibility boost with minimal code changes.
**Action:** Always include keyboard dismissal and semantic ARIA linking for global modal components.
