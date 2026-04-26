# MI-Tech Frontend Memory

## Stack
- **Framework**: Vite + React 18 + TypeScript
- **Styling**: Vanilla CSS with CSS custom properties (design tokens in `App.css`)
- **No Tailwind, no component library** (pure custom components)
- **Font**: Montserrat (Google Fonts)
- **Design language**: Glassmorphism, dark/light theme via `[data-theme]` attribute

## Architecture
- **Entry**: `frontend/src/App.tsx` (single-page app with tab-based navigation via `activeTab` state)
- **Tabs**: dashboard, shopify (Orders), reports, products, automation, communication, tickets, customers, marketing, social, planner, users, feedback, settings
- **Major components**: `SettingsTab.tsx`, `Customers.tsx`, `Products.tsx`, `GSTReports.tsx`, `Planner.tsx`, `WhatsAppChat.tsx`, `Feedback.tsx`, `OrderDetailsModal.tsx`, `MarketingDashboard.tsx`, `SocialDashboard.tsx`
- **API layer**: `api.ts` with `API_BASE` constant, `fetchWithAuth()` wrapper for JWT auth
- **State**: Local component state + `localStorage` persistence for dates, columns, theme

## Configuration UI
- `SettingsTab.tsx` renders categorized `app_configs` from the backend
- Categories: `business`, `shopify`, `amazon`, `meta_shared`, `marketing`, `social_media`, `whatsapp`, `feedback`, `system`
- Each category has an icon, color, and title in `CATEGORY_META`
- Secret fields are masked with a toggle reveal

## Key Patterns
- Admin-only features gated by `userRole === 'admin'` (parsed from JWT)
- Toast notifications via `useToast()` context
- Confirmation dialogs via `useConfirm()` context
- Date range persisted to backend via `/api/settings/date-range`
- Column selector with localStorage persistence

## Mobile Frontend
- Separate app at `frontend-mobile/` (Vite + React)
- Customer-facing, not admin
