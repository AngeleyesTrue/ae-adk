# Typography Reference

Font selection, modular scales, fluid typography, and type system architecture. Based on pbakaus/impeccable (Apache 2.0).

---

## Font Selection Protocol

### Step 1: Define Brand Voice

Write 3 concrete words for the brand voice. NOT "modern" or "elegant" — those are dead categories.

Examples of useful descriptors:
- "warm and mechanical and opinionated"
- "calm and clinical and careful"
- "fast and dense and unimpressed"
- "handmade and a little weird"
- "sharp and minimal and confident"

### Step 2: Imagine the Physical Object

The right font matches the brand as a physical object:
- Museum exhibit caption
- Hand-painted shop sign
- 1970s mainframe terminal manual
- Fabric label on the inside of a coat
- Children's book printed on cheap newsprint
- Tax form
- Vintage airline ticket

Whichever physical object fits the three brand words is pointing at the right KIND of typeface.

### Step 3: Browse With Intention

Browse font catalogs with the physical object in mind:
- Google Fonts: https://fonts.google.com/
- Pangram Pangram: https://pangrampangram.com/
- Adobe Fonts: https://fonts.adobe.com/
- Future Fonts: https://www.futurefonts.xyz/
- ABC Dinamo: https://abcdinamo.com/
- Klim Type Foundry: https://klim.co.nz/
- Velvetyne: https://velvetyne.fr/

**Reject the first thing that "looks designy."** That is the trained-data reflex. Keep looking.

### Step 4: Cross-Check

The right font for an "elegant" brief is NOT necessarily a serif. The right font for a "technical" brief is NOT necessarily a sans-serif. The right font for a "warm" brief is NOT Fraunces. If the final pick lines up with the reflex pattern, go back to Step 3.

---

## Banned Reflex Fonts (SOFT Anti-Pattern)

These are training-data defaults that create monoculture. Override requires explicit rationale in `.impeccable.md`.

**Sans-Serif:** Inter, Roboto, Arial, Open Sans, Lato, Montserrat, DM Sans, Plus Jakarta Sans, Outfit, Instrument Sans

**Serif:** DM Serif Display, DM Serif Text, Fraunces, Newsreader, Lora, Crimson, Crimson Pro, Crimson Text, Playfair Display, Cormorant, Cormorant Garamond, Instrument Serif

**Monospace:** IBM Plex Mono, Space Mono

**Other:** IBM Plex Sans, IBM Plex Serif, Space Grotesk, Syne

**Syne is the worst offender** — the most overused "distinctive" display font and an instant AI design tell. Never use it. Extra scrutiny required for any Syne override.

---

## Modular Scale

Use fewer sizes with more contrast. A 5-size system covers most needs:

| Role | Size | Use Case |
|------|------|----------|
| xs | 0.75rem (12px) | Captions, legal text |
| sm | 0.875rem (14px) | Secondary UI, metadata |
| base | 1rem (16px) | Body text |
| lg | 1.25-1.5rem (20-24px) | Subheadings, lead text |
| xl+ | 2-4rem (32-64px) | Headlines, hero text |

### Scale Ratios

| Ratio | Name | Character |
|-------|------|-----------|
| 1.25 | Major Third | Subtle, refined |
| 1.333 | Perfect Fourth | Balanced, versatile |
| 1.5 | Perfect Fifth | Dramatic, bold |

Pick one ratio and commit. Minimum 1.25 ratio between steps for clear hierarchy.

---

## Fluid Typography

### When to Use Fluid Type (clamp)

- Headings and display text on marketing/content pages
- Text that dominates the layout and needs to breathe across viewport sizes

### When to Use Fixed rem Scales

- App UIs, dashboards, data-dense interfaces
- No major design system (Material, Polaris, Primer, Carbon) uses fluid type in product UI
- Body text (even on marketing pages — size difference across viewports too small)

### clamp() Pattern

```css
/* Fluid heading: min 2rem, scales with viewport, max 4rem */
h1 { font-size: clamp(2rem, 5vw + 1rem, 4rem); }
```

