## 2025-05-15 - Consistent Search 'Clear' Pattern
**Learning:** This app's design system favors using absolute-positioned 'Clear' buttons inside search inputs with manual hover state management via `onMouseEnter`/`onMouseLeave` inline styles. This ensures micro-UX consistency across different modules like Customers and Automation.
**Action:** Always include a 'Clear' button for search inputs that resets both the query and pagination, using the established inline styling pattern for consistency.

## 2026-05-01 - State-driven Hover Pattern for Interactive Elements
**Learning:** Manual hover states for interactive elements (like Search 'Clear' buttons) should be managed using React `useState` (e.g., `isHovered`) rather than direct DOM style manipulation via `e.currentTarget.style`. This ensures visual state consistency, especially when elements are conditionally rendered.
**Action:** Use `useState` and `onMouseEnter`/`onMouseLeave` to drive hover styles. Remember to reset the hover state in the `onClick` handler if the element might be unmounted.
