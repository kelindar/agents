---
name: to-spec
description: Turn the current conversation into a standalone incomplete SPEC-NNN.md file in the project's specs/todo folder — no interview or external services.
disable-model-invocation: true
---

This skill takes the current conversation context and codebase understanding and produces a spec (also known as a PRD). Do not interview the user; synthesize what is already known and make reasonable assumptions explicit in the spec.

The project keeps specs in two folders: `specs/todo` contains incomplete specs and `specs/live` contains fully implemented specs. New specs belong in `specs/todo`; move one to `specs/live` only after its implementation is complete.

The skill is fully standalone. Do not use an issue tracker, labels, setup skills, or external services.

## Process

1. Identify the project root. Use the repository root when inside a repository; otherwise use the current working directory.

2. Explore the project enough to understand its current state. Use the project's domain vocabulary and respect relevant ADRs.

3. Determine the next spec ID from files directly inside `<project-root>/specs/todo` and `<project-root>/specs/live` named `SPEC-<ID>.md`. Use one greater than the highest existing ID, or `000` when none exist. Format the ID with at least three digits (`000`, `001`, ..., `999`, `1000`). Ignore files outside these folders and files that do not match this naming scheme.

4. Create `specs/todo` if needed and save the incomplete spec as:

   `<project-root>/specs/todo/SPEC-<ID>.md`

5. Prefer the highest existing test seam that exercises the feature's external behavior. Record the chosen seam and any assumptions in the spec; do not pause for confirmation.

6. Apply the `write-for-humans` skill to the draft so the spec is readable, concrete, and human-sounding without changing its meaning or required structure.

7. Write only the local spec file using the template below. Do not publish it elsewhere. Report the created path when finished.

<spec-template>

## Problem Statement

The problem that the user is facing, from the user's perspective.

## Solution

The solution to the problem, from the user's perspective.

## User Stories

A numbered list of user stories covering the feature. Each user story should use this format:

1. As an <actor>, I want a <feature>, so that <benefit>

<user-story-example>
1. As a mobile bank customer, I want to see the balance on my accounts, so that I can make better-informed decisions about my spending.
</user-story-example>

## Implementation Decisions

A list of implementation decisions, including relevant:

- Modules to build or modify
- Interfaces to modify
- Technical clarifications
- Architectural decisions
- Schema changes
- API contracts
- Specific interactions

Do not include specific file paths or code snippets because they may become outdated.

Exception: if a prototype produced a snippet that captures a decision more precisely than prose can (such as a state machine, reducer, schema, or type shape), include only the decision-rich portion and note that it came from a prototype.

## Testing Decisions

A list of testing decisions, including:

- Tests of external behavior rather than implementation details
- The modules or system boundaries to test
- Similar existing tests that provide prior art

## Out of Scope

The things that are outside this spec.

## Further Notes

Assumptions and any other relevant notes.

</spec-template>
