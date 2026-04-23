## 2025-05-15 - Consistent Search 'Clear' Pattern
**Learning:** This app's design system favors using absolute-positioned 'Clear' buttons inside search inputs with manual hover state management via `onMouseEnter`/`onMouseLeave` inline styles. This ensures micro-UX consistency across different modules like Customers and Automation.
**Action:** Always include a 'Clear' button for search inputs that resets both the query and pagination, using the established inline styling pattern for consistency.

## 2025-05-22 - Standard Modal UX Pattern
**Learning:** Premium modals (e.g., `SocialComposerModal.tsx`) should feature a standard top-right close button using an SVG icon, `aria-label="Close modal"`, and `title="Close"`. They should also support closing by clicking the modal overlay. Hover effects (e.g., background: `var(--bg-hover)`) must be implemented via CSS classes in a component-level <style> block rather than inline `onMouseEnter`/`onMouseLeave` handlers to align with maintainability standards.
**Action:** Implement modal close buttons and overlay-click-to-close as a standard practice. Use kebab-case CSS properties in `<style>` blocks.

## 2025-05-22 - Character Count Micro-UX
**Learning:** For textareas where brevity is expected (e.g., social posts), including a real-time character count display below the field improves the user experience. To ensure accessibility and polish, implement `aria-live="polite"` and proper pluralization (e.g., '1 character' vs '2 characters').
**Action:** Add character counters to constrained text inputs with proper ARIA attributes.
