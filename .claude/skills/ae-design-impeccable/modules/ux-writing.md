# UX Writing

Microcopy, button labels, error messages, empty states, and content strategy for clear communication.

---

## Button Labels

Never use generic labels: "OK", "Submit", "Yes", "No", "Click here".

Use verb + object format that describes the specific action:
- "Save changes" (not "Save" or "OK")
- "Create account" (not "Submit")
- "Delete message" (not "Yes" in a confirmation dialog)
- "Send invitation" (not "Confirm")

**Destructive action labels**:
- "Delete" = permanent, irreversible removal
- "Remove" = recoverable, can be re-added (e.g., remove from list but item still exists)
- Always specify what is being destroyed: "Delete 5 items" not just "Delete"

**Show counts for bulk actions**: "Delete 5 items", "Move 3 files", "Archive 12 messages". The count confirms scope and prevents accidental bulk operations.

**Cancel buttons**: Use specific alternatives when possible. "Keep editing" is more reassuring than "Cancel" when dismissing an unsaved-changes dialog.

---

## Error Messages

Every error message must answer three questions:
1. What happened?
2. Why did it happen?
3. How does the user fix it?

Templates by error type:

**Format error**: "[Field] must be [correct format]. Example: [example]"
- "Email must be a valid address. Example: name@company.com"

**Missing required**: "[Field] is required to [purpose]."
- "Password is required to secure your account."

**Permission denied**: "You don't have permission to [action]. Contact [who] for access."
- "You don't have permission to delete this project. Contact the project owner for access."

**Network error**: "Could not connect to [service]. Check your connection and try again."
- "Could not connect to the server. Check your internet connection and try again."

**Server error**: "Something went wrong on our end. Please try again. If the problem continues, contact support."

Rules:
- Never blame the user: "Invalid input" implies user stupidity. "Please enter a valid email" is neutral.
- Never use technical jargon in user-facing errors: "500 Internal Server Error" means nothing to users.
- Never use passive voice: "An error was encountered" is vague. "We couldn't save your changes" is clear.
- Provide a specific recovery action, not just "try again".

---

## Empty States

Empty states are onboarding opportunities, not dead ends.

Structure: Acknowledge briefly, explain value, provide action.

GOOD:
- "No projects yet. Create your first project to start organizing your work." [Create project]
- "Your inbox is empty. Messages from your team will appear here."

BAD:
- "No items."
- "Nothing to display."
- "0 results found."

Empty state checklist:
- Explain what will appear here once populated
- Show a clear primary action to create the first item
- If the empty state is a search result, suggest broader search terms or filters to remove
- Use an illustration only if it communicates something the text does not (not just decoration)

---

## Voice vs Tone

**Voice** is the consistent personality of the product. It does not change. It is defined by the brand.

**Tone** adapts to the user's emotional moment:
- **Success**: Celebratory, positive. "Your changes are saved."
- **Error**: Empathetic, helpful. "Something went wrong. Here's how to fix it."
- **Loading/waiting**: Reassuring, specific. "Saving your draft..."
- **Destructive action**: Serious, explicit. "This will permanently delete all 5 items. This cannot be undone."
- **Onboarding**: Welcoming, encouraging. "Welcome! Let's set up your workspace."

Critical rule: Never use humor for error states. "Oops! Something went wrong" trivializes the user's frustration. Errors are not funny to the person experiencing them.

---

## Accessibility Writing

**Link text**: Must make sense standalone, out of visual context. Screen reader users navigate by link lists.
- WRONG: "Click here", "Read more", "Learn more"
- RIGHT: "View pricing plans", "Read the migration guide", "Download the 2024 annual report"

**Alt text**: Describes the information the image conveys, not the image itself.
- WRONG: "A photo of a graph" (describes the image)
- RIGHT: "Revenue grew 40% from Q1 to Q4 2024" (describes the information)
- Use `alt=""` (empty string) for purely decorative images. This tells screen readers to skip them entirely. Omitting `alt` entirely causes screen readers to read the filename.

**Icon buttons**: Must have `aria-label` when there is no visible text:
```html
<button aria-label="Close dialog">
  <svg><!-- X icon --></svg>
</button>
```

---

## Translation Planning

Design copy with translation in mind from the start. Retrofitting internationalization is 10x harder.

**Length expansion**: Translations are longer than English.
- German: +30% longer
- French: +20% longer
- East Asian languages: can be shorter but require different line heights

Design implications: buttons, labels, and navigation must accommodate 30% longer text without breaking layout. Use `min-width` not fixed `width`.

**Avoid concatenation**: Never build sentences from fragments.
- WRONG: `"You have " + count + " new " + (count === 1 ? "message" : "messages")`
- RIGHT: Use ICU MessageFormat or similar: `"You have {count, plural, one {# new message} other {# new messages}}"`

**Full sentences as single strings**: Translators need complete sentences to produce natural translations. Word order varies by language.
- WRONG: Two separate strings "Created by" and "on" assembled around a name and date
- RIGHT: Single string "Created by {author} on {date}"

**Avoid abbreviations**: "Jan", "Mon", "amt" do not translate predictably. Use the locale's date/time formatter.

---

## Terminology Consistency

Pick one term for each concept. Use it everywhere. Users build mental models around consistent vocabulary.

Common choices (pick one column):

| Concept | Option A | Option B | Option C |
|---|---|---|---|
| Remove permanently | Delete | Remove | Trash |
| Configuration | Settings | Preferences | Options |
| Undo | Undo | Revert | Restore |
| Create new | Create | New | Add |
| Duplicate | Duplicate | Copy | Clone |
| Organizing | Folders | Categories | Groups |

Once chosen, document it in a glossary and enforce it across all UI surfaces. "Settings" in the nav, "Preferences" in the menu, and "Options" in the dialog is three terms for one concept -- that is a bug.

---

## Avoid Redundancy

If the heading explains it, the introduction paragraph is redundant. If the label explains it, the placeholder is redundant.

Principles:
- Say it once, say it well
- Every word must earn its place on screen
- If removing a sentence changes nothing about the user's understanding, remove it
- Instructions that repeat the obvious insult the user's intelligence: a text field labeled "Email" does not need placeholder text "Enter your email address"

Before writing any microcopy, ask: "Does the user already know this from context?" If yes, do not write it.

---

## Loading State Copy

Be specific about what is happening:
- GOOD: "Saving your draft...", "Uploading 3 files...", "Searching 10,000 records..."
- BAD: "Loading...", "Please wait...", "Processing..."

For long waits (over 5 seconds), set expectations:
- "This may take up to 30 seconds..."
- "Processing large file (2 of 5)..."
- Show a progress indicator with the specific copy

For very long operations (over 30 seconds):
- Allow the user to navigate away with a notification on completion
- "We'll notify you when the export is ready."
