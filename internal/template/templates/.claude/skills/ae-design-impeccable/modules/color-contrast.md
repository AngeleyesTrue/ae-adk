# Color & Contrast Reference

OKLCH color spaces, tinted neutrals, palette construction, and dark mode. Based on pbakaus/impeccable (Apache 2.0).

---

## OKLCH Over HSL

**Stop using HSL.** Use OKLCH instead. It is perceptually uniform — equal steps in lightness LOOK equal.

HSL problem: 50% lightness in yellow looks bright, 50% in blue looks dark. OKLCH fixes this.

```
oklch(lightness chroma hue)
- Lightness: 0-100% (0 = black, 100 = white)
- Chroma: ~0-0.4 (color intensity, 0 = gray)
- Hue: 0-360 (color wheel angle)
```

### Critical OKLCH Rule

**Reduce chroma as you approach white or black.** High chroma at extreme lightness looks garish.

| Lightness | Appropriate Chroma |
|-----------|-------------------|
| 85-100% (near white) | 0.03-0.08 |
| 50-85% (midtones) | 0.08-0.20 |
| 15-50% (darker) | 0.05-0.15 |
| 0-15% (near black) | 0.005-0.05 |

Example: Light blue at 85% lightness wants ~0.08 chroma, not the 0.15 of the base color.

### Hue Selection

The hue is a brand decision. Do NOT reach for:
- Blue (hue ~250) by reflex — dominant AI default
- Warm orange (hue ~60) by reflex — second most common AI default

The hue should come from the brand, not from "cool = tech" or "warm = friendly" formulas.

---

## Tinted Neutrals

**Pure gray is dead.** A neutral with zero chroma feels lifeless next to colored brand elements.

Add 0.005-0.015 chroma to ALL neutrals, tinted toward the brand hue:

```css
/* Instead of pure gray */
--neutral-50: oklch(97% 0 0);        /* Dead, lifeless */
--neutral-500: oklch(50% 0 0);       /* Dead, lifeless */

/* Use tinted neutrals (example: brand hue = 250, blue) */
--neutral-50: oklch(97% 0.005 250);  /* Subtle, alive */
--neutral-500: oklch(50% 0.01 250);  /* Cohesive */
```

The hue you tint toward should come from THIS project's brand:
- If brand is teal → neutrals lean teal
- If brand is amber → neutrals lean amber
- NOT from "warm = friendly, cool = tech" formulas

**Avoid** always tinting warm orange or always tinting cool blue. Those create their own monoculture.

---

## The 60-30-10 Rule

About visual WEIGHT, not pixel count:

| Portion | Role | Examples |
|---------|------|----------|
| 60% | Neutral surfaces | Backgrounds, white space, base surfaces |
| 30% | Secondary | Text, borders, inactive states |
| 10% | Accent | CTAs, highlights, focus states |

Accents work BECAUSE they are rare. Overuse kills their power.

Common mistake: using the accent color everywhere because it's "the brand color."

---

## Palette Structure

| Role | Purpose | Specification |
|------|---------|--------------|
| Primary | Brand, CTAs, key actions | 1 color, 3-5 lightness shades |
| Neutral | Text, backgrounds, borders | 9-11 shade scale (tinted, not gray) |
| Semantic | Success, error, warning, info | 4 colors, 2-3 shades each |
| Surface | Cards, modals, overlays | 2-3 elevation levels |

**Skip secondary/tertiary unless you need them.** Most apps work fine with one accent color.

---

## Theme Derivation

Theme (light vs dark) should be DERIVED from audience and viewing context, not picked by default.

| Context | Theme | Reason |
|---------|-------|--------|
| Trading dashboard, fast sessions | Dark | Reduced eye strain, focus |
| Hospital portal, anxious patients late at night | Light | Calming, readable |
| Children's reading app | Light | Natural, familiar |
| Motorcycle forum, garage at 9pm | Dark | Matches environment |
| SRE observability dashboard, dark office | Dark | Matches environment |
| Wedding planning, Sunday morning | Light | Bright, optimistic |
| Music player, headphone listening at night | Dark | Matches context |
| Food magazine, coffee break | Light | Appetizing, energetic |

Do NOT default to light "to play it safe." Do NOT default to dark "to look cool." Both are lazy defaults.

---

## WCAG Contrast Requirements

| Content Type | AA Minimum | AAA Target |
|--------------|-----------|------------|
| Body text | 4.5:1 | 7:1 |
| Large text (18px+ or 14px bold) | 3:1 | 4.5:1 |
| UI components, icons | 3:1 | 4.5:1 |
| Non-essential decorations | None | None |

**Gotcha**: Placeholder text still needs 4.5:1. That light gray placeholder? Usually fails WCAG.

### Dangerous Combinations

- Light gray text on white (the #1 accessibility fail)
- Gray text on any colored background (looks washed out — use a shade of the background color)
- Red text on green (8% of men can't distinguish)
- Blue text on red (vibrates visually)
- Yellow text on white (almost always fails)
- Thin light text on images (unpredictable contrast)

---

## Never Pure Black or Pure Gray

Pure gray (`oklch(50% 0 0)`) and pure black (`#000`) don't exist in nature. Real shadows and surfaces always have a color cast. Even 0.005 chroma is enough to feel natural.

---

## Dark Mode

### Dark Mode Is Not Inverted Light Mode

| Light Mode | Dark Mode |
|------------|-----------|
| Shadows for depth | Lighter surfaces for depth (no shadows) |
| Dark text on light | Light text on dark (reduce font weight) |
| Vibrant accents | Desaturate accents slightly |
| White backgrounds | Never pure black — dark surfaces at oklch 12-18% |

### Surface Elevation Scale

Depth comes from surface lightness, not shadow. Build a 3-step scale:

```css
--surface-base: oklch(15% 0.01 {brand_hue});    /* Lowest */
--surface-raised: oklch(20% 0.01 {brand_hue});   /* Cards, panels */
--surface-overlay: oklch(25% 0.01 {brand_hue});  /* Modals, dropdowns */
```

Use the SAME hue and chroma as brand color, only vary lightness.

### Dark Mode Typography

Reduce body text weight slightly (e.g., 350 instead of 400). Light text on dark reads as heavier than dark text on light. Add 0.05-0.1 to line-height for the same reason.

### Token Architecture

Two layers: primitive tokens and semantic tokens.

```css
/* Primitives (same in all themes) */
--blue-500: oklch(55% 0.15 250);

/* Semantic (redefined per theme) */
--color-primary: var(--blue-500);
--color-surface: var(--neutral-100);  /* Light mode */
--color-surface: var(--neutral-900);  /* Dark mode */
```

For dark mode, only redefine the semantic layer. Primitives stay the same.

---

## Alpha Is a Design Smell

Heavy use of transparency (`rgba`, `hsla`, `oklch(X% X X / 0.5)`) usually indicates an incomplete palette.

Problems:
- Unpredictable contrast (depends on background)
- Performance overhead (compositing)
- Inconsistency across surfaces

Alternative: Define explicit overlay colors for each context.

Exception: Focus rings and interactive states where see-through is functionally needed.

---

## Testing Tools

- WebAIM Contrast Checker: https://webaim.org/resources/contrastchecker/
- Browser DevTools → Rendering → Emulate vision deficiencies
- Polypane: https://polypane.app/ (real-time accessibility testing)
- OKLCH Picker: https://oklch.com/