The middle value (e.g., `5vw + 1rem`) controls scaling rate. Add a rem offset so text doesn't collapse to 0 on small screens.

---

## Line Height and Measure

- Line-height scales inversely with line length: narrow columns need tighter leading, wide columns need more
- Light text on dark backgrounds: ADD 0.05-0.1 to normal line-height (light type reads as lighter weight)
- Cap line length at 65-75ch using `max-width: 65ch`
- Never disable zoom (`user-scalable=no` breaks accessibility)
- Minimum 16px body text (smaller strains eyes, fails WCAG on mobile)

---

## Font Pairing

The non-obvious truth: you often don't need a second font. One well-chosen family in multiple weights creates cleaner hierarchy than two competing typefaces.

When pairing IS needed, contrast on multiple axes:
- Serif + Sans (structure contrast)
- Geometric + Humanist (personality contrast)
- Condensed display + Wide body (proportion contrast)

**Never pair fonts that are similar but not identical** (e.g., two geometric sans-serifs). They create visual tension without clear hierarchy.

---

## OpenType Features

Most developers don't know these exist. Use them for polish:

| Feature | CSS | Use Case |
|---------|-----|----------|
| Tabular numbers | `font-variant-numeric: tabular-nums` | Data tables, alignment |
| Proper fractions | `font-variant-numeric: diagonal-fractions` | Recipes, measurements |
| Small caps | `font-variant-caps: all-small-caps` | Abbreviations |
| No ligatures | `font-variant-ligatures: none` | Code blocks |
| Kerning | `font-kerning: normal` | Body text (usually default) |

Check font features at: https://wakamaifondue.com/

---

## Text Color in OKLCH

Use OKLCH for text colors to maintain perceptual uniformity across hierarchy levels. Tint text toward the brand hue for cohesion.

| Role | OKLCH Value | Use Case |
|------|-------------|----------|
| Primary text | `oklch(20% 0.01 {brand_hue})` | Headings, body text, CTAs |
| Secondary text | `oklch(40% 0.01 {brand_hue})` | Subheadings, descriptions |
| Muted text | `oklch(55% 0.01 {brand_hue})` | Captions, metadata, timestamps |
| Disabled text | `oklch(65% 0.005 {brand_hue})` | Inactive elements |

For dark mode, invert the lightness scale:

| Role | OKLCH Value (Dark) |
|------|-------------------|
| Primary text | `oklch(92% 0.01 {brand_hue})` |
| Secondary text | `oklch(75% 0.01 {brand_hue})` |
| Muted text | `oklch(55% 0.005 {brand_hue})` |

Always verify contrast ratios meet WCAG requirements. See `${CLAUDE_SKILL_DIR}/modules/color-contrast.md` for full palette construction and accessibility guidelines.

---

## System Fonts

Underrated for performance-first apps:

```css
font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", system-ui, sans-serif;
```

Looks native, loads instantly, highly readable. Consider for apps where performance > personality.

---

## Web Font Loading

Prevent layout shift with proper loading strategy:

```css
@font-face {
  font-family: 'CustomFont';
  src: url('font.woff2') format('woff2');
  font-display: swap;
}

/* Match fallback metrics to minimize shift */
@font-face {
  font-family: 'CustomFont-Fallback';
  src: local('Arial');
  size-adjust: 105%;
  ascent-override: 90%;
  descent-override: 20%;
  line-gap-override: 10%;
}
```

Tools like Fontaine (https://github.com/unjs/fontaine) calculate overrides automatically.

---

## Token Architecture

Name tokens semantically, not by value:

| Good | Bad |
|------|-----|
| `--text-body` | `--font-size-16` |
| `--text-heading` | `--font-size-32` |
| `--text-caption` | `--font-size-12` |
| `--leading-tight` | `--line-height-1-2` |
| `--tracking-wide` | `--letter-spacing-0-05` |

Include in token system: font stacks, size scale, weights, line-heights, letter-spacing.
