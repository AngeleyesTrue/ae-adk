# Motion Design

Animation timing, easing, performance, and accessibility for purposeful UI motion.

---

## Duration Rule (100/300/500)

Match duration to the type of change:

| Change Type | Duration | Examples |
|---|---|---|
| Instant feedback | 100-150ms | Button press, toggle, checkbox |
| State changes | 200-300ms | Dropdown open, tab switch, color change |
| Layout changes | 300-500ms | Accordion expand, panel slide, card reorder |
| Entrance animations | 500-800ms | Page section reveal, hero animation |

**Exit rule**: Exit animations run at 75% of entrance duration. Users are already done with the element -- fast exits respect their intent. A 400ms entrance becomes a 300ms exit.

Anything over 800ms feels sluggish in UI context. Save longer durations for marketing/hero animations only.

---

## Easing

Never use the generic `ease` keyword. It is a mediocre compromise that fits nothing well.

Use specific curves matched to motion direction:

**ease-out** (entering/appearing): Element decelerates into its final position. The object arrives and settles.
```css
--ease-out: cubic-bezier(0.16, 1, 0.3, 1);
```

**ease-in** (leaving/disappearing): Element accelerates away. The object departs with increasing speed.
```css
--ease-in: cubic-bezier(0.4, 0, 1, 1);
```

**ease-in-out** (toggles/morphs): Element that stays on screen but changes state. Symmetric acceleration and deceleration.
```css
--ease-in-out: cubic-bezier(0.4, 0, 0.2, 1);
```

**Exponential curves for micro-interactions**: Sharper deceleration curves feel more responsive for small UI elements:

```css
--ease-out-quart: cubic-bezier(0.25, 1, 0.5, 1);
--ease-out-quint: cubic-bezier(0.22, 1, 0.36, 1);
--ease-out-expo:  cubic-bezier(0.19, 1, 0.22, 1);
```

Use exponential curves for tooltips, small popups, and micro-feedback where snappiness matters.

---

## BANNED: Bounce and Elastic Effects

Bounce and elastic easing are banned. They have been dated since approximately 2015.

Real physical objects decelerate smoothly -- they do not bounce at their destination. Bounce/elastic effects:
- Draw attention to the animation itself, not the content
- Feel gimmicky and unprofessional
- Break the illusion of direct manipulation
- Signal "we prioritized flair over function"

The only exception: a deliberate playful/gaming context where bouncy physics is part of the brand identity. Even then, use sparingly.

---

## Only Animate transform and opacity

These two properties are composited on the GPU. Everything else triggers layout recalculation (reflow) or paint, causing jank on lower-end devices.

**Never animate**: `width`, `height`, `top`, `left`, `margin`, `padding`, `border`, `font-size`

**For height animations** (accordion expand/collapse), use the grid-template-rows trick:

```css
.accordion-content {
  display: grid;
  grid-template-rows: 0fr;
  transition: grid-template-rows 300ms var(--ease-out);
}
.accordion-content.open {
  grid-template-rows: 1fr;
}
.accordion-content > div {
  overflow: hidden;
}
```

This animates `grid-template-rows` from `0fr` to `1fr`, achieving a smooth height animation without animating `height` directly.

---

## Staggered Animations

Stagger child elements using CSS custom properties for the index:

```css
.stagger-item {
  animation: fadeSlideIn 300ms var(--ease-out) both;
  animation-delay: calc(var(--i, 0) * 50ms);
}
```

Set `--i` on each element (0, 1, 2, 3...) either in HTML or via JavaScript.

**Cap the total stagger time**: 10 items at 50ms = 500ms total stagger. That is the maximum before the sequence feels slow. If you have 20 items, do not stagger to 1000ms -- cap at 500ms and let later items animate in groups.

For lists with unknown length, clamp the delay:

```css
animation-delay: calc(min(var(--i, 0), 10) * 50ms);
```

---

## prefers-reduced-motion

This is NOT optional. Approximately 35% of adults over 40 are affected by motion sensitivity. Vestibular disorders, migraines, and seizure conditions are common.

Implementation strategy:

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

Then selectively restore functional animations:

```css
@media (prefers-reduced-motion: reduce) {
  .progress-bar { transition-duration: 200ms !important; }
  .spinner { animation-duration: 1s !important; animation-iteration-count: infinite !important; }
}
```

Rules for reduced motion:
- Replace spatial movement (slide, fly) with crossfade (opacity only)
- Preserve functional animations: progress bars, loading spinners, scroll position indicators
- Never remove animation entirely if it communicates state (loading, success, error)
- Test with the setting enabled -- it should feel complete, not broken

---

## Perceived Performance

The 80ms threshold: any response under 80ms feels instantaneous to users. Optimize critical interaction feedback to hit this target.

**Optimistic UI** for low-stakes actions: Show the success state immediately, revert on failure. "Like" buttons, toggle switches, adding items to lists -- users expect instant response. The network round-trip happens in the background.

Do NOT use optimistic UI for:
- Payment processing
- Destructive actions (delete, send)
- Actions with side effects on other users

**Easing toward completion**: When showing progress, ease-in the last portion. A progress bar that decelerates near 100% makes the task feel shorter than linear progress. This is a well-documented psychological effect.

---

## Performance

**will-change**: Do not add `will-change` preemptively. It consumes GPU memory for every element it is applied to. Only add it when you have measured a specific jank issue and confirmed that `will-change` resolves it. Remove it after the animation completes if possible.

**Intersection Observer for scroll animations**: Never use scroll event listeners for triggering animations. Intersection Observer is purpose-built for this:

```javascript
const observer = new IntersectionObserver((entries) => {
  entries.forEach(entry => {
    if (entry.isIntersecting) {
      entry.target.classList.add('animate-in');
      observer.unobserve(entry.target); // Fire once
    }
  });
}, { threshold: 0.15 });
```

**Motion tokens**: Create a token system for durations and easings, just like spacing and color tokens. This ensures consistency and makes global timing adjustments trivial:

```css
:root {
  --duration-instant: 100ms;
  --duration-fast:    200ms;
  --duration-normal:  300ms;
  --duration-slow:    500ms;
  --ease-default:     cubic-bezier(0.16, 1, 0.3, 1);
  --ease-enter:       cubic-bezier(0.16, 1, 0.3, 1);
  --ease-exit:        cubic-bezier(0.4, 0, 1, 1);
  --ease-move:        cubic-bezier(0.4, 0, 0.2, 1);
}
```
