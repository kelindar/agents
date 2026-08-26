---
name: to-tickets
description: Turn a to-spec spec and approved agent plan into the smallest useful set of local or tracker tickets. Use after grilling, to-spec, and planning when publishing work without cluttering GitHub; default to one delivery issue and add sub-issues only for independently runnable work.
---

# To Tickets

Package agreed work for execution. Treat the spec as the authority on product scope and the approved agent plan as the authority on implementation order. Do not reopen either discussion.

## Process

### 1. Read the source of truth

1. Resolve the project root.
2. Read the explicitly referenced spec, or the newest relevant `specs/SPEC-<ID>.md`, in full.
3. Use the latest approved agent plan in the conversation. If no plan exists, derive the minimum execution order from the spec without interviewing the user unless the sources materially conflict.
4. Explore only enough code to verify that ticket boundaries and dependencies match the current system. Use project vocabulary and respect relevant ADRs.

Preserve the spec's user stories, implementation decisions, testing decisions, out-of-scope items, and assumptions. Tickets may organize that scope; they must not expand or reinterpret it.

### 2. Spend issues sparingly

Start with one delivery issue for the whole spec. Keep the agent plan as a checklist inside it.

Create a sub-issue only when it is useful on its own because it:

- can be assigned to a different agent or worked in parallel;
- can be merged and verified independently;
- has a real blocking edge that matters to scheduling; or
- cannot reasonably fit in one fresh context window.

Keep setup, prefactoring, tests, documentation, cleanup, migrations, and final verification in the issue that needs them unless they independently meet a rule above. Do not create issues for individual files, architectural layers, user stories, or ordinary plan steps.

If one issue can deliver the spec, create no parent-plus-child duplicate: the single issue is the delivery issue. Prefer at most five sub-issues; exceed that only when the spec contains more independently mergeable delivery units.

For a wide refactor that cannot land green as one change, use expand–migrate–contract sub-issues. Batch migrations only as required by blast radius.

### 3. Draft the hierarchy

Make each issue a vertical, verifiable slice. For every proposed sub-issue, be able to state why it deserves a separate tracker object; otherwise fold it into its parent checklist.

Add only genuine blockers. A preferred order is not a blocking edge. Every acceptance criterion must trace to the spec, and the complete hierarchy must cover the spec without adding scope.

If the agent plan was already approved, publish without repeating a granularity quiz. Otherwise show the compact hierarchy once and get approval before writing to an external tracker.

### 4. Publish idempotently

Before creating anything, search open and closed tickets for the spec ID, spec path, and matching titles. Reuse or update the existing delivery issue and sub-issues instead of creating duplicates.

- **GitHub or another tracker:** create or reuse the delivery issue, then attach separately justified work as native sub-issues. Use a task list when native sub-issues are unavailable. Do not create a new label vocabulary; apply an existing agent-ready label only to currently unblocked leaf issues when appropriate.
- **Local fallback:** when no tracker is configured, write the same hierarchy under `.scratch/<feature-slug>/issues/`, with `00-<feature-slug>.md` for the delivery issue and numbered child files only when needed.

Create sub-issues in dependency order so blocking references resolve. Do not close the delivery issue.

## Delivery issue template

```markdown
## Spec

<repo-relative reference to SPEC-NNN.md, when available>

## Outcome

<the spec's solution in concise user-facing terms>

## Acceptance criteria

- [ ] <externally observable criterion traced to a user story>

## Execution plan

- [ ] <implementation step, or linked sub-issue when separately justified>

## Testing

- [ ] <highest useful external-behaviour test seam from the spec>

## Out of scope

- <relevant exclusions from the spec>
```

## Sub-issue template

```markdown
## Parent

<delivery issue reference>

## Outcome

<the independently verifiable vertical slice>

## Acceptance criteria

- [ ] <criterion traced to the parent spec>

## Blocked by

<real blocking issue references, or "None">
```

Avoid specific file paths and code snippets except for the repo-relative spec reference or a decision-rich prototype excerpt already preserved by the spec.
