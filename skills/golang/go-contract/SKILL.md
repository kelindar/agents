---
name: go-contract
description: Design and review Go package contracts. Use before any change to caller-visible API or behavior.
---

# Go contract

Caller-visible behavior is the API. Minimize concepts, not declarations.

## Establish the contract

Inspect exports, real callers, examples, tests, and docs. Draft the smallest realistic common call, including failure and cleanup when relevant. Keep an export only when it enables a useful caller action.

Settle each relevant promise: ownership, mutation and aliasing, concurrency, blocking and cancellation, ordering and bounds, partial progress and retry, durability, and lifecycle. The contract is ready when callers need not infer behavior.

## Decide

- Give the package one job and vocabulary. Split concepts when callers use them independently.
- Prefer one constructor with a ready result. Use `New` when the package constructs one main type; add a type word only to distinguish peers. Put required identity in parameters; reserve options for optional policy or tuning. State an explicit zero-option default for each option. Name option functions for the policy, such as `Retries(3)`; a `With` prefix adds no meaning. Export configuration when callers use it as data.
- Return package-owned concrete types. Accept narrow interfaces at real substitution seams. Add extension points for current alternate implementations, not hypothetical ones. Honor standard contracts in full.
- Make `package.Identifier` natural and omit words supplied by the package name. Prefer one-word functions, methods, and types. For two words, put the action or qualifier first (`ReadFile`, `MaxSize`, `FileStore`). Use a third word only when two cannot preserve a caller-visible distinction. Name interfaces by capability, booleans as assertions, and Go initialisms conventionally.
- Use semantic verbs: `New` creates ready values; `Open` accesses existing resources; `Parse` and `Decode` convert; `Get` looks up; `Sync` makes durable; `Flush` exposes buffered work; `Clone` creates independent ownership; `Copy` transfers. Use `Add`, `Append`, `Insert`, and `Put` according to their collection semantics.
- Treat `Manager`, `Processor`, `Handler`, `Service`, `Data`, `Process`, `Execute`, and `Do` as vague until the domain gives them exact meaning.
- Put `context.Context` first on blocking or cancellable calls and retain it only for that call.
- Validate before side effects. Return operational failures as errors; reserve panics for violated programmer invariants. Wrap errors with the caller's action. Export error identities only for stable branching.
- State ownership, concurrency safety, ordering, bounds, aliasing, partial progress, retry safety, and `Close` effects, repeatability, retryability, and post-close validity. Mark borrowed data read-only with a lifetime; provide `Clone` for independent retention or mutation.
- Use slices for bounded eager results, iterators for large or fallible streams, and channels when concurrent delivery is part of the contract.
- Treat names, signatures, defaults, errors, ordering, blocking, ownership, and `Close` behavior as compatibility commitments. Prefer additive changes. Extending an exported interface breaks implementers; prefer a sibling interface or concrete method.

Verify each changed commitment with a compiling call site, interface assertion, or focused test. Run narrow package tests, then `go test ./...`. For review-only work, report caller impact and compatibility risk.
