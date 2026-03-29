## 2025-05-15 - [Accessible Dropdown & Global Focus States]
**Learning:** For apps with a custom design system that disables default focus outlines, a global `:focus-visible` style is a crucial accessibility win that doesn't compromise visual design for mouse users. Component-level accessibility for dropdowns requires both ARIA roles/attributes and keyboard event listeners (like Escape) to be truly inclusive.
**Action:** Always check `index.css` for `outline: none` and ensure `:focus-visible` is implemented. When building popovers/dropdowns, use the `useEffect` hook to manage global keyboard listeners for the `Escape` key.

## 2025-05-16 - [Search 'Clear' Button Pattern]
**Learning:** Search 'Clear' buttons must reset both the search query and the pagination state (e.g., `setPage(1)`) to avoid displaying empty results on high page numbers after the filter is cleared.
**Action:** Always include `setPage(1)` (or equivalent pagination reset) in clear button click handlers. Explicitly set `border: 'none'` and `background: 'transparent'` for such buttons to ensure cross-browser visual consistency.

## 2025-05-17 - [Login & OTP Form Accessibility]
**Learning:** Standardizing label-input associations via `htmlFor` and `id` is a fundamental accessibility requirement that is often overlooked in custom UI components. For authentication flows, adding `inputMode="numeric"` and `autoComplete="one-time-code"` significantly improves the mobile UX by triggering the correct keyboard and allowing the OS to autofill verification codes from SMS/WhatsApp.
**Action:** Always verify that every form input has a corresponding label with matching `htmlFor`/`id`. For OTP/verification fields, always include `inputMode="numeric"` and `autoComplete="one-time-code"`.
