---
name: ae-design-impeccable
description: >
  Impeccable design system skill providing AI Slop anti-patterns,
  context gathering protocol, self-verification tests, and 7 domain
  reference modules. Eliminates generic AI-generated design patterns.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-09"
  modularized: "true"
  tags: "design, impeccable, ai-slop, anti-patterns, context-gathering, verification, typography, color, spatial, motion, interaction, responsive, ux-writing"
  related-skills: "moai-design-craft, moai-domain-uiux, agency-design-system"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["impeccable", "ai slop", "anti-pattern", "design audit", "design context", "squint test", "brand test", "tinted neutral", "OKLCH", "modular scale", "design normalize", "design polish", "typography", "font", "color", "palette", "animation", "motion", "spacing"]
  agents: ["expert-frontend", "designer"]
  phases: ["plan", "run", "review"]
---

# Impeccable Design Skill

Distinctive, production-grade frontend design that avoids generic AI aesthetics. Based on pbakaus/impeccable (Apache 2.0), restructured for ae-adk.

## Quick Reference

This skill provides five capabilities:

1. **AI Slop Anti-Patterns** — Curated ban list preventing generic AI-generated design. Two severity tiers: HARD (absolute ban) and SOFT (overridable with rationale).
2. **Context Gathering Protocol** — Mandatory pre-design information collection. No design without knowing the audience.
3. **Self-Verification Tests** — Three post-design checks: AI Slop Test, Squint Test, Brand Test.
4. **Domain Reference Modules** — Seven specialized guides: Typography, Color, Spatial, Motion, Interaction, Responsive, UX Writing.
5. **Design Commands** — Five slash commands: `/design:audit`, `/design:normalize`, `/design:polish`, `/design:critique`, `/design:extract`.

**Activation**: Loaded automatically when trigger keywords are detected by expert-frontend or designer agents. Not directly user-invocable.

---

## Core Design Principles

1. **Bold Direction Over Safe Defaults** — Commit to a clear aesthetic point-of-view. Both refined minimalism and bold maximalism work. The key is intentionality, not intensity.
2. **Context Before Code** — Never start design work without knowing the audience, brand, and purpose. Code tells you what was built, not who it's for.
3. **The AI Slop Test** — If someone said "AI made this," would they believe it immediately? That's the problem. A distinctive interface makes someone ask "how was this made?"
4. **OKLCH Over HSL** — Perceptually uniform color space. Equal lightness steps look equal. Tint all neutrals toward brand hue.
5. **Variety Over Monoculture** — Every project deserves its own aesthetic. Vary fonts, themes, layouts, and approaches. Never converge on common choices.

---

## AI Slop Anti-Patterns (Summary)

### HARD (Absolute Ban — No Exceptions)

