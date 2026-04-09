---
name: design:extract
description: >
  Extract design context from an existing website or codebase.
  Analyzes colors, typography, spacing, and patterns to build
  or update .impeccable.md design context.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "command"
  status: "active"
  updated: "2026-04-09"
  tags: "design, extract, context, analysis, reverse-engineer"
---

# design:extract

Reverse-engineer design context from existing code or a live website. Populate or update the project's .impeccable.md configuration.

## Usage

`/design:extract [source]` where source can be a URL, file path, or directory.

## Extraction Targets

### Colors

- Extract all color values from CSS, Tailwind config, and design token files
- Identify color roles: primary, neutral, semantic (success/warning/error/info), and surface
- Detect color space: OKLCH, HSL, RGB, or hex
- Flag pure black (#000), pure white (#fff), and pure gray usage as potential anti-patterns
- Map extracted colors to the 60-30-10 weight distribution model

### Typography

- Extract font families, including fallback stacks
- Identify the size scale and detect modular ratio if present
- Catalog font weights in use and detect missing weight files (faux bold indicators)
- Extract line-height values and correlate with container widths
- Detect if fonts are from reflex/distinctive foundries or generic selections
- Check font-display strategy and loading optimization

### Spacing

- Extract all spacing values from margins, paddings, and gaps
- Identify the underlying scale system: 4pt, 8pt, or custom
- Detect rhythm patterns: consistent versus varied section spacing
- Flag off-scale values that break the spacing system

### Layout

- Identify grid systems: CSS Grid, Flexbox patterns, or framework grid classes
- Extract breakpoint definitions and responsive behavior
- Catalog card usage patterns: sizes, repetition counts, and variation
- Detect container query usage versus media query patterns

### Motion

- Extract transition and animation property values
- Identify easing curves: custom cubic-bezier, named easings, or spring functions
- Detect bounce/elastic easing usage as potential anti-patterns
- Catalog animation durations and check against the 100/300/500 rule

## Output

The extraction produces:

1. **Design Context Summary**: A structured analysis suitable for directly populating .impeccable.md sections (brand personality, color palette, typography, spacing, layout preferences)
2. **Anti-Pattern Violations**: Any detected violations of impeccable design rules found in the existing code, with severity and location
3. **Write Offer**: After presenting the summary, offer to write or update the project's .impeccable.md with the extracted context

## Source-Specific Behavior

### For URLs

Use browser automation tools (chrome MCP) or WebFetch to analyze live pages. Extract computed styles, font loading, and rendered layout information. Capture both desktop and mobile viewport states when possible.

### For Code (File Paths and Directories)

Use Grep and Glob to scan CSS files, component files, Tailwind configuration, design token files, and theme definitions. Parse both inline styles and external stylesheets. Check for CSS custom properties (variables) as the primary token source.
