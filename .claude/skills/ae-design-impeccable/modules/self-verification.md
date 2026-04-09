# Self-Verification Tests

Three verification tests to run after any design output is generated. These ensure design quality and brand consistency.

---

## Test 1: AI Slop Test

**Purpose**: Detect patterns that instantly reveal AI-generated design.

### Procedure

1. Scan all generated/modified CSS, markup, and content
2. Check against HARD anti-patterns (reference: `${CLAUDE_SKILL_DIR}/modules/ai-slop-antipatterns.md`)
3. Check against SOFT anti-patterns
4. For SOFT violations: check `.impeccable.md` Anti-Pattern Overrides for valid rationale

### HARD Anti-Pattern Checklist

- [ ] No pure black (#000) or pure white (#fff) in large areas
- [ ] No zero-chroma grays (pure gray without tint)
- [ ] No purple-blue AI gradients or cyan-on-dark palette
- [ ] No glassmorphism or neon glow effects
- [ ] No gradient text (background-clip: text with gradient)
- [ ] No side-stripe borders (border-left/right > 1px on cards/callouts)
- [ ] No decorative sparklines conveying no data
- [ ] No hero metric template layout
- [ ] No large icons above every heading
- [ ] No AI copy cliches ("In today's fast-paced world", "Unlock the potential", etc.)

### SOFT Anti-Pattern Checklist

- [ ] No reflex fonts (Inter, Roboto, Syne, etc.) without override rationale
- [ ] No identical card grids without override rationale
- [ ] No bounce/elastic easing without override rationale
- [ ] No everything-centered layouts without override rationale
- [ ] No cards-inside-cards without override rationale
- [ ] No uniform spacing (variety creates hierarchy)

### Verdict

- **PASS**: Zero HARD violations AND zero unoverridden SOFT violations
- **FAIL**: Any HARD violation (immediate, non-negotiable) OR any SOFT violation without valid .impeccable.md override

---

## Test 2: Squint Test (Programmatic Heuristic)

**Purpose**: Verify visual hierarchy exists. The original Squint Test is visual ("blur your eyes and check hierarchy"). This is the programmatic equivalent for AI agents.

### Check A: Font-Size Hierarchy

Examine all font-size declarations in the output:
- Count distinct font-size steps
- Calculate ratio between consecutive steps
- **PASS**: At least 3 distinct size steps with minimum 1.25 ratio between them
- **FAIL**: Fewer than 3 steps, or ratio < 1.25 between any consecutive steps (flat hierarchy)

### Check B: Color Contrast Differentiation

Examine text color contrast ratios against backgrounds:
- Primary content (headings, CTAs) should have highest contrast
- Secondary content (body text) should have moderate contrast
- Tertiary content (metadata, captions) should have lower contrast
- **PASS**: Contrast ratios are differentiated by hierarchy level (primary > secondary > tertiary)
- **FAIL**: All text has similar contrast ratio regardless of importance level

### Check C: Whitespace Proportionality

Examine margin and padding values:
- Headings should have more space above them than paragraph spacing
- Section separators should have more space than element separators
- Spacing should increase with hierarchy level
- **PASS**: Spacing values increase proportionally with hierarchy (heading space > paragraph space > inline space)
- **FAIL**: Uniform spacing throughout, or inverse proportionality

### Verdict

- **PASS**: 2 out of 3 checks pass
- **FAIL**: Fewer than 2 checks pass

### Severity

Squint Test failure is a **WARNING**, not a CRITICAL. The design should be improved but can proceed with justification.

---

## Test 3: Brand Test

**Purpose**: Verify consistency with project design context defined in .impeccable.md.

### Prerequisites

- `.impeccable.md` must exist at project root
- Brand section must have at least name + tone
- If .impeccable.md doesn't exist or has no Brand section: **SKIP** this test

### Checks

**Color Palette Check:**
- Primary color matches or complements the defined palette in .impeccable.md
- No off-brand colors used in prominent positions
- Neutrals tinted toward the brand hue (not toward a different hue)

**Typography Check:**
- Font stack matches or is compatible with defined typography in .impeccable.md
- If specific fonts defined, they are actually used
- Font pairing follows the defined aesthetic direction

**Tone Alignment Check:**
- Visual mood aligns with defined personality words
- Formal brand → no playful animations, casual language
- Playful brand → not overly corporate or stiff
- Technical brand → appropriate information density, not decorative
- Motion and interaction style consistent with stated tone

### Verdict

- **PASS**: All defined brand elements are consistent with .impeccable.md
- **FAIL**: Any brand element contradicts .impeccable.md definitions
- **SKIP**: No .impeccable.md or no Brand section defined

### Severity

Brand Test failure is a **WARNING**. Should be fixed for brand consistency but may proceed if justified.

---

## Overall Verification Flow

```
1. Run AI Slop Test
   ├── FAIL (HARD violation) → CRITICAL: Must fix before proceeding
   └── PASS → Continue

2. Run Squint Test
   ├── FAIL → WARNING: Should improve hierarchy
   └── PASS → Continue

3. Run Brand Test
   ├── SKIP → No .impeccable.md, acceptable
   ├── FAIL → WARNING: Should align with brand
   └── PASS → Continue

4. Final Verdict:
   - All PASS → Design approved
   - AI Slop FAIL → BLOCKED (must fix)
   - Squint/Brand FAIL → APPROVED WITH WARNINGS
```

---

## When to Run

- After generating new UI components
- After significant visual changes to existing components
- Before submitting design work for review
- As part of `/design:audit` command
- During agency GAN loop (Builder-Evaluator cycle)

## Integration

- AI Slop Test references: `${CLAUDE_SKILL_DIR}/modules/ai-slop-antipatterns.md`
- Brand Test references: `.impeccable.md` at project root
- `/design:audit` runs all three tests and produces a combined report
