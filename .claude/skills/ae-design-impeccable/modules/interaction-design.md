# Interaction Design

States, forms, modals, keyboard navigation, and gesture patterns for robust interactive UI.

---

## Eight Interactive States

Every interactive element must account for eight states. Missing states create broken experiences.

| State | Description | Key Consideration |
|---|---|---|
| Default | Resting appearance | Must clearly signal interactivity |
| Hover | Mouse pointer over element | Touch devices never see this -- do not hide functionality behind hover |
| Focus | Keyboard navigation landed here | Design separately from hover -- keyboard users never trigger hover |
| Active | Being pressed/clicked | Brief visual feedback confirming the press registered |
| Disabled | Not currently available | Reduced opacity (0.4-0.5) + `cursor: not-allowed` + `pointer-events: none` |
| Loading | Processing an action | Replace content with spinner or skeleton, preserve element dimensions |
| Error | Action failed | Red/danger color + icon + descriptive message |
| Success | Action completed | Green/success color + icon + brief confirmation |

Critical mistake: designing hover and focus as the same style. They serve different users with different input methods. A keyboard user tabbing through a form sees focus states, never hover. Design both independently.

---

## Focus Rings

**Never** use `outline: none` without providing a replacement focus indicator. This is an accessibility violation that makes keyboard navigation impossible.

Use `:focus-visible` to show focus rings only for keyboard navigation (not mouse clicks):

```css
:focus-visible {
  outline: 2px solid var(--color-focus);
  outline-offset: 2px;
}

:focus:not(:focus-visible) {
  outline: none;
}
```

Focus ring requirements:
- Thickness: 2-3px (1px is invisible on many displays)
- Contrast ratio: 3:1 minimum against surrounding background
- Offset: 2-3px from the element edge (prevents visual collision with borders)
- Color: A dedicated focus color that works on both light and dark backgrounds

For elements with rounded corners, use `border-radius` on the outline (or use `box-shadow` as an alternative focus indicator that respects border-radius).

---

## Form Design

**Placeholders are not labels.** Always use a visible `<label>` element. Placeholders disappear on input, leaving users with no context for what they typed. Placeholder text also fails contrast requirements by design (it must look different from actual input).

**Validation timing:**
- Validate on blur (when the user leaves the field), not on every keystroke
- Exception: password strength meters can update on keystroke because the feedback is encouraging, not corrective
- Show errors immediately on blur; clear errors as soon as the user begins correcting
- Never validate while the user is still typing -- it creates anxiety

**Error placement:**
- Show error messages directly below the relevant field
- Connect errors to fields with `aria-describedby`
- Use both color AND an icon/text -- color alone fails for colorblind users
- Group related errors at the top of the form only for submission-level validation

```html
<label for="email">Email</label>
<input id="email" type="email" aria-describedby="email-error" />
<span id="email-error" role="alert">Please enter a valid email address</span>
```

---

## Loading States

Preference hierarchy for loading feedback:

1. **Optimistic UI** (best): Show the result immediately. Revert on failure. Use for low-stakes: likes, toggles, list additions.
2. **Skeleton screens** (good): Show a preview of the content shape. Users perceive this as faster than spinners because it sets expectations about what is coming.
3. **Spinners** (acceptable): Use only when content shape is unpredictable. Always pair with descriptive text ("Saving your draft..." not just "Loading...").
4. **Progress bars** (for long operations): Use when the operation has measurable progress. Indeterminate bars are just horizontal spinners.

Never use optimistic UI for:
- Payment processing
- Sending messages to other users
- Destructive or irreversible actions

---

## Modals

Use native `<dialog>` element with `showModal()`. It provides:
- Built-in backdrop
- Focus trapping (focus stays inside the dialog)
- Escape key dismissal
- Proper accessibility semantics

Apply the `inert` attribute to all content behind the modal. This prevents screen readers and keyboard from accessing background content:

```html
<main id="app" inert><!-- background content --></main>
<dialog id="modal"><!-- modal content --></dialog>
```

For tooltips, dropdowns, and non-modal overlays, prefer the **Popover API** (`popover` attribute) over custom solutions. It handles:
- Light dismiss (click outside to close)
- Proper stacking context
- Accessibility semantics

---

## Dropdown Positioning

Never use `position: absolute` inside a container with `overflow: hidden`. The dropdown will be clipped at the container boundary, appearing cut off or invisible.

Solutions in order of preference:
1. **CSS Anchor Positioning** (modern): Declarative positioning that respects viewport boundaries
2. **Popover API**: Built-in stacking that escapes overflow contexts
3. **`position: fixed`** with manual coordinate calculation: Calculate position from `getBoundingClientRect()`, then apply fixed positioning. Handles scroll correctly.

For dropdown direction (above vs below), check available viewport space and flip when there is insufficient room below the trigger element.

---

## Destructive Actions

Undo is better than confirmation dialogs. Users click through "Are you sure?" dialogs mindlessly -- muscle memory bypasses the safety gate.

Pattern:
1. Execute the destructive action immediately (or mark for deletion)
2. Show an undo toast/snackbar with a timer (5-10 seconds)
3. Actually delete after the timer expires
4. If the user clicks "Undo", restore immediately

This is faster for the common case (the user meant to delete) and safer for the edge case (accidental deletion can be reversed).

Reserve confirmation dialogs only for:
- Actions affecting other users (removing team members, publishing)
- Truly irreversible actions with no undo path (account deletion)
- Bulk operations on large datasets

---

## Keyboard Navigation

**Roving tabindex** for component groups: In a group of related items (toolbar buttons, radio group, tab list), only one item should be in the tab order at a time. Arrow keys move focus within the group.

```
Tab → enters the group (focuses the active/first item)
Arrow keys → move between items within the group
Tab → exits the group to the next focusable element
```

This prevents tab-trapping users in long lists. A 20-item toolbar should require 1 Tab to enter and 1 Tab to leave, not 20 Tabs to traverse.

**Skip links**: Every page needs a skip-to-content link as the first focusable element. Keyboard users should not have to tab through the entire navigation on every page:

```html
<a href="#main-content" class="sr-only focus:not-sr-only">
  Skip to main content
</a>
```

---

## Gesture Discoverability

Swipe gestures, long press, and drag operations are invisible by default. Users cannot discover what they cannot see.

Rules:
- Always provide a visible fallback for every gesture action (a button, menu item, or icon)
- Hint at swipe gestures with slight visual peek (show the edge of the action behind the item)
- Use animation or onboarding to teach unfamiliar gestures on first use
- Never make gestures the only way to perform an action
