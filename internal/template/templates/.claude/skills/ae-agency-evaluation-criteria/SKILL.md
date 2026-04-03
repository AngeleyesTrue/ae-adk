---
name: ae-agency-evaluation-criteria
description: >
  Quality evaluation criteria for AI Agency project output covering design quality,
  originality, completeness, and functionality scoring with weighted dimensions
  and testing requirements.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-04"
  tags: "agency, evaluation, quality, scoring, testing"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["evaluation", "scoring", "quality check", "review"]
  agents: ["agency-evaluator"]
  phases: ["build"]
---

# Agency Evaluation Criteria

Quality scoring rubric for Agency creative production output.

## Scoring Dimensions

### Design Quality (30%)
- 0.25: Major visual issues, broken layout
- 0.50: Functional but generic, minor inconsistencies
- 0.75: Good visual fidelity, consistent tokens
- 1.00: Pixel-perfect, responsive, polished

### Originality (20%)
- 0.25: Direct template copy
- 0.50: Modified template with some custom elements
- 0.75: Unique layout with creative touches
- 1.00: Distinctive visual identity, memorable

### Completeness (30%)
- 0.25: Missing major sections
- 0.50: Core sections present, extras missing
- 0.75: All sections present, minor gaps
- 1.00: Full coverage including edge cases

### Functionality (20%)
- 0.25: Major broken interactions
- 0.50: Basic interactions work
- 0.75: All interactions work, good UX
- 1.00: Smooth interactions, loading states, error handling

## Pass Threshold

Overall weighted score >= 0.75 required to pass.
