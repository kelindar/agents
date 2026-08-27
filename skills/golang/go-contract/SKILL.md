---
name: go-contract
description: Review caller-visible Go contracts before any API change. Use whenever an agent considers adding, removing, renaming, or changing an exported declaration or observable package behavior, including signatures, defaults, errors, ownership, lifecycle, concurrency, or compatibility.
---

# Go contract

Caller code is the contract. Minimize the concepts a caller must learn, not merely the number of declarations.

## Work from caller code

1. Trace the package, every exported declaration, real callers, examples, tests, and documentation. The inventory is complete when each export maps to a useful caller action or is marked as a removal candidate.
2. Write the smallest realistic call site for the common path. Include construction, primary work, errors, durability, and cleanup only when the package owns those concerns. The call site is complete when its operation order and error path are visible.
3. Check every relevant contract dimension: ownership, mutation, concurrency, blocking, cancellation, ordering, boundaries, partial progress, retry, durability, and lifecycle. The contract is complete when callers need not guess any relevant behavior.
4. Choose the smallest set of declarations that supports the call site, then apply the rules below.

## Shape the package

- Give the package one job and one vocabulary. Split concepts only when callers use them independently.
- Prefer one constructor, useful defaults, and a ready result. Put required identity in parameters. Reserve options for optional policy or tuning, and keep their representation package-owned.
- Export configuration when callers must construct, inspect, compare, or serialize it as data.
- Return package-owned concrete types. Accept narrow interfaces at demonstrated substitution boundaries.
- Adopt a standard contract only when the package honors its full semantics.
- Add an extension point when a current second implementation or caller seam requires it.

## Name the contract

Make `package.Identifier` read naturally without repetition. Use caller-domain nouns and precise verbs. Keep one word for each concept, and give synonyms distinct meanings.

Distinguish `New` (create ready), `Open` (access existing), `Parse`/`Decode` (convert), `Get` (lookup), `Sync` (durability), `Flush` (visible buffering), `Clone` (independent ownership), and `Copy` (transfer). Use `Add`, `Append`, `Insert`, and `Put` only for their collection, order, position, and key semantics.

Treat `Manager`, `Processor`, `Handler`, `Service`, `Data`, `Process`, `Execute`, and `Do` as warning signs. Keep one only when the domain gives it an exact meaning. Name interfaces by capability, write booleans as assertions, and preserve Go initialisms.

## Make behavior explicit

- Put `context.Context` first on blocking or cancellable operations. Use it for that call's lifetime.
- Validate before side effects. Report operational failures as errors, and reserve panics for broken programmer invariants. Wrap causes with the caller's action. Export error identities only for stable branching.
- Document ownership and concurrency safety. For `Close`, state what it commits, whether callers may repeat or retry it, and what remains valid afterward.
- Document ordering, boundary inclusivity, filter combination, partial progress, and retry safety.
- Mark borrowed data read-only and state its lifetime. Provide `Clone` when callers need independent retention or mutation.
- Use slices for bounded eager results and iterators for large, lazy, fallible results. Use channels when concurrent delivery is part of the contract, not merely for iteration.

## Protect evolution

Treat names, signatures, defaults, errors, ordering, blocking, ownership, and `Close` behavior as promises. Evolve with additive methods, helpers, options, or sibling narrow interfaces. Keep existing exported interfaces unchanged.

Verify each changed promise with a compiling example, an interface assertion, or a focused test. Cover the relevant boundaries, errors, cancellation, repeated close, concurrency, and aliasing cases. Run the narrow package tests, then `go test ./...`. Verification is complete when the common call site compiles and every changed promise has evidence.

Present the call site first. Then report the proposed or reviewed surface, explicit guarantees, compatibility risks, and exports to remove or keep private. Separate required contract changes from optional polish.
