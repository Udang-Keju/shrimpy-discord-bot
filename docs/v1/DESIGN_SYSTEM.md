# Design System
## Project: **Shrimpy** 🦐 — Dashboard Visual Identity & Color System

> **Version**: 1.0.0-draft
> **Status**: In Review
> **Last Updated**: 2026-06-21
> **Applies To**: Next.js Web Dashboard (`dashboard/`)

---

## Table of Contents

1. [Brand Identity](#1-brand-identity)
2. [Color Palette — Dark Theme](#2-color-palette--dark-theme)
3. [Color Palette — Light Theme](#3-color-palette--light-theme)
4. [Semantic Color Tokens](#4-semantic-color-tokens)
5. [Typography](#5-typography)
6. [Spacing & Layout](#6-spacing--layout)
7. [Border Radius & Shadows](#7-border-radius--shadows)
8. [Component Tokens](#8-component-tokens)
9. [CSS Custom Properties](#9-css-custom-properties)
10. [Theme Switching](#10-theme-switching)

---

## 1. Brand Identity

**Shrimpy** draws its visual personality from the ocean — the warm coral tones of a Shrimpy's shell, the deep navy of the sea at night, and the bright teal of shallow tropical waters.

### Core Brand Values

| Value | Manifestation |
|-------|---------------|
| **Playful** | Rounded corners, soft shadows, emoji-friendly UI |
| **Trustworthy** | Consistent spacing, clear hierarchy, accessible contrast |
| **Ocean-inspired** | Coral pinks, ocean teals, deep navy, sandy neutrals |
| **Modern** | Clean typography, glassmorphism accents, smooth transitions |

### Brand Colors (Raw)

These are the raw hue anchors from which all theme tokens are derived:

| Name | Hex | HSL | Description |
|------|-----|-----|-------------|
| **Shrimpy Coral** | `#FF7B6B` | `hsl(6, 100%, 71%)` | Primary brand color — warm coral pink |
| **Deep Coral** | `#E8503A` | `hsl(9, 78%, 57%)` | Darker coral for light-mode primary |
| **Ocean Teal** | `#4ECDC4` | `hsl(177, 52%, 55%)` | Accent — bright tropical teal |
| **Deep Teal** | `#2A9D8F` | `hsl(174, 57%, 39%)` | Darker teal for light-mode accent |
| **Sandy Cream** | `#FFF8F2` | `hsl(27, 100%, 97%)` | Lightest background for light theme |
| **Deep Navy** | `#1A1830` | `hsl(244, 32%, 14%)` | Darkest background for dark theme |
| **Warm Sand** | `#F5DEB3` | `hsl(39, 77%, 83%)` | Tertiary warm tone |
| **Pearl** | `#F5F0FF` | `hsl(260, 100%, 97%)` | Light text on dark backgrounds |

---

## 2. Color Palette — Dark Theme

Inspired by a deep ocean at night — rich navy backgrounds with glowing coral and teal accents.

```
Dark Theme Preview:
┌─────────────────────────────────────────────────────────┐
│ Background    #1A1830  ████████████████████████████████ │
│ Surface       #242140  ████████████████████████████████ │
│ Surface+      #2E2B52  ████████████████████████████████ │
│ Border        #3D3960  ████████████████████████████████ │
│ Primary       #FF7B6B  ████████████████████████████████ │
│ Primary Hover #FF9485  ████████████████████████████████ │
│ Accent        #4ECDC4  ████████████████████████████████ │
│ Text Primary  #F5F0FF  ████████████████████████████████ │
│ Text Muted    #9B93C0  ████████████████████████████████ │
└─────────────────────────────────────────────────────────┘
```

| Token | Hex | HSL | Usage |
|-------|-----|-----|-------|
| `--bg-base` | `#1A1830` | `hsl(244, 32%, 14%)` | Page background |
| `--bg-surface` | `#242140` | `hsl(244, 31%, 19%)` | Cards, panels, sidebars |
| `--bg-surface-elevated` | `#2E2B52` | `hsl(244, 30%, 25%)` | Modals, dropdowns, popovers |
| `--bg-surface-hover` | `#383562` | `hsl(244, 30%, 30%)` | Hover state for surface elements |
| `--border-subtle` | `#3D3960` | `hsl(244, 24%, 30%)` | Dividers, card borders |
| `--border-default` | `#5451A0` | `hsl(244, 30%, 48%)` | Input borders, focus rings base |
| `--primary` | `#FF7B6B` | `hsl(6, 100%, 71%)` | Buttons, links, active states |
| `--primary-hover` | `#FF9485` | `hsl(6, 100%, 76%)` | Hover state for primary |
| `--primary-muted` | `#FF7B6B26` | `hsl(6, 100%, 71%, 15%)` | Primary tint backgrounds |
| `--accent` | `#4ECDC4` | `hsl(177, 52%, 55%)` | Highlights, badges, charts |
| `--accent-hover` | `#6ED9D2` | `hsl(177, 52%, 64%)` | Hover state for accent |
| `--accent-muted` | `#4ECDC426` | `hsl(177, 52%, 55%, 15%)` | Accent tint backgrounds |
| `--success` | `#52D99B` | `hsl(150, 62%, 59%)` | Success states |
| `--warning` | `#FFB347` | `hsl(35, 100%, 64%)` | Warning states |
| `--danger` | `#FF5757` | `hsl(0, 100%, 67%)` | Error, destructive actions |
| `--info` | `#5BA4FF` | `hsl(214, 100%, 68%)` | Informational states |
| `--text-primary` | `#F5F0FF` | `hsl(260, 100%, 97%)` | Body text, headings |
| `--text-secondary` | `#BDB8E0` | `hsl(244, 34%, 80%)` | Secondary labels |
| `--text-muted` | `#9B93C0` | `hsl(244, 26%, 66%)` | Placeholder text, hints |
| `--text-disabled` | `#5C5880` | `hsl(244, 18%, 42%)` | Disabled elements |
| `--text-on-primary` | `#FFFFFF` | `hsl(0, 0%, 100%)` | Text on coral buttons |

---

## 3. Color Palette — Light Theme

Inspired by a warm tropical beach — sandy whites, bright coral, and clear teal waters.

```
Light Theme Preview:
┌─────────────────────────────────────────────────────────┐
│ Background    #FFF8F2  ████████████████████████████████ │
│ Surface       #FFFFFF  ████████████████████████████████ │
│ Surface+      #FFF0E8  ████████████████████████████████ │
│ Border        #F0D5C8  ████████████████████████████████ │
│ Primary       #E8503A  ████████████████████████████████ │
│ Primary Hover #C93F2B  ████████████████████████████████ │
│ Accent        #2A9D8F  ████████████████████████████████ │
│ Text Primary  #1A0F1F  ████████████████████████████████ │
│ Text Muted    #7A5C6E  ████████████████████████████████ │
└─────────────────────────────────────────────────────────┘
```

| Token | Hex | HSL | Usage |
|-------|-----|-----|-------|
| `--bg-base` | `#FFF8F2` | `hsl(27, 100%, 97%)` | Page background |
| `--bg-surface` | `#FFFFFF` | `hsl(0, 0%, 100%)` | Cards, panels |
| `--bg-surface-elevated` | `#FFF0E8` | `hsl(23, 100%, 95%)` | Modals, popovers |
| `--bg-surface-hover` | `#FFE8DC` | `hsl(18, 100%, 93%)` | Hover state for surface elements |
| `--border-subtle` | `#F0D5C8` | `hsl(18, 62%, 86%)` | Dividers, card borders |
| `--border-default` | `#D4A898` | `hsl(15, 42%, 72%)` | Input borders |
| `--primary` | `#E8503A` | `hsl(9, 78%, 57%)` | Buttons, links, active states |
| `--primary-hover` | `#C93F2B` | `hsl(9, 64%, 48%)` | Hover state for primary |
| `--primary-muted` | `#E8503A1A` | `hsl(9, 78%, 57%, 10%)` | Primary tint backgrounds |
| `--accent` | `#2A9D8F` | `hsl(174, 57%, 39%)` | Highlights, badges, charts |
| `--accent-hover` | `#1E7A6F` | `hsl(174, 59%, 30%)` | Hover state for accent |
| `--accent-muted` | `#2A9D8F1A` | `hsl(174, 57%, 39%, 10%)` | Accent tint backgrounds |
| `--success` | `#27AE60` | `hsl(145, 63%, 42%)` | Success states |
| `--warning` | `#E67E22` | `hsl(28, 80%, 52%)` | Warning states |
| `--danger` | `#E74C3C` | `hsl(6, 78%, 57%)` | Error, destructive actions |
| `--info` | `#2980B9` | `hsl(204, 64%, 44%)` | Informational states |
| `--text-primary` | `#1A0F1F` | `hsl(280, 28%, 9%)` | Body text, headings |
| `--text-secondary` | `#4A3558` | `hsl(278, 22%, 28%)` | Secondary labels |
| `--text-muted` | `#7A5C6E` | `hsl(330, 14%, 42%)` | Placeholder text, hints |
| `--text-disabled` | `#B09AAA` | `hsl(330, 12%, 65%)` | Disabled elements |
| `--text-on-primary` | `#FFFFFF` | `hsl(0, 0%, 100%)` | Text on coral buttons |

---

## 4. Semantic Color Tokens

These tokens map to UI meanings regardless of theme. Always use semantic tokens in components, never raw hex values.

| Semantic Token | Dark Value | Light Value | Usage |
|----------------|------------|-------------|-------|
| `--color-background` | `#1A1830` | `#FFF8F2` | Page root background |
| `--color-surface` | `#242140` | `#FFFFFF` | Card/panel backgrounds |
| `--color-surface-raised` | `#2E2B52` | `#FFF0E8` | Elevated overlays |
| `--color-border` | `#3D3960` | `#F0D5C8` | Default border color |
| `--color-primary` | `#FF7B6B` | `#E8503A` | Brand primary (coral) |
| `--color-primary-fg` | `#FFFFFF` | `#FFFFFF` | Text on primary elements |
| `--color-accent` | `#4ECDC4` | `#2A9D8F` | Brand accent (teal) |
| `--color-text` | `#F5F0FF` | `#1A0F1F` | Default body text |
| `--color-text-muted` | `#9B93C0` | `#7A5C6E` | Secondary/muted text |
| `--color-success` | `#52D99B` | `#27AE60` | Positive/success |
| `--color-warning` | `#FFB347` | `#E67E22` | Caution/warning |
| `--color-danger` | `#FF5757` | `#E74C3C` | Error/destructive |

---

## 5. Typography

### Font Stack

| Role | Font | Fallback | Import |
|------|------|----------|--------|
| **Display / Headings** | [Outfit](https://fonts.google.com/specimen/Outfit) | system-ui, sans-serif | Google Fonts |
| **Body / UI** | [Inter](https://fonts.google.com/specimen/Inter) | system-ui, sans-serif | Google Fonts |
| **Monospace / Code** | [JetBrains Mono](https://fonts.google.com/specimen/JetBrains+Mono) | 'Courier New', monospace | Google Fonts |

### Type Scale

| Token | Size | Weight | Line Height | Usage |
|-------|------|--------|-------------|-------|
| `--text-xs` | `0.75rem` (12px) | 400 | 1.5 | Captions, badges |
| `--text-sm` | `0.875rem` (14px) | 400 | 1.5 | Secondary labels, table cells |
| `--text-base` | `1rem` (16px) | 400 | 1.6 | Body text |
| `--text-lg` | `1.125rem` (18px) | 500 | 1.5 | Emphasized body |
| `--text-xl` | `1.25rem` (20px) | 600 | 1.4 | Card titles |
| `--text-2xl` | `1.5rem` (24px) | 700 | 1.3 | Section headings |
| `--text-3xl` | `1.875rem` (30px) | 700 | 1.2 | Page titles |
| `--text-4xl` | `2.25rem` (36px) | 800 | 1.1 | Hero headlines |

---

## 6. Spacing & Layout

### Spacing Scale (8px base grid)

| Token | Value | Usage |
|-------|-------|-------|
| `--space-1` | `4px` | Micro gaps (icon padding) |
| `--space-2` | `8px` | Compact padding |
| `--space-3` | `12px` | Small padding |
| `--space-4` | `16px` | Default padding |
| `--space-5` | `20px` | Medium spacing |
| `--space-6` | `24px` | Section padding |
| `--space-8` | `32px` | Large padding |
| `--space-10` | `40px` | Extra large |
| `--space-12` | `48px` | Section gaps |
| `--space-16` | `64px` | Major section breaks |

### Layout

| Element | Value |
|---------|-------|
| **Sidebar width** | `260px` |
| **Content max width** | `1200px` |
| **Navbar height** | `64px` |
| **Card padding** | `24px` |
| **Mobile breakpoint** | `768px` |

---

## 7. Border Radius & Shadows

### Border Radius

| Token | Value | Usage |
|-------|-------|-------|
| `--radius-sm` | `6px` | Badges, tags, small inputs |
| `--radius-md` | `10px` | Buttons, inputs |
| `--radius-lg` | `16px` | Cards, panels |
| `--radius-xl` | `24px` | Modals |
| `--radius-full` | `9999px` | Pills, avatars |

### Shadows — Dark Theme

| Token | Value | Usage |
|-------|-------|-------|
| `--shadow-sm` | `0 1px 3px rgba(0,0,0,0.4)` | Subtle lift |
| `--shadow-md` | `0 4px 16px rgba(0,0,0,0.5)` | Cards |
| `--shadow-lg` | `0 8px 32px rgba(0,0,0,0.6)` | Modals |
| `--shadow-primary` | `0 4px 20px rgba(255,123,107,0.3)` | Coral glow on hover |
| `--shadow-accent` | `0 4px 20px rgba(78,205,196,0.25)` | Teal glow |

### Shadows — Light Theme

| Token | Value | Usage |
|-------|-------|-------|
| `--shadow-sm` | `0 1px 3px rgba(100,60,40,0.08)` | Subtle lift |
| `--shadow-md` | `0 4px 16px rgba(100,60,40,0.12)` | Cards |
| `--shadow-lg` | `0 8px 32px rgba(100,60,40,0.15)` | Modals |
| `--shadow-primary` | `0 4px 20px rgba(232,80,58,0.25)` | Coral glow |
| `--shadow-accent` | `0 4px 20px rgba(42,157,143,0.2)` | Teal glow |

---

## 8. Component Tokens

### Buttons

| State | Background | Text | Border |
|-------|-----------|------|--------|
| **Primary default** | `--color-primary` | `#FFFFFF` | none |
| **Primary hover** | `--primary-hover` | `#FFFFFF` | none |
| **Primary active** | Darken 10% | `#FFFFFF` | none |
| **Secondary default** | `--bg-surface` | `--color-text` | `--color-border` |
| **Secondary hover** | `--bg-surface-hover` | `--color-text` | `--color-primary` |
| **Danger default** | `--color-danger` | `#FFFFFF` | none |
| **Ghost** | transparent | `--color-text-muted` | none |
| **Ghost hover** | `--bg-surface` | `--color-text` | none |

### Status Badge Colors

| Status | Dark bg | Dark text | Light bg | Light text |
|--------|---------|-----------|----------|------------|
| **Open** | `#52D99B26` | `#52D99B` | `#27AE601A` | `#27AE60` |
| **Claimed** | `#4ECDC426` | `#4ECDC4` | `#2A9D8F1A` | `#2A9D8F` |
| **Closed** | `#9B93C026` | `#9B93C0` | `#B09AAA1A` | `#7A5C6E` |
| **Archived** | `#5C588026` | `#5C5880` | `#D4A8981A` | `#7A5C6E` |

### Priority Badge Colors

| Priority | Dark bg | Dark text | Light bg | Light text |
|----------|---------|-----------|----------|------------|
| **Low** | `#52D99B26` | `#52D99B` | `#27AE601A` | `#27AE60` |
| **Medium** | `#4ECDC426` | `#4ECDC4` | `#2A9D8F1A` | `#2A9D8F` |
| **High** | `#FFB34726` | `#FFB347` | `#E67E221A` | `#E67E22` |
| **Urgent** | `#FF575726` | `#FF5757` | `#E74C3C1A` | `#E74C3C` |

---

## 9. CSS Custom Properties

Full CSS variable sheet to be placed in `dashboard/app/globals.css`:

```css
/* ============================================
   Shrimpy DESIGN SYSTEM — globals.css
   ============================================ */

@import url('https://fonts.googleapis.com/css2?family=Outfit:wght@400;500;600;700;800&family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap');

/* ─────────────────────────────────────────────
   DARK THEME (default)
   ───────────────────────────────────────────── */
:root,
[data-theme="dark"] {
  /* Backgrounds */
  --bg-base:             #1A1830;
  --bg-surface:          #242140;
  --bg-surface-elevated: #2E2B52;
  --bg-surface-hover:    #383562;

  /* Borders */
  --border-subtle:  #3D3960;
  --border-default: #5451A0;
  --border-focus:   #FF7B6B;

  /* Brand */
  --primary:         #FF7B6B;
  --primary-hover:   #FF9485;
  --primary-muted:   rgba(255, 123, 107, 0.15);
  --primary-fg:      #FFFFFF;
  --accent:          #4ECDC4;
  --accent-hover:    #6ED9D2;
  --accent-muted:    rgba(78, 205, 196, 0.15);

  /* Status */
  --success: #52D99B;
  --warning: #FFB347;
  --danger:  #FF5757;
  --info:    #5BA4FF;

  /* Text */
  --text-primary:   #F5F0FF;
  --text-secondary: #BDB8E0;
  --text-muted:     #9B93C0;
  --text-disabled:  #5C5880;
  --text-on-primary: #FFFFFF;

  /* Shadows */
  --shadow-sm:      0 1px 3px rgba(0, 0, 0, 0.4);
  --shadow-md:      0 4px 16px rgba(0, 0, 0, 0.5);
  --shadow-lg:      0 8px 32px rgba(0, 0, 0, 0.6);
  --shadow-primary: 0 4px 20px rgba(255, 123, 107, 0.3);
  --shadow-accent:  0 4px 20px rgba(78, 205, 196, 0.25);

  /* Semantic aliases */
  --color-background:   var(--bg-base);
  --color-surface:      var(--bg-surface);
  --color-surface-raised: var(--bg-surface-elevated);
  --color-border:       var(--border-subtle);
  --color-primary:      var(--primary);
  --color-primary-fg:   var(--primary-fg);
  --color-accent:       var(--accent);
  --color-text:         var(--text-primary);
  --color-text-muted:   var(--text-muted);
  --color-success:      var(--success);
  --color-warning:      var(--warning);
  --color-danger:       var(--danger);
}

/* ─────────────────────────────────────────────
   LIGHT THEME
   ───────────────────────────────────────────── */
[data-theme="light"] {
  /* Backgrounds */
  --bg-base:             #FFF8F2;
  --bg-surface:          #FFFFFF;
  --bg-surface-elevated: #FFF0E8;
  --bg-surface-hover:    #FFE8DC;

  /* Borders */
  --border-subtle:  #F0D5C8;
  --border-default: #D4A898;
  --border-focus:   #E8503A;

  /* Brand */
  --primary:         #E8503A;
  --primary-hover:   #C93F2B;
  --primary-muted:   rgba(232, 80, 58, 0.1);
  --primary-fg:      #FFFFFF;
  --accent:          #2A9D8F;
  --accent-hover:    #1E7A6F;
  --accent-muted:    rgba(42, 157, 143, 0.1);

  /* Status */
  --success: #27AE60;
  --warning: #E67E22;
  --danger:  #E74C3C;
  --info:    #2980B9;

  /* Text */
  --text-primary:   #1A0F1F;
  --text-secondary: #4A3558;
  --text-muted:     #7A5C6E;
  --text-disabled:  #B09AAA;
  --text-on-primary: #FFFFFF;

  /* Shadows */
  --shadow-sm:      0 1px 3px rgba(100, 60, 40, 0.08);
  --shadow-md:      0 4px 16px rgba(100, 60, 40, 0.12);
  --shadow-lg:      0 8px 32px rgba(100, 60, 40, 0.15);
  --shadow-primary: 0 4px 20px rgba(232, 80, 58, 0.25);
  --shadow-accent:  0 4px 20px rgba(42, 157, 143, 0.2);

  /* Semantic aliases (inherit structure from dark) */
  --color-background:   var(--bg-base);
  --color-surface:      var(--bg-surface);
  --color-surface-raised: var(--bg-surface-elevated);
  --color-border:       var(--border-subtle);
  --color-primary:      var(--primary);
  --color-primary-fg:   var(--primary-fg);
  --color-accent:       var(--accent);
  --color-text:         var(--text-primary);
  --color-text-muted:   var(--text-muted);
  --color-success:      var(--success);
  --color-warning:      var(--warning);
  --color-danger:       var(--danger);
}

/* ─────────────────────────────────────────────
   TYPOGRAPHY
   ───────────────────────────────────────────── */
:root {
  --font-display: 'Outfit', system-ui, sans-serif;
  --font-body:    'Inter', system-ui, sans-serif;
  --font-mono:    'JetBrains Mono', 'Courier New', monospace;

  --text-xs:   0.75rem;
  --text-sm:   0.875rem;
  --text-base: 1rem;
  --text-lg:   1.125rem;
  --text-xl:   1.25rem;
  --text-2xl:  1.5rem;
  --text-3xl:  1.875rem;
  --text-4xl:  2.25rem;
}

/* ─────────────────────────────────────────────
   SPACING
   ───────────────────────────────────────────── */
:root {
  --space-1:  4px;
  --space-2:  8px;
  --space-3:  12px;
  --space-4:  16px;
  --space-5:  20px;
  --space-6:  24px;
  --space-8:  32px;
  --space-10: 40px;
  --space-12: 48px;
  --space-16: 64px;
}

/* ─────────────────────────────────────────────
   BORDER RADIUS
   ───────────────────────────────────────────── */
:root {
  --radius-sm:   6px;
  --radius-md:   10px;
  --radius-lg:   16px;
  --radius-xl:   24px;
  --radius-full: 9999px;
}

/* ─────────────────────────────────────────────
   BASE STYLES
   ───────────────────────────────────────────── */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html {
  font-family: var(--font-body);
  background-color: var(--bg-base);
  color: var(--text-primary);
  font-size: 16px;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-display);
  color: var(--text-primary);
}

a {
  color: var(--primary);
  text-decoration: none;
  transition: color 0.15s ease;
}

a:hover {
  color: var(--primary-hover);
}

code, pre, kbd {
  font-family: var(--font-mono);
}

/* ─────────────────────────────────────────────
   TRANSITIONS
   ───────────────────────────────────────────── */
:root {
  --transition-fast:   100ms ease;
  --transition-base:   200ms ease;
  --transition-slow:   350ms ease;
  --transition-theme:  300ms ease; /* for theme switching */
}
```

---

## 10. Theme Switching

Theme is toggled via a `data-theme` attribute on the `<html>` element. The default theme is **dark**.

### Implementation in Next.js

```typescript
// dashboard/lib/theme.ts

export type Theme = 'dark' | 'light';

export function getSystemTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  return window.matchMedia('(prefers-color-scheme: dark)').matches
    ? 'dark'
    : 'light';
}

export function applyTheme(theme: Theme) {
  document.documentElement.setAttribute('data-theme', theme);
  localStorage.setItem('Shrimpy-theme', theme);
}

export function getSavedTheme(): Theme {
  if (typeof window === 'undefined') return 'dark';
  const saved = localStorage.getItem('Shrimpy-theme') as Theme | null;
  return saved ?? getSystemTheme();
}
```

### Storing Theme Preference

- Theme preference is stored in `localStorage` under the key `Shrimpy-theme`.
- On initial load, the system default is respected if no preference is saved.
- The theme toggle button is rendered in the top navigation bar.

### Preventing Flash of Unstyled Content (FOUC)

Add this inline script to `dashboard/app/layout.tsx` **before any CSS loads**:

```tsx
// In <head>, before any stylesheets:
<script
  dangerouslySetInnerHTML={{
    __html: `
      (function() {
        var theme = localStorage.getItem('Shrimpy-theme') ||
          (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
        document.documentElement.setAttribute('data-theme', theme);
      })();
    `,
  }}
/>
```

---

*End of Design System — Shrimpy v1.0.0-draft*
