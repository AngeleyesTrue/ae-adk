# Responsive Design

Mobile-first strategy, content-driven breakpoints, input detection, and cross-device adaptation.

---

## Mobile-First

Write base styles for mobile. Add complexity upward with `min-width` media queries.

Desktop-first (max-width) means mobile devices download and parse styles they immediately override. Mobile-first means the smallest devices get only the styles they need. Larger screens add layout complexity incrementally.

```css
/* Base: mobile (single column, stacked) */
.layout { display: flex; flex-direction: column; }

/* Tablet and up */
@media (min-width: 768px) {
  .layout { flex-direction: row; }
}

/* Desktop and up */
@media (min-width: 1024px) {
  .layout { max-width: 1200px; margin-inline: auto; }
}
```

This is not just a performance optimization -- it forces you to design for the most constrained environment first, ensuring the core experience works everywhere.

---

## Content-Driven Breakpoints

Do not chase device sizes. The iPhone 15 Pro Max is 430px today; next year it will be something else. Device-based breakpoints are a maintenance treadmill.

Instead:
1. Start at the narrowest viewport (320px)
2. Slowly stretch the viewport wider
3. When the design breaks or looks awkward, add a breakpoint at that width
4. Repeat until the widest reasonable viewport (1440-1920px)

Three breakpoints usually suffice for most layouts. Common content-driven values cluster around 640px, 768px, and 1024px, but use whatever your content dictates.

**`clamp()` for fluid values**: Instead of jumping between fixed values at breakpoints, use `clamp()` for smooth scaling:

```css
h1 {
  font-size: clamp(1.75rem, 4vw + 0.5rem, 3rem);
}

.container {
  padding-inline: clamp(1rem, 5vw, 3rem);
}
```

`clamp(minimum, preferred, maximum)` creates fluid typography and spacing that scales continuously rather than jumping at breakpoints.

---

## Input Method Detection

Screen width tells you nothing about input method. A 1024px-wide iPad uses touch. A 360px-wide Galaxy phone docked to a monitor uses a mouse.

Use interaction media queries:

```css
/* Touch device: larger targets, no hover effects */
@media (pointer: coarse) {
  .button { min-height: 48px; padding: 12px 24px; }
}

/* Mouse device: smaller targets acceptable, hover available */
@media (pointer: fine) {
  .button { min-height: 36px; }
}

/* Device supports hover */
@media (hover: hover) {
  .card:hover { transform: translateY(-2px); }
}

/* Device does NOT support hover */
@media (hover: none) {
  /* Do not rely on hover for any functionality */
  /* Show all information upfront */
}
```

Critical rule: Never hide essential functionality behind hover. It must be accessible on touch devices that have no hover state. Hover enhancements are decorative only.

---

## Safe Areas

Modern devices have notches, rounded corners, and dynamic islands that overlap content. Handle these with `env()` safe area insets.

```css
body {
  padding-top: env(safe-area-inset-top);
  padding-right: env(safe-area-inset-right);
  padding-bottom: env(safe-area-inset-bottom);
  padding-left: env(safe-area-inset-left);
}
```

Requires the viewport meta tag to include `viewport-fit=cover`:

```html
<meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover">
```

Without `viewport-fit=cover`, the browser adds its own safe area padding that you cannot control, wasting screen space. With it, you get full-bleed control and explicit safe area management.

Apply safe area padding to:
- Fixed/sticky navigation bars (top and bottom)
- Bottom action bars and tab bars
- Floating action buttons near screen edges

---

## Responsive Images

Use `srcset` with width descriptors and the `sizes` attribute for resolution-appropriate image loading:

```html
<img
  srcset="image-400.jpg 400w, image-800.jpg 800w, image-1200.jpg 1200w"
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
  src="image-800.jpg"
  alt="Description"
/>
```

The browser selects the smallest image that satisfies the layout width at the device's pixel density. A 400px-wide slot on a 2x display needs the 800w image.

Use `<picture>` for **art direction** -- when different viewports need different image crops, not just different sizes:

```html
<picture>
  <source media="(max-width: 768px)" srcset="hero-mobile.jpg" />
  <source media="(max-width: 1200px)" srcset="hero-tablet.jpg" />
  <img src="hero-desktop.jpg" alt="Description" />
</picture>
```

Art direction examples: a wide landscape shot on desktop cropped to a close-up portrait on mobile. Different compositions, not just rescaled versions.

---

## Layout Adaptation Patterns

**Navigation**: Three progressive stages based on available width:
1. **Narrow**: Hamburger menu (off-canvas)
2. **Medium**: Compact bar (icons + key items, overflow menu)
3. **Wide**: Full horizontal navigation

**Tables**: Tables are inherently horizontal. On narrow screens:
- Transform rows into card-like stacked layouts
- Use `data-label` attributes to repeat column headers per cell
- Or use horizontal scroll with a sticky first column

**Progressive disclosure**: Use `<details>` / `<summary>` to collapse secondary content on mobile, expanding it on desktop:

```html
<details open class="md:open">
  <summary class="md:hidden">Advanced options</summary>
  <div><!-- content shown by default on desktop, collapsible on mobile --></div>
</details>
```

---

## Testing

DevTools device simulation is necessary but insufficient. It does not capture:
- Actual touch behavior and gesture precision
- Real rendering performance and frame drops
- Physical screen readability and contrast
- Browser-specific quirks (Safari's viewport height, Chrome's address bar)
- Network conditions on real cellular connections

Minimum real device testing matrix:
- **iPhone** (Safari): Most common mobile browser, unique viewport behavior
- **Android** (Chrome): Most common mobile OS, widest hardware range
- **Tablet**: Test the "awkward middle" between mobile and desktop layouts

A cheap Android device (~$100) reveals performance issues that high-end development machines and simulators hide. If it runs smoothly on budget hardware, it runs smoothly everywhere.
