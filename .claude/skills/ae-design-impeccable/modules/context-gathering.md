# Context Gathering Protocol

Design without context produces generic output. This protocol defines when and how to collect project design context before any design work.

---

## When Context Gathering is Required

**Full context gathering required:**
- New design system creation
- New page design
- New component design (non-trivial)
- Brand or visual identity work
- Major redesign or visual overhaul

**Context gathering skipped (HARD anti-pattern check still applies):**
- Minor CSS fixes (color adjustments, spacing tweaks)
- Typo corrections in existing UI
- Bug fixes that don't change visual design
- Accessibility fixes (contrast, focus indicators)

---

## Gathering Order

Check sources in this priority order. Stop at the first source that provides sufficient context.

### 1. Loaded Instructions (Instant)

If the current agent's loaded instructions contain a **Design Context** section, proceed immediately. This is the fastest path.

### 2. .impeccable.md (Fast)

Read `.impeccable.md` from the project root. If it exists and contains Brand + Design Decisions sections, proceed.

### 3. .agency/context/visual-identity.md (Read-Only Reference)

If `.impeccable.md` doesn't exist, check `.agency/context/visual-identity.md`:
- **READ-ONLY** — never write to this file (Agency Constitution compliance)
- Use as default values to pre-fill .impeccable.md
- If visual-identity.md has brand information, create .impeccable.md with those defaults
- Reverse sync (.impeccable.md → visual-identity.md) is manual only or via `/agency brief`

### 4. Ask User (Required)

If no source provides context, ask the user. Do NOT attempt to infer context from the codebase — code tells you what was built, not who it's for or what it should feel like.

---

## Required Information Checklist

### Brand (Minimum Required)

- [ ] **Brand name**: The product or company name
- [ ] **Personality** (3 concrete words): NOT "modern" or "elegant" — those are dead categories. Use specific descriptors like "warm and mechanical and opinionated" or "calm and clinical and careful"
- [ ] **Emotional goals**: What should the interface make users feel? (confidence, delight, calm, urgency)

### Audience

- [ ] **Who uses this**: Role, demographics, expertise level
- [ ] **Context of use**: When, where, on what device, under what conditions
- [ ] **Jobs to be done**: What tasks are users trying to accomplish

### Tone and Theme

- [ ] **Tone**: formal / casual / playful / corporate / technical / organic / editorial / raw
- [ ] **Theme derivation**: Light or dark based on USER CONTEXT, not default preference:
  - Trading dashboard, late-night sessions → dark
  - Hospital portal, anxious patients on phone → light
  - Children's reading app → light
  - Developer tools, dark office → dark
  - Wedding planning, Sunday morning → light
  - Music player, headphone listening at night → dark
  - Food magazine, coffee break browsing → light

### Technical Constraints

- [ ] **CSS approach**: Tailwind / CSS Modules / Styled Components / vanilla CSS
- [ ] **Framework**: React / Vue / Svelte / vanilla / other
- [ ] **Performance requirements**: Loading time targets, bundle size constraints
- [ ] **Browser support**: Modern only / IE11+ / specific requirements

### Anti-References

- [ ] **What this should NOT look like**: Specific sites, apps, or styles to avoid
- [ ] **Reference sites**: Sites that capture the right feel (specify WHAT about them)

---

## .impeccable.md P0 Minimal Schema

The minimal viable design context file. Created at project root (not in .moai/ or .claude/).

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
- {pattern_name}: {rationale for allowing this SOFT anti-pattern}
```

### Schema Rules

1. **Brand section**: Always required. Minimum: name + tone.
2. **Design Decisions section**: Populated as decisions are made during design sessions. Empty at creation is acceptable.
3. **Anti-Pattern Overrides section**: ONLY for SOFT anti-patterns. Each entry needs explicit rationale. HARD anti-patterns can NEVER appear here.
4. **Location**: Project root (`/.impeccable.md`), not in `.moai/` or `.claude/`
5. **Version control**: Committed to the repository (project-specific design decisions are shared knowledge)

---

## P3 Extended Schema (Future)

The P0 schema extends to include Session History, Grid System, and Motion preferences:

```markdown
# .impeccable.md - Project Design Context

## Brand
- Name: {brand_name}
- Tone: {formal|casual|playful|corporate|...}
- Target Audience: {description}

## Design Decisions
- Primary Color: {hex/oklch}
- Typography: {font_stack}
- Grid System: {spec}
- Motion: {preference}

## Anti-Pattern Overrides
- {pattern}: {rationale}

## Session History
- {date}: {decision summary}
```

---

## Agency Integration Rules

1. `.agency/context/visual-identity.md` is always READ-ONLY from this skill's perspective
2. If visual-identity.md has TBD values, do not use those as defaults
3. If visual-identity.md has concrete values, use as defaults for .impeccable.md Brand section
4. Never trigger writes to visual-identity.md (Agency Constitution Section 3)
5. For reverse sync needs, instruct user to run `/agency brief` or edit manually
6. Both files can coexist — .impeccable.md takes precedence for design decisions

---

## Context Freshness

- Reload .impeccable.md at the start of each design session
- If significant design decisions change during a session, update .impeccable.md immediately
- Session History (P3 schema) tracks changes over time for continuity
- Stale context is better than no context — use existing .impeccable.md even if dated
