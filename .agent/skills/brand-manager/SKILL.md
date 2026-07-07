---
name: brand-manager
description: Rules and guidelines to maintain premium UI branding, CSS variables, and cohesive animations across Vite + React apps.
---

# 🎨 Brand Manager — UI Branding & CSS Guidelines

Use this skill to guide and audit styling changes, ensuring the user interface remains premium, cohesive, clean, and interactive.

---

## 🌟 Visual Core Values

1. **Autonomy of Aesthetic**: Every UI surface must look deliberately designed, not generic. 
2. **Intentional Motion**: Use animations to orient the user, stage information, and reinforce actions. Avoid gratuitous spinners or bouncy elements that don't add functional value.
3. **CSS-First Tokens**: Never hardcode colors, font families, margins, or padding. Always map them to CSS custom properties defined in `index.css`.
4. **Adaptive Dark Mode**: All design elements must automatically adapt to dark mode (using the `[data-theme="dark"]` prefix) with carefully chosen low-contrast borders and elevated surfaces.

---

## 🛠️ CSS Token Architecture

Always utilize the variables defined in [index.css](file:///Users/siddiqs_office/Documents/Personal%20Dev/GST%20Invoice%20Manager/frontend/src/index.css):

### 🎨 Backgrounds & Borders
- **Main App Background**: `var(--bg-color)`
- **Surface/Card Container**: `var(--surface-color)` (light: `#ffffff`, dark: `#0d0d12`)
- **Default Border**: `var(--border-color)`
- **Interactive Hover bg**: `var(--bg-hover)`

### 💎 Glassmorphism & Cards
- **Premium Glass Card**: Use class `.glass-card-premium` for panels that overlay complex backgrounds.
- **Adapative Metric Card**: Use class `.metric-card-adaptive` for dashboard widgets.
- **Base Card**: Use class `.card` (includes hover lift + border transitions by default).

### ⚡ Accent Colors
- **Brand/Primary Accent**: `var(--accent-color)` (default is green `#10b981`)
- **Hover Accent**: `var(--accent-hover)`
- **Subtle Background Accent**: `var(--accent-subtle)` (perfect for tags or active button states)

---

## 🏗️ Predefined Component Classes in `index.css`

When building or updating UI elements, reuse these existing CSS styles instead of writing inline styles or custom selectors:

### 1. Buttons
- **Primary Action**: `.btn-primary` (solid accent background with subtle shadow, translates slightly on hover).
- **Secondary Action**: `.btn-secondary` (outlined surface color background, reacts to hover).
- **Toolbar Buttons**: `.toolbar-btn` (minimal icon-button wrapper).

### 2. Badge & Status Indicators
- **Standard Badges**: `.badge` combined with state subclasses:
  - Active/Green: `.badge-success`
  - Warning/Orange: `.badge-warning`
  - Danger/Red: `.badge-error`
- **Pill Badges (with dot)**: `.badge-pill` combined with color modifiers:
  - Success: `.badge-pill-success`
  - Warning: `.badge-pill-warning`
  - Danger: `.badge-pill-danger`
  - Gray: `.badge-pill-gray`
  - Accent/Info: `.badge-pill-info` or `.badge-pill-yellow`

### 3. Modals & Dialogs
- **Backdrop**: `.modal-overlay` (includes blur filter).
- **Body**: `.modal-content` (uses `--surface-color` and `--radius-xl` with scroll controls).
- **Headers/Footers**: `.modal-header`, `.modal-title`, `.modal-footer` (handles alignment & button gaps).

### 4. Layout & Grid Structures
- **Dashboard Grid**: `.dashboard-grid` (responsive CSS grid using standard gaps).
- **Table Container**: `.table-container` (auto-styled border and header sections).
- **Tables**: `table`, `th`, `td` (use Montserrat, padded cells, hover highlights on `tr`).

### 5. Form Elements
- **Form Groups**: `.form-group` (vertical layout spacing for label + input).
- **Form Rows**: `.form-row` (flex row layout that splits children inputs equally).
- **Inputs**: `input[type="..."]`, `select`, `textarea` (styled inputs with custom focus rings using `var(--accent-subtle)`).

### 6. Interactive Dropdowns & Upload Zones
- **Status Popovers**: `.status-popover` (elevated absolute-positioned status picker).
- **Upload Zones**: `.upload-zone` (dashed interactive border supporting hover states), `.upload-zone-icon`, `.upload-zone-text`.
- **File Previews**: `.file-preview-card`, `.file-info`, `.file-name`, `.file-status`.

---

## 🎬 Animation Guidelines

To achieve premium, clean movement:
1. **Entry Transitions**: Use `@keyframes fadeInUp` (adds translation + opacity fade) or `.nav-item-stagger` for loading animations.
2. **Interactive States**: Always define transitions for `hover` and `focus` states.
   - Use standard transitions: `transition: all var(--transition-base);` or `transition: transform var(--transition-base), box-shadow var(--transition-base);`.
3. **Micro-animations**:
   - Cards should hover up slightly: use `.hover-lift` class or `transform: translateY(-4px); box-shadow: var(--shadow-lg);`.
   - Sidebar item transition: `transition: width var(--transition-slow);`.

---

## ❌ Anti-Patterns (Avoid these!)
- **❌ Hardcoded Hexes**: Using `#ffffff` or `#1a1a1a` instead of `var(--surface-color)` or `var(--bg-color)`.
- **❌ Inconsistent Radii**: Using raw values like `border-radius: 5px` or `border-radius: 10px`. Use token classes or variables: `var(--radius-sm)` (8px), `var(--radius-md)` (12px), `var(--radius-lg)` (16px), or `var(--radius-xl)` (24px).
- **❌ Missing Focus States**: Creating interactive buttons/inputs without a `focus-visible` ring.
- **❌ Rainbow Palettes**: Using multiple primary colors. Keep a single accent color (`--accent-color`) and semantic badges for states (`--status-active`, `--status-warning`, `--status-danger`).

---

## 💡 Proposing Brand Updates
When suggesting UI upgrades:
- Recommend clean whitespace, strong alignment (e.g. Montserrat font stacks), and layered transparency instead of flat blocks of color.
- Group layouts logically using CSS Grid (`.dashboard-grid`) or Flexbox with standard gap spacing.
