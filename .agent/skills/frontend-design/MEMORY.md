# MI-Tech Frontend Design Memory

## Design System
- **Font**: Montserrat (Google Fonts), weights 400-700
- **Theme**: Dark/Light via `[data-theme="dark"]` / `[data-theme="light"]` on `<html>`
- **Aesthetic**: Glassmorphism — `backdrop-filter: blur()`, translucent backgrounds, subtle borders
- **Colors**: CSS custom properties (design tokens) defined in `App.css`
- **Icons**: Inline SVGs (no icon library)
- **Animations**: CSS transitions on hover, subtle transforms
- **No component library** — all custom

## Design Tokens (CSS Variables)
```
--bg-primary, --bg-secondary, --bg-card, --bg-input
--text-primary, --text-secondary, --text-tertiary
--border-color, --border-color-light
--accent-primary, --accent-success, --accent-danger, --accent-warning
--gradient-primary
```

## Component Patterns
- **Cards**: `.card` class with glassmorphism effect
- **Buttons**: `.btn-primary`, `.btn-secondary` with hover transitions
- **Modals**: Centered overlay with backdrop blur
- **Tables**: Alternating row colors, sticky headers
- **Status badges**: Color-coded pills (paid=green, pending=amber, cancelled=red)
- **Settings**: Categorized cards with collapsible sections, secret masking

## Layout
- **Sidebar**: Collapsible navigation with icons + labels
- **Main content**: Scrollable area with sticky header
- **Mobile**: Responsive via media queries, `isMobile` state detection
