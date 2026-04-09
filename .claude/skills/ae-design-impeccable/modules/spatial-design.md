# Spatial Design

Spatial relationships, grids, hierarchy, and optical adjustments for pixel-perfect layouts.

---

## 4pt Spacing System

Use a 4pt base grid, not 8pt. 8pt is too coarse for fine-grained UI work.

Scale: 4, 8, 12, 16, 24, 32, 48, 64, 96px.

Name tokens semantically, never by raw value:

- CORRECT: `--space-xs`, `--space-sm`, `--space-md`, `--space-lg`, `--space-xl`
- WRONG: `--spacing-4`, `--spacing-8`, `--spacing-16`

Semantic names survive scale changes. If the base shifts from 4pt to 6pt, `--space-sm` still makes sense while `--spacing-4` becomes a lie.

---

## Gap Over Margins

Use `gap` for sibling spacing. It eliminates margin collapse entirely.

Margin collapse is the single most confusing CSS behavior for both humans and AI. Two adjacent margins do not stack -- they overlap. `gap` on a flex or grid parent sidesteps the problem completely.

Reserve `margin` only for:
- Spacing between unrelated sections (structural separation)
- Negative margins for optical alignment corrections

Never use `margin-bottom` on children to create uniform spacing between siblings. Use `gap` on the parent.

---

## Self-Adjusting Grid

The breakpoint-free responsive grid pattern:

```css
.grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--space-lg);
}
```

This single declaration handles 1-column mobile through 4-column desktop without a single media query.

For complex layouts, use named grid areas instead of column/row numbers. Named areas are self-documenting and survive template changes:

```css
.layout {
  grid-template-areas:
    "header header"
    "sidebar main"
    "footer footer";
}
```

---

## Visual Hierarchy (Squint Test)

The squint test: blur or squint at the page. You should immediately identify:
1. The most important element
2. The second most important element
3. Clear visual groupings

If everything looks equally important, the hierarchy has failed.

Build hierarchy through multiple dimensions simultaneously:
- **Size**: Primary headings at 3:1 ratio minimum vs body text
- **Weight**: Bold vs regular creates instant separation
- **Color contrast**: High contrast for primary, muted for secondary
- **Position**: Top-left (LTR) gets scanned first
- **Whitespace**: More space around important elements isolates them

Best practice: Use 2-3 dimensions at once. Size + weight for headings. Color + position for CTAs. Using only one dimension (e.g., size alone) creates weak hierarchy.

---

## Cards Are Not Required

Cards are the most overused UI pattern. They add visual noise, consume space, and create nesting nightmares.

Spacing and alignment alone create grouping. The Gestalt principle of proximity is sufficient -- items near each other are perceived as related. A card boundary is redundant when spacing already communicates the relationship.

Rules:
- Never nest cards inside cards. If you need cards-in-cards, the information architecture is wrong.
- Use cards only when items are independently actionable (clickable/draggable).
- For lists of similar items, a simple row with consistent spacing outperforms cards.
- For dashboards, use whitespace and subtle dividers instead of card borders.

---

## Container Queries

Viewport queries are for page-level layout decisions. Container queries are for component-level adaptation.

A card placed in a sidebar should adapt to the sidebar width, not the viewport width. If the viewport is 1400px wide but the sidebar is 300px, the card needs the narrow layout.

```css
.card-container {
  container-type: inline-size;
}

@container (min-width: 400px) {
  .card { /* horizontal layout */ }
}

@container (max-width: 399px) {
  .card { /* vertical stacked layout */ }
}
```

Use viewport queries (`@media`) for:
- Overall page layout (sidebar visible/hidden)
- Navigation mode switches (hamburger/full)

Use container queries (`@container`) for:
- All reusable components
- Any element that could live in different-width containers

---

## Optical Adjustments

Geometric centering and mathematical alignment often look wrong to the human eye. Optical adjustments correct these perception gaps.

**Text at margins**: Text set at `margin-left: 0` looks indented due to the glyph's internal sidebearing. Apply `-0.05em` negative margin-left to make text appear flush.

**Icon centering**: Icons that are geometrically centered within a circle or button often appear off-center. Common fixes:
- Play/arrow icons: shift right by 1-2px to compensate for the visual weight being left of geometric center
- Text + icon buttons: the icon's optical weight differs from text -- adjust `vertical-align` or add 1px offset

**Circle/rounded elements**: A circle inside a square that is mathematically centered appears to sit too low. Shift up by ~1px.

These adjustments are small (1-2px) but the difference between "something feels off" and "this looks right" in production.

---

## Touch Targets

Minimum interactive target size: 44x44px (WCAG 2.5.8, Apple HIG).

The visual element can be smaller. Expand the tap area using:
- `padding` on the interactive element itself
- `::before` or `::after` pseudo-elements with absolute positioning to extend the hit area beyond the visual boundary

```css
.small-icon-button {
  position: relative;
  /* Visual size: 24px */
}
.small-icon-button::before {
  content: '';
  position: absolute;
  inset: -10px; /* Expands tap target to 44px */
}
```

Never rely on the visual size alone. A 16px icon button with no padding expansion is unusable on mobile.

---

## Depth and Elevation

Use a semantic z-index scale with named tokens:

```
--z-dropdown:       100
--z-sticky:         200
--z-modal-backdrop: 300
--z-modal:          400
--z-toast:          500
--z-tooltip:        600
```

Never use arbitrary z-index values (999, 99999). They create z-index wars that are impossible to debug.

**Shadow guidelines**: Shadows communicate elevation. If a shadow is clearly visible as a distinct visual element, it is too strong. Effective shadows are subtle -- they create depth perception without drawing attention to themselves.

Layer shadow intensity by elevation level:
- Low elevation (cards, buttons): barely visible, tight offset
- Medium elevation (dropdowns, popovers): noticeable on close inspection
- High elevation (modals): visible but not harsh, wider spread

Multiple layered shadows (2-3 declarations) look more natural than a single heavy shadow. Real-world light creates soft gradients, not hard edges.
