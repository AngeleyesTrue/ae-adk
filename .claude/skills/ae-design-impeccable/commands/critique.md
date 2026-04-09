---
name: design:critique
description: >
  Independent design critique from the anti-pattern perspective.
  Complements moai-design-craft's Intent-First critique with
  AI Slop detection and brand consistency analysis.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "command"
  status: "active"
  updated: "2026-04-09"
  tags: "design, critique, review, anti-pattern, brand"
---

# design:critique

Provide an independent design critique focused on anti-patterns and brand consistency. This COMPLEMENTS moai-design-craft's critique which takes an Intent-First perspective.

## Usage

`/design:critique [target]` where target is a file, component, page, or directory to critique.

## Critique Dimensions

### AI Slop Analysis

- Would someone immediately say "AI made this"? What specific patterns reveal AI generation?
- What makes this design memorable versus forgettable?
- Does it feel like a template or like a deliberate design decision?
- Are there any "uncanny valley" design indicators: too perfect, too symmetrical, too generic?

### Distinctiveness

- Does the design have a clear point of view or personality?
- Is there a "one thing someone will remember" after seeing this?
- Is there variety in layout, rhythm, and component treatment, or is it monotonous repetition?
- Does it avoid the "same card repeated 3-6 times" pattern?

### Brand Alignment

- Does the visual language match the brand personality defined in .impeccable.md?
- Are colors, typography, and tone consistent with the brand context?
- Would this design feel at home next to the brand's other materials?
- Are there visual elements that contradict the brand's stated identity?

### Technical Quality

- Proper OKLCH color space usage for perceptual uniformity?
- Tinted neutrals instead of pure gray?
- Responsive approach that adapts layout rather than just scaling?
- Motion quality: purposeful easing, appropriate durations, reduced-motion support?

## Critique Output Format

```
## Design Critique (Anti-Pattern Perspective)

### First Impression
[What stands out immediately - both positive observations and concerns]

### AI Slop Indicators
[Specific patterns that reveal AI generation, or confirmation of genuine distinctiveness]

### Distinctiveness Score: X/10
[What makes this memorable, what makes it generic, and why the score was assigned]

### Brand Consistency
[Alignment analysis against .impeccable.md definitions and brand context]

### Top 3 Recommendations
[The three most impactful improvements, ordered by expected visual impact]
```

## Relationship with moai-design-craft

These two critique perspectives serve different purposes and are designed to work together:

- **moai-design-craft** `/moai review --critique`: Intent-First perspective evaluating whether the implementation aligns with the stated design intent
- **ae-design-impeccable** `/design:critique`: Anti-pattern perspective evaluating AI Slop indicators and brand consistency

When both critiques run on the same target, their outputs are presented as clearly separated sections. On conflicting recommendations between the two perspectives, the Intent-First critique from moai-design-craft takes precedence, since design intent is the higher-order concern.
