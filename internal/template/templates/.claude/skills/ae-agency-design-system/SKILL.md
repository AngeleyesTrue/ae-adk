---
name: ae-agency-design-system
description: >
  Visual design system patterns for AI Agency projects covering color palettes,
  typography, spacing, layout, and component design tokens. Enforces brand visual
  consistency and eliminates AI-generated design anti-patterns.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-04"
  tags: "agency, design-system, tokens, typography, colors"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["design system", "design tokens", "color palette", "typography"]
  agents: ["agency-designer"]
  phases: ["build"]
---

# Agency Design System

Visual design system patterns for Agency creative production.

## Design Token Categories

### Colors
- Primary: Brand primary color + 5 shades (50-900)
- Neutral: Gray scale for text, borders, backgrounds
- Semantic: Success, Warning, Error, Info
- Surface: Background layers (1-4 depth levels)

### Typography
- Display: Hero headlines (clamp-based responsive)
- Heading: Section headers (h1-h6)
- Body: Paragraph text (base 16px, line-height 1.5)
- Caption: Small text, labels

### Spacing
- Base unit: 4px
- Scale: 0, 1, 2, 3, 4, 5, 6, 8, 10, 12, 16, 20, 24
- Section padding: 64px-128px vertical

## Anti-Patterns

- Random colors not from the palette
- Inconsistent spacing between similar elements
- Missing hover/focus states
- Non-responsive typography
