# AI Slop Anti-Patterns

Curated ban list preventing generic AI-generated design. Two severity tiers with specific forbidden patterns and recommended alternatives.

Based on pbakaus/impeccable (Apache 2.0). Restructured for ae-adk with HARD/SOFT classification per SPEC-DESIGN-001.

---

## Severity Classification

- **HARD**: Absolute ban. No exceptions. No context override. These patterns are instant AI design tells.
- **SOFT**: Default ban. Override allowed via `.impeccable.md` Anti-Pattern Overrides section with explicit rationale.

---

## HARD Anti-Patterns (Absolute Ban)

### Colors

**H-CLR-01: Pure Black**
- Banned: `#000000`, `#000`, `rgb(0,0,0)`, `oklch(0% 0 0)`
- Why: Pure black doesn't exist in nature. Real shadows and surfaces always have a color cast.
- Alternative: Use tinted dark — `oklch(8-12% 0.005-0.01 {brand_hue})`. Even 0.005 chroma feels natural.

**H-CLR-02: Pure Gray (Zero Chroma)**
- Banned: Any gray with zero chroma — `#808080`, `#ccc`, `#999`, `oklch(X% 0 0)`, `hsl(0, 0%, X%)`
- Why: Pure gray feels lifeless next to colored brand elements. Creates no subconscious cohesion.
- Alternative: Tinted neutrals — add 0.005-0.015 chroma toward brand hue. `oklch(50% 0.01 {brand_hue})`.

**H-CLR-03: Pure White for Large Areas**
- Banned: `#ffffff`, `#fff` as background or large surface color
- Why: Like pure black, pure white never appears in nature and creates harsh contrast.
- Alternative: Off-white with subtle tint — `oklch(98% 0.005 {brand_hue})`.

**H-CLR-04: The AI Color Palette**
- Banned: Cyan-on-dark, purple-to-blue gradients, neon accents on dark backgrounds
- Why: The dominant AI-design default palette from 2024-2025. Instant recognition as AI output.
- Alternative: Derive colors from brand context. The hue should come from the brand, not from "cool = tech" formulas.

### Effects

**H-EFX-01: Glassmorphism**
- Banned: `backdrop-filter: blur()` combined with semi-transparent backgrounds used decoratively
- Why: Overused decorative pattern that adds visual noise without purpose.
- Alternative: Use purposeful transparency only where see-through serves function (overlays, modals). Define explicit overlay colors.

**H-EFX-02: Neon Glow**
- Banned: `box-shadow` with high-spread, high-opacity colored glows, `text-shadow` with colored glow
- Why: "Cool" without design intent. The lazy default for dark-mode AI interfaces.
- Alternative: Subtle, purposeful shadows. If clearly visible, it's too strong.

**H-EFX-03: Gradient Text**
- Banned: `background-clip: text` (or `-webkit-background-clip: text`) combined with any gradient background
- Why: Decorative rather than meaningful. Top three AI design tells.
- Alternative: Solid color for text. Use weight or size for emphasis, not gradient fill.

**H-EFX-04: Side-Stripe Borders**
- Banned: `border-left` or `border-right` with width >1px on cards, list items, callouts, alerts
- Includes: Hard-coded colors AND CSS variables — `border-left: 3px solid var(--color-warning)` is still banned
- Why: The single most overused "design touch" in admin/dashboard/medical UIs. Never looks intentional.
- Alternative: Full borders, background tints, leading numbers/icons, or no visual indicator at all. Do NOT swap to `box-shadow: inset`.

**H-EFX-05: Decorative Sparklines**
- Banned: Tiny charts that look sophisticated but convey no meaningful data
- Why: Visual noise masquerading as information density.
- Alternative: Show real data or remove. Every chart must answer a question.

**H-EFX-06: Generic Drop Shadows on Rounded Rectangles**
- Banned: `border-radius` + `box-shadow` as the only visual treatment for cards/containers
- Why: Safe, forgettable, could be any AI output.
- Alternative: Vary treatment — some cards use borders, some use background tints, some use nothing.

### Layout

**H-LAY-01: Hero Metric Template**
- Banned: Big number + small label + supporting stats arranged in the standard AI dashboard pattern with gradient accent
- Why: The most recognizable AI layout for dashboards and landing pages.
- Alternative: Design metrics display specific to the content. Not every number needs a "hero card."

**H-LAY-02: Icon Headers**
- Banned: Large icons with rounded corners placed above every heading
- Why: Rarely adds value. Makes sites look templated.
- Alternative: Use icons only where they genuinely aid comprehension. Most headings don't need icons.

### Copy

