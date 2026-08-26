---
name: go-contract
description: Design or review exported Go package APIs for precise naming, small contracts, idiomatic composition, explicit semantics, and durable compatibility. Use for new packages, API reviews, constructor or option cleanup, exported-name decisions, interface boundaries, lifecycle and ownership contracts, or pre-release checks.
---

# Go Contract

Design from the caller's code. Make the API feel smaller than its implementation.

## Start with the contract

1. Read the package, callers, examples, tests, and documentation. Trace real flows before changing declarations.
2. Inventory exports. Ask of each: what useful caller action becomes impossible if this stays private?
3. Write the smallest realistic call site showing construction, primary operation, errors, durability, and cleanup.
4. State ownership, concurrency, mutation, ordering, cancellation, durability, retry, and lifecycle guarantees plainly. Imprecision reveals an unfinished contract.
5. Choose the smallest surface that makes the common path obvious and misuse difficult.

## Shape the package

- Give the package one coherent job and vocabulary. Split only separate concepts.
- Prefer one constructor, safe defaults, and a ready result. Put required identity values in parameters; use private functional options for optional policy or tuning.
- Export configuration only when callers treat it as data.
- Follow “return structs, accept interfaces”: return package-owned concrete types and accept narrow interfaces at real substitution seams.
- Reuse accurate standard contracts; never imitate them incompletely.
- Add extension points only for a real second implementation or demonstrated seam.

## Name the contract

Make `package.Identifier` read naturally without repetition. Use caller-domain nouns and precise verbs. Keep one word per concept; synonyms must mean different things.

Distinguish `New` (create ready), `Open` (access existing), `Parse`/`Decode` (convert), `Get` (lookup), `Sync` (durability), `Flush` (visible buffering), `Clone` (independent ownership), and `Copy` (transfer). Use `Add`, `Append`, `Insert`, and `Put` only for their collection, order, position, and key semantics.

Reject vague exports such as `Manager`, `Processor`, `Handler`, `Service`, `Data`, `Process`, `Execute`, and `Do` unless the domain makes them exact. Name interfaces by capability. Make booleans assertions. Preserve Go initialisms.

## Make behavior explicit

- Put `context.Context` first on blocking or cancellable operations; never store caller contexts.
- Validate before side effects. Return errors, not panics. Wrap causes with caller-facing action; export identities only for stable branching.
- Document ownership, concurrency safety, what `Close` commits, whether it is idempotent or retryable, and what remains valid.
- Document ordering, boundary inclusivity, filter combination, partial progress, and retry safety.
- Mark borrowed data read-only and state its lifetime. Provide `Clone` for retention or mutation.
- Use slices for bounded eager results and iterators for large, lazy, fallible results. Avoid iteration channels.

## Protect evolution

Treat names, signatures, defaults, errors, ordering, blocking, ownership, compatibility, and `Close` as promises. Prefer additive methods, helpers, options, or narrow interfaces. Avoid extending exported interfaces.

Verify observable contracts with compiling examples, interface assertions, and focused tests for boundaries, errors, cancellation, repeated close, concurrency, and aliasing. Run `go test ./...`.

Present the call site first. Then report the surface, guarantees, risks, and exports to remove.