| Category | Banned Patterns |
|----------|----------------|
| Colors | Pure black (#000), pure gray (0 chroma), purple-blue AI gradients, cyan-on-dark |
| Effects | Glassmorphism, neon glow, gradient text (background-clip:text), side-stripe borders >1px |
| Layout | Hero metric template (big number + stats + gradient), large icons above every heading |
| Copy | "In today's fast-paced world", "Unlock the potential", AI marketing cliches |

### SOFT (Default Ban — Override via .impeccable.md with rationale)

| Category | Banned Patterns |
|----------|----------------|
| Fonts | Reflex font list (Inter, Roboto, Syne, DM Sans, Playfair, 20+ others) |
| Layouts | Identical card grids, 3-column icon-text, everything centered, cards-in-cards |
| Motion | Bounce/elastic easing, generic `ease`, animating layout properties |

Full details: `${CLAUDE_SKILL_DIR}/modules/ai-slop-antipatterns.md`

---

## Context Gathering

### When Required

- New design system, page, or component design
- Brand or visual identity work

### When Skipped

- Minor style fixes, color adjustments, typo corrections (HARD anti-pattern check still applies)

### Resolution Order

1. Check loaded instructions for Design Context section
2. Check `.impeccable.md` at project root
3. Check `.agency/context/visual-identity.md` (read-only reference)
4. If none found: ask user before proceeding

Full protocol: `${CLAUDE_SKILL_DIR}/modules/context-gathering.md`

---

## .impeccable.md — P0 Minimal Schema

```markdown
# .impeccable.md - Project Design Context

## Brand
- Name: {brand_name}
- Tone: {formal|casual|playful|corporate|...}
- Target Audience: {description}

## Design Decisions
- Primary Color: {hex/oklch}
- Typography: {font_stack}

## Anti-Pattern Overrides
- {pattern}: {rationale for allowing SOFT anti-pattern exception}
```

Place at project root. HARD anti-patterns can never be overridden. Only SOFT anti-patterns with explicit rationale.

---

## Self-Verification Tests

Run after any design output is generated:

| Test | What It Checks | Pass Criteria |
|------|---------------|---------------|
| **AI Slop Test** | HARD + SOFT anti-pattern violations | Zero HARD violations, zero unoverridden SOFT violations |
| **Squint Test** | Visual hierarchy via programmatic heuristic | 2 of 3 checks pass (font-size steps, contrast differentiation, whitespace proportionality) |
| **Brand Test** | Consistency with .impeccable.md | All defined brand elements match |

Full details: `${CLAUDE_SKILL_DIR}/modules/self-verification.md`

---

## Module Index

| Module | File | Trigger Keywords | Priority |
|--------|------|-----------------|----------|
| AI Slop Anti-Patterns | `modules/ai-slop-antipatterns.md` | design, build, UI, component | P0 |
| Context Gathering | `modules/context-gathering.md` | new design, design system, brand | P0 |
| Self-Verification | `modules/self-verification.md` | review, verify, test, audit | P1 |
| Typography | `modules/typography-reference.md` | typography, font, type, modular scale | P1 |
| Color & Contrast | `modules/color-contrast.md` | color, palette, OKLCH, tinted neutral | P1 |
| Spatial Design | `modules/spatial-design.md` | layout, grid, spacing, hierarchy | P2 |
| Motion Design | `modules/motion-design.md` | animation, motion, transition, easing | P2 |
| Interaction Design | `modules/interaction-design.md` | interaction, form, focus, modal | P2 |
| Responsive Design | `modules/responsive-design.md` | responsive, mobile, breakpoint, container query | P2 |
| UX Writing | `modules/ux-writing.md` | copy, label, error message, empty state | P2 |

Modules load on-demand (Level 3 Progressive Disclosure). Only relevant modules are loaded per design session.

---

## Command Reference

| Command | Purpose | Details |
|---------|---------|---------|
| `/design:audit` | Full design quality audit (AI Slop + accessibility + brand) | `${CLAUDE_SKILL_DIR}/commands/audit.md` |
| `/design:normalize` | Detect and suggest corrections for AI Slop patterns | `${CLAUDE_SKILL_DIR}/commands/normalize.md` |
| `/design:polish` | Fine-tune spacing, alignment, color, motion details | `${CLAUDE_SKILL_DIR}/commands/polish.md` |
| `/design:critique` | Independent design critique (anti-pattern perspective) | `${CLAUDE_SKILL_DIR}/commands/critique.md` |
| `/design:extract` | Extract design context from existing website/code | `${CLAUDE_SKILL_DIR}/commands/extract.md` |

---

## Works Well With

### moai-design-craft
- **ae-design-impeccable**: Anti-pattern detection and brand consistency
- **moai-design-craft**: Intent-First design philosophy and design memory
- On conflicting recommendations: Intent-First wins
- Critique outputs: separated sections (Intent-First / Anti-Pattern)

### moai-domain-uiux
- **ae-design-impeccable**: Color theory (OKLCH, tinted neutrals), typography selection
- **moai-domain-uiux**: WCAG compliance, design tokens (DTCG standard), theming, icons
- No overlap. ae provides rules, moai provides token infrastructure.

### agency-design-system
- ae provides anti-pattern data that can inform agency's Dynamic Zone
- Actual seeding is via learner agent or manual editing (Agency Constitution)
- ae reads `.agency/context/visual-identity.md` as reference (never writes)

### moai-design-tools
- ae provides design rules and quality checks
- moai-design-tools handles rendering (Figma MCP, Pencil renderer)

---

## Usage Notes

- **Not user-invocable**: Agents load this skill automatically via trigger keyword matching
- **Token budget**: Max ~15,000 tokens concurrent (Level 2 5K + 2-3 modules ~3K each)
- **Project context**: `.impeccable.md` is project-specific, stored at project root
- **Agency context**: `.agency/context/visual-identity.md` is read-only reference
- **Upstream**: moai-* skills are never modified (upstream immutable)
- **License**: Apache 2.0, compatible with pbakaus/impeccable source
- **Reference**: `${CLAUDE_SKILL_DIR}/reference.md` for source links and tools
