## 2025-05-15 - Consistent Search 'Clear' Pattern
**Learning:** This app's design system favors using absolute-positioned 'Clear' buttons inside search inputs with manual hover state management via `onMouseEnter`/`onMouseLeave` inline styles. This ensures micro-UX consistency across different modules like Customers and Automation.
**Action:** Always include a 'Clear' button for search inputs that resets both the query and pagination, using the established inline styling pattern for consistency.
