---
name: design:normalize
description: >
  Detect AI Slop patterns in existing code and suggest corrections.
  Scans CSS, component files, and markup for anti-pattern violations.
  Provides specific before/after code suggestions.
user-invocable: true
metadata:
  version: "1.0.0"
  category: "command"
  status: "active"
  updated: "2026-04-09"
  tags: "design, normalize, ai-slop, correction, fix"
---

# design:normalize

Detect and auto-correct AI Slop patterns in existing code. Functions as a linter for design quality.

## Usage

`/design:normalize [target]` where target is a file path, component name, or directory.

## Detection Patterns

### CSS / Tailwind Scan

- Pure black/white/gray values: #000, #fff, #808080 and equivalents
- Gradient text effects: background-clip:text with gradient backgrounds
- Oversized decorative borders: border-left or border-right greater than 1px
- Bounce and elastic easing: cubic-bezier curves mimicking bounce or elastic motion
- Glassmorphism patterns: backdrop-filter:blur combined with semi-transparent backgrounds

### Markup Scan

- Icon grids with identical cards: repeated card structures with icon + heading + text, all same dimensions
- Hero metric templates: "100+", "10K+", "99%" counter sections with identical styling
- Centered-everything layouts: every section using text-center or items-center with no alignment variation

### Content Scan

- AI copy cliches: "In today's fast-paced world", "Seamlessly integrate", "Elevate your", "Leverage the power of", "Cutting-edge solutions"
- Filler paragraphs with no specific claims or concrete details
- Identical sentence structures repeated across sections

## Correction Behavior

For each detection, provide:
1. **Location**: File path and line number
2. **Finding**: What was detected and the specific pattern matched
3. **Problem**: Why this pattern is problematic (references SKILL.md anti-pattern rules)
4. **Replacement**: Specific replacement code with explanation

### Pattern Priority

- **HARD patterns**: Always suggest a fix. These are universally problematic regardless of project context.
- **SOFT patterns**: Check `.impeccable.md` overrides first. Only suggest a fix if no valid override exists for this pattern.

## Output Format

List of findings organized by file, with before/after suggestions for each detection. The command does NOT auto-apply changes. All suggestions are presented for user approval before any modifications.

## Integration

Typically run after `/design:audit` to fix issues discovered during the audit pass. Can also be run independently as a periodic code quality check.

The normalize command reads the same `.impeccable.md` configuration as other design commands, respecting all brand overrides and SOFT pattern exemptions defined there.
