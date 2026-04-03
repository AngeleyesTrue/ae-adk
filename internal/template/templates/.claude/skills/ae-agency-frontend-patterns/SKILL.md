---
name: ae-agency-frontend-patterns
description: >
  Frontend development patterns for AI Agency projects covering tech stack
  preferences, component architecture, file structure, and coding conventions.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-04"
  tags: "agency, frontend, react, nextjs, components"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["frontend", "component", "react", "nextjs", "build"]
  agents: ["agency-builder"]
  phases: ["build"]
---

# Agency Frontend Patterns

Frontend development patterns for Agency creative production.

## Default Tech Stack

- Framework: Next.js (App Router)
- Styling: Tailwind CSS
- Components: Composition-based (no class inheritance)
- State: React hooks (useState, useEffect)
- Animation: CSS transitions + Framer Motion for complex

## Component Architecture

- Atomic design: atoms -> molecules -> organisms -> templates -> pages
- Each component: single responsibility
- Props over context for component configuration
- Composition over configuration

## File Structure

```
src/
  app/           # Next.js App Router pages
  components/
    ui/          # Atomic design components
    sections/    # Page sections (hero, features, etc.)
    layouts/     # Layout wrappers
  styles/
    tokens.css   # Design token CSS variables
    globals.css  # Global styles
```
