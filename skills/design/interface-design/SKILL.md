---
name: interface-design
description: Design and refine product interfaces, or maintain concise project design rules in specs/DESIGN.md. Use for dashboards, tools, settings, admin panels, and other interactive products that need clear hierarchy, reusable tokens, complete states, or visual verification. Not for marketing or brand-only work.
---

# Interface design

Make the interface feel specific to its product and easy to use. Apply every relevant rule before declaring the work done.

## Workflow

1. Inspect the existing app, design system, tokens, components, and `specs/DESIGN.md` if present.
2. Identify the person, their main task, and the intended feel. For a new direction, name the product's domain, natural color world, one signature element, and the obvious defaults to avoid.
3. Recommend a direction with concrete reasons. Ask for confirmation only when the choice is ambiguous or costly to reverse.
4. Reuse the project's components and conventions. Patch the implementation and run relevant checks.
5. Verify non-trivial UI in a browser or rendered specimen at desktop and mobile widths. Fix visual defects before presenting.

## MUST

- Give each view one focal point. Make it win through position, space, size, weight, or contrast. Deliberately demote secondary content.
- Choose a coherent visual direction from the product's domain. Carry it through layout, color, typography, density, navigation, and component treatment.
- Use a consistent type scale, spacing unit, radius scale, surface hierarchy, and one depth strategy.
- Use semantic tokens for foreground, background, borders, controls, brand, and status. Keep at least four text levels: primary, secondary, tertiary, and muted.
- Prefer native controls, then the project's accessible primitives, then a proven headless primitive. Hand-roll stateful controls only when no suitable primitive exists, including the full keyboard, focus, and ARIA behavior.
- Reuse the existing design system and styling convention. Extract a shared component on the second real reuse.
- Cover default, hover, active, focus, and disabled states for controls. Cover loading, empty, error, unavailable, and populated states for data.
- Keep hit areas at least 40px, ideally 44px. Use tabular numerals for changing values and respect `prefers-reduced-motion`.
- Animate only when it helps orientation or feedback. Prefer `transform` and `opacity`, exact transition properties, fast ease-out timing, and faster exits.
- Verify the result visually. At minimum, check hierarchy while squinting, product specificity with the name removed, consistency with tokens, and responsive behavior.
- Treat visual verification as a completion criterion. If the user chooses to skip it or the environment blocks it, respect that constraint but report the result as visually unverified.

## SHOULD

- Let weight and tone do more hierarchy work than small font-size differences.
- Group related controls tightly and separate major regions with more space. Vary density with purpose.
- Use quiet borders and small surface shifts. Inputs should read as inset; popovers should sit above their parent surface.
- Keep color scarce and meaningful. Use one main accent unless the domain requires more.
- Use concentric radii for nested shapes, optical alignment for icons, balanced heading wraps, and readable body line height.
- Render a small specimen when a visual tool is available and the direction is easier to judge by seeing it. Keep reasoning in text and implementation in the project.
- Keep user-facing design rationale short. State the recommendation and the reason, not a private design monologue.

## NEVER

- Ship a generic dashboard template, default typography, flat hierarchy, repeated card grids, or decorative color without a product-specific reason.
- Use a `div` as a button, rebuild an installed control, or omit keyboard and focus behavior.
- Mix depth strategies, random spacing, unrelated radii, multiple arbitrary accent colors, or hardcoded colors where semantic tokens exist.
- Use harsh borders, dramatic shadows, large radii on small controls, or different hues merely to distinguish surface levels.
- Use negative margins, escape-hatch `calc()`, or absolute positioning to evade the intended layout flow.
- Animate frequent actions, use `transition: all`, animate layout properties, start entrances at `scale(0)`, or ignore reduced-motion preferences.
- Present non-trivial UI work without visual verification. State the gap if the environment prevents it.
- Call a component complete while required interaction or data states remain deferred.

## Project design file

Treat `specs/DESIGN.md` as the project's durable design contract.

- Read it before proposing or implementing UI changes.
- After establishing or changing reusable design decisions, offer to create or update it.
- When the user asks to save reusable design decisions, create or update exactly `specs/DESIGN.md`.
- Keep it project-specific and concise. Record durable recommendations, exact tokens or measurements, and reusable component patterns. Omit process notes, one-off choices, and design history.
- Merge new guidance into the existing rule instead of repeating it.
- Keep project design rules in this file, not global memory or `.interface-design/system.md`.

Use this structure:

```markdown
# Design

## MUST

- [Required project-specific design rule.]

## SHOULD

- [Preferred project-specific recommendation.]

## NEVER

- [Project-specific failure to avoid.]
```

If the user asks for design status, audit `specs/DESIGN.md` against the current UI. If the file is absent, infer repeated rules from the implementation and propose the smallest useful version.

## Commands

- `/interface-design:design-review` performs a strict craft and hierarchy review, with before and after rendering when available.
- `/interface-design:design-deslop` performs a fast, diff-scoped visual cleanup.
