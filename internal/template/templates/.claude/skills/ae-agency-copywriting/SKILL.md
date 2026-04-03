---
name: ae-agency-copywriting
description: >
  Copy rules, tone, structure, and anti-patterns for AI Agency website content.
  Enforces brand voice consistency, section-level copy structure, and JSON output
  contracts for automated page generation workflows.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-04"
  tags: "agency, copywriting, brand-voice, content"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["copywriting", "copy", "brand voice", "content", "headline"]
  agents: ["agency-copywriter"]
  phases: ["build"]
---

# Agency Copywriting

Copy rules and structure for AI Agency website content production.

## Brand Voice Rules

- Use concrete numbers over vague claims ("3x faster" not "much faster")
- Avoid AI slop phrases: "revolutionize", "cutting-edge", "seamless", "leverage"
- Write in active voice
- Keep sentences under 20 words where possible
- Lead with benefits, not features

## Section Copy Structure

Each page section must have:
- Headline (max 8 words)
- Subheadline (max 20 words)
- Body copy (2-3 paragraphs max)
- CTA text (max 5 words)

## Anti-Patterns

- Generic placeholder text ("Lorem ipsum", "Your text here")
- Buzzword-heavy copy without substance
- Inconsistent tone between sections
- Missing CTA on actionable sections
