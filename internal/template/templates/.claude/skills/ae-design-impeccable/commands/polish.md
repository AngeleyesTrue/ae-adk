---
name: design:polish
description: >
  Fine-tune design details: spacing alignment, color refinement,
  typography consistency, and micro-adjustments for production quality.
  The final pass before shipping.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "command"
  status: "active"
  updated: "2026-04-09"
  tags: "design, polish, spacing, alignment, refinement"
---

# design:polish

Pre-ship design polish pass. Focus on refinement and micro-adjustments, not restructuring.

## Usage

`/design:polish [target]` where target is an optional focus area (file, component, or directory).

## Polish Areas

### Spacing Consistency

- Ensure all spacing values follow the 4pt base scale: 4, 8, 12, 16, 24, 32, 48, 64, 96
- Flag non-scale spacing values (e.g., padding: 15px, margin: 22px)
- Verify rhythm variation: not the same spacing everywhere. Sections should breathe differently based on content hierarchy.
- Check that related elements share consistent internal spacing while maintaining visual grouping through proximity differences.

### Typography Polish

- Check that the type scale follows a modular ratio (e.g., 1.25 Major Third, 1.333 Perfect Fourth)
- Verify line-height inversely scales with container width: narrower columns get tighter leading
- Ensure minimum 16px body text size for readability
- Check that web fonts use font-display: swap to prevent FOIT
- Verify OpenType feature usage where appropriate: tabular-nums on data tables, liga and kern enabled on body text
- Check that font-weight values match available font weights (no faux bold from missing weight files)

### Color Polish

- Verify colors are defined in OKLCH color space for perceptual uniformity
- Check that neutral colors are tinted (warm or cool bias) rather than pure gray
- Verify the 60-30-10 color weight distribution: 60% dominant, 30% secondary, 10% accent
- Check dark mode adjustments: surface lightness follows an inverted scale, text weight is reduced by one step, accent colors maintain adequate contrast

### Alignment and Optical Corrections

- Check text optical alignment: visually align text baselines, not bounding boxes
- Verify icon centering includes optical adjustments (icons with asymmetric weight need offset)
- Check interactive touch target sizes: minimum 44px on all tappable elements
- Verify that visual weight distribution feels balanced even when mathematical centering is not used

### Motion Polish

- Verify transition durations follow the 100/300/500 rule: micro-interactions at 100ms, standard transitions at 300ms, complex animations at 500ms
- Check easing curves: no generic `ease` or `linear` on UI transitions. Use purpose-built curves (ease-out for entrances, ease-in for exits).
- Verify prefers-reduced-motion media query handling: all non-essential motion has a reduced or disabled alternative
- Check stagger timing on sequenced animations: cap total stagger duration to prevent sluggish cascades

## Output

Organized list of micro-adjustments grouped by polish area. Each item includes the specific CSS property, current value, recommended value, and the rationale for the change.

## Scope Boundaries

This command handles small, targeted adjustments only. If major structural issues are discovered during the polish pass, recommend `/design:normalize` for anti-pattern corrections or `/design:audit` for a full design review instead of attempting large-scale changes.
