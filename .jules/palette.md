## 2026-05-22 - [Secure Authenticated Downloads]
**Learning:** Using `window.open` for authenticated file downloads (like invoices) bypasses the application's auth header logic unless tokens are passed in the URL (insecure). Fetching as a Blob via `fetchWithAuth` ensures security and allows for UI loading feedback.
**Action:** Always prefer `fetchWithAuth` + `URL.createObjectURL` for secure, async-feedback file downloads.

## 2026-05-22 - [Icon-only Button Accessibility]
**Learning:** Many icon-only buttons in the existing codebase missed `type="button"`, causing accidental form submissions, and lacked `aria-label` for screen readers.
**Action:** Ensure all icon-only buttons have explicit `type="button"` and descriptive `aria-label`.

## 2026-05-24 - [Tab-Aware Search Shortcut]
**Learning:** In applications with many tabs that share similar search components, a global search shortcut must dynamically identify the *visible* input to avoid focusing a hidden element from a background tab.
**Action:** Use `document.querySelectorAll` with visibility checks (`getBoundingClientRect().width > 0`) to target the active search input.
