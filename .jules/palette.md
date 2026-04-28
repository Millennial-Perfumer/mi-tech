## 2025-05-15 - Consistent Search 'Clear' Pattern
**Learning:** This app's design system favors using absolute-positioned 'Clear' buttons inside search inputs with manual hover state management via `onMouseEnter`/`onMouseLeave` inline styles. This ensures micro-UX consistency across different modules like Customers and Automation.
**Action:** Always include a 'Clear' button for search inputs that resets both the query and pagination, using the established inline styling pattern for consistency.
## 2026-04-28 - Standard Modal UX Pattern
**Learning:** Premium modals should include a top-right close button, real-time character count for text inputs (with `aria-live="polite"`), and support closing via overlay click. To satisfy 'no custom CSS' guardrails, use inline styles for layout adjustments and reuse existing `useToast` hooks instead of native alerts. Hover states for icon buttons can be managed via `onMouseEnter`/`onMouseLeave` to maintain consistency with the search input pattern.
**Action:** Use this composite pattern for all new or enhanced modal components to ensure accessibility and professional feel without adding CSS bloat.
