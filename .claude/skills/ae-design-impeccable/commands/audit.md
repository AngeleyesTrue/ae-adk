---
name: design:audit
description: >
  Full design quality audit combining AI Slop detection,
  accessibility reference, and brand consistency check.
  Outputs categorized findings with scores and improvement suggestions.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "command"
  status: "active"
  updated: "2026-04-09"
  tags: "design, audit, quality, ai-slop, accessibility, brand"
---

# Design Quality Audit

Comprehensive design quality audit for existing code and components. Scans for anti-patterns, accessibility issues, and brand consistency violations.

---

## Usage

```
/design:audit [target]
```

Where `target` is optional. If provided, it can be:
- A file path (e.g., `src/components/Hero.tsx`)
- A component name (e.g., `Hero`, `Navigation`)
- A page section (e.g., `landing page`, `dashboard`)

If no target is provided, audit the most recently modified UI files or the current working directory's frontend code.

---

## Audit Process

### Step 1: Discover Target Files

If a target is specified, locate the relevant files. If not, find all UI component files (`.tsx`, `.jsx`, `.vue`, `.svelte`, `.html`, `.css`, `.scss`) in the project.

Use Grep and Glob to identify files. Prioritize recently modified files if the set is large.

### Step 2: AI Slop Check (40% weight)

Scan for anti-patterns defined in the Impeccable AI Slop reference.

Load the anti-pattern definitions:
```
Read ${CLAUDE_SKILL_DIR}/modules/ai-slop-antipatterns.md
```

**HARD anti-patterns (instant fail, score = 0):**
Scan for each HARD violation. Any single HARD violation causes the entire AI Slop category to score 0. These cannot be overridden.

**SOFT anti-patterns (check for overrides):**
Scan for each SOFT violation. For each found, check whether the project's `.impeccable.md` configuration explicitly overrides it with a valid justification. Overridden SOFT violations do not count against the score.

Scoring:
- 0 HARD + 0 unoverridden SOFT = 100
- 0 HARD + 1-2 unoverridden SOFT = 80
- 0 HARD + 3-5 unoverridden SOFT = 60
- 0 HARD + 6+ unoverridden SOFT = 40
- Any HARD violation = 0 (auto-fail)

### Step 3: Accessibility Reference (30% weight)

Check basic accessibility criteria. Detailed accessibility auditing is delegated to moai-domain-uiux -- this step covers only the Impeccable baseline.

Checks:
- **Font size**: Body text >= 16px (1rem). Anything below fails.
- **Touch targets**: Interactive elements >= 44x44px effective size. Check padding and pseudo-element expansions.
- **Focus indicators**: All interactive elements must have `:focus-visible` or equivalent focus styles. `outline: none` without replacement is a fail.
- **Color contrast**: Text must meet WCAG AA (4.5:1 for normal text, 3:1 for large text). Use computed styles or design token values.
- **Alt text**: Images must have `alt` attributes. Decorative images must have `alt=""`.
- **Labels**: Form inputs must have associated `<label>` elements or `aria-label`.

Scoring:
- Each check scores pass/fail
- Score = (passing checks / total checks) * 100

### Step 4: Brand Consistency (30% weight)

Compare against the project's `.impeccable.md` Brand section.

If `.impeccable.md` exists and has a Brand section:
- **Color palette adherence**: Are the colors used in the code from the defined palette? Flag any off-palette colors.
- **Typography consistency**: Are the fonts, weights, and sizes consistent with the defined type scale?
- **Tone alignment**: Does the microcopy match the defined voice/tone (requires reading UX copy in the code)?

If `.impeccable.md` does not exist or has no Brand section:
- Score this category as "N/A - No brand configuration found"
- Suggest creating `.impeccable.md` with brand definitions
- Do not penalize the overall score

Scoring:
- Each sub-check scores 0-100
- Category score = average of sub-checks

---

## Output Format

Present findings in this structure:

```
## Design Audit Report

Target: [file/component/page audited]
Date: [current date]

---

### AI Slop Check: PASS/FAIL
Score: [X]/100 (weight: 40%)

**HARD violations: [count]**
[For each: file, line, pattern name, description]

**SOFT violations: [count] ([overridden count] overridden)**
[For each: file, line, pattern name, override status]

---

### Accessibility: PASS/WARNING
Score: [X]/100 (weight: 30%)

- Font size issues: [count]
- Touch target issues: [count]
- Focus indicator issues: [count]
- Contrast issues: [count]
- Alt text issues: [count]
- Label issues: [count]

[Details for each issue: file, line, description, fix suggestion]

---

### Brand Consistency: PASS/WARNING/N/A
Score: [X]/100 (weight: 30%)

- Color adherence: [PASS/FAIL - details]
- Typography adherence: [PASS/FAIL - details]
- Tone alignment: [PASS/FAIL - details]

---

### Overall Score: [X]/100

Calculation: (AI Slop * 0.4) + (Accessibility * 0.3) + (Brand * 0.3)

### Priority Fixes (Top 3)

1. [Most impactful improvement with specific file and action]
2. [Second most impactful improvement]
3. [Third most impactful improvement]
```

---

## Scoring Rules

- Any HARD violation = AI Slop score is 0, which makes the maximum possible overall score 60/100
- Categories scored independently, then weighted sum produces overall
- "N/A" categories are excluded from the weighted calculation (remaining categories are re-weighted proportionally)
- Score >= 80: PASS (production ready)
- Score 60-79: WARNING (improvements recommended before launch)
- Score < 60: FAIL (must fix before launch)

---

## Integration Notes

This audit is complementary to moai-design-craft's `/moai review --critique` command. The difference:

- **`/design:audit`** (this command): Anti-pattern perspective. Scans for known bad patterns and violations against measurable criteria. Objective, rule-based scoring.
- **`/moai review --critique`**: Intent-First perspective. Evaluates whether the design fulfills its stated intent. Subjective, craft-based assessment.

For a comprehensive design review, run both commands. They evaluate different dimensions of quality.

---

## Allowed Tools

Read, Grep, Glob (read-only audit -- no file modifications)