**H-CPY-01: AI Marketing Cliches**
- Banned phrases (non-exhaustive):
  - "In today's fast-paced world..."
  - "Unlock the potential of..."
  - "Seamlessly integrate..."
  - "Leverage cutting-edge..."
  - "Revolutionize your workflow..."
  - "Take your X to the next level"
  - "Empower your team to..."
  - "Streamline your process..."
  - "Harness the power of..."
  - "Transform the way you..."
- Why: Instant recognition as AI-generated filler text.
- Alternative: Write specific, concrete copy about what the product actually does. Use real numbers, real outcomes, real user language.

---

## SOFT Anti-Patterns (Default Ban, Overridable)

Override mechanism: Add entry to `.impeccable.md` Anti-Pattern Overrides section with explicit rationale. Only SOFT patterns can be overridden.

### Fonts

**S-FNT-01: Reflex Font List**
- Default banned fonts:
  - Sans-serif: Inter, Roboto, Arial, Open Sans, Lato, Montserrat, DM Sans, Plus Jakarta Sans, Outfit, Instrument Sans
  - Serif: DM Serif Display, DM Serif Text, Fraunces, Newsreader, Lora, Crimson, Crimson Pro, Crimson Text, Playfair Display, Cormorant, Cormorant Garamond, Instrument Serif
  - Monospace: IBM Plex Mono, Space Mono
  - Mixed: IBM Plex Sans, IBM Plex Serif, Space Grotesk, Syne
- Why: Training-data defaults that create monoculture across projects.
- Override example: `Inter: Chosen for data-dense dashboard where readability at small sizes is critical and personality is not the goal.`
- Special: **Syne is the worst offender** — "the most overused distinctive display font, instant AI design tell." Extra scrutiny required for any Syne override.

**S-FNT-02: Monospace as Technical Shorthand**
- Banned: Using monospace typography as lazy shorthand for "developer/technical" vibes
- Why: Real technical products use proper type hierarchies.
- Override example: `Monospace body: Terminal emulator UI where monospace is functionally required.`

**S-FNT-03: Single Font Family**
- Banned: Using only one font family for the entire page
- Why: Pair a distinctive display font with a refined body font for hierarchy.
- Override example: `Single font: Design system uses one variable font across all weights for consistency.`

### Layouts

**S-LAY-01: Identical Card Grids**
- Banned: Same-sized cards with icon + heading + text pattern repeated endlessly
- Why: The most common AI layout pattern. Monotonous.
- Override example: `Feature comparison grid: Equal cards required for fair visual comparison of plan tiers.`

**S-LAY-02: Three-Column Icon-Text**
- Banned: Three columns of icon + text as a default section layout
- Why: Overused AI landing page pattern.
- Override example: `Feature highlights: Three core features genuinely best presented as equal-weight items.`

**S-LAY-03: Everything Centered**
- Banned: Centering all text and elements
- Why: Left-aligned text with asymmetric layouts feels more designed.
- Override example: `Marketing hero: Centered hero section for formal brand presentation.`

**S-LAY-04: Cards Inside Cards**
- Banned: Nesting card components within card components
- Why: Visual noise. Flatten hierarchy using spacing, typography, and subtle dividers.
- Override example: Generally no valid override — flatten instead.

**S-LAY-05: Uniform Spacing**
- Banned: Same padding/margin everywhere
- Why: Without rhythm, layouts feel monotonous. Vary spacing for hierarchy.
- Override example: Not typically overridable — spacing variety is always better.

### Motion

**S-MOT-01: Bounce/Elastic Easing**
- Banned: `cubic-bezier` curves producing bounce or elastic overshoot effects
- Why: Dated since ~2015. Real objects decelerate smoothly. Draws attention to animation, not content.
- Override example: `Playful onboarding: Bounce on a game-like tutorial for children's product where playful feel is brand-critical.`

**S-MOT-02: Generic Ease**
- Banned: `transition-timing-function: ease` as default
- Why: Compromise that's rarely optimal. Use specific curves for specific contexts.
- Override example: Not typically overridable — specific curves always better.

**S-MOT-03: Layout Property Animation**
- Banned: Animating `width`, `height`, `padding`, `margin`
- Why: Causes layout recalculation. Performance anti-pattern regardless of context.
- Override example: Not recommended — performance impact is universal. Use `transform` and `opacity` instead. For height: `grid-template-rows: 0fr → 1fr`.

### Other

**S-OTH-01: Default Dark Mode**
- Banned: Defaulting to dark mode "to look cool" without deriving from user context
- Why: Theme should be derived from audience and viewing context, not aesthetic preference.
- Override example: `Dark default: Developer tools product where users work in dark environments and community expectation is dark.`

**S-OTH-02: Modals as Default**
- Banned: Using modals when better alternatives exist (inline expansion, new page, drawer)
- Why: Modals are lazy. They interrupt flow and require dismissal.
- Override example: `Confirmation modal: Irreversible destructive action requiring explicit confirmation.`
