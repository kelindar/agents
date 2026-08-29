---
name: go-code
description: Write Go code. Use whenever implementing, fixing, refactoring, or adding tests in Go.
---

# Go code

Make the smallest idiomatic change. Apply every relevant rule before declaring the work done.

## MUST

- Apply `$go-contract` before changing caller-visible API, behavior, or names. Inspect real callers; settle ownership, defaults, lifecycle, and compatibility; test the contract. State ownership when slices, maps, or pointers cross the API.
- Use package context for caller-visible names. For one primary type and constructor, prefer `client.Client` and `client.New`.
- Validate before side effects. Wrap operational errors with the caller's action. Put `context.Context` first on blocking or cancellable calls and keep it scoped to that call.
- Preserve evaluation order and side effects. Write related ladders of three or more simple guards as one boolean `switch` when that is clearer.
- Keep performance visible. Think Casey Muratori. Avoid added allocations, copying, indirection, interface dispatch, contention, and poor data locality.
- Pair `foo.go` with `foo_test.go`; source stems containing `_` are exempt. Name top-level tests `Test` plus at most three camel-case words; put scenarios in `t.Run`.
- Use Testify in every test. Use `assert` for checks and `require` only when the test cannot continue.
- When tests change, run exactly `go run ~/.agents/skills/golang/go-code/scripts/testlint.go`; use `-root` for another repository. Fix its findings, run `gofmt` on changed Go files, focused package tests, then `go test ./...`.

## SHOULD

- Think Rob Pike. Clear is better than clever. Make the zero value useful.
- Prefer the standard library, concrete types, and narrow interfaces at real substitution seams. Add abstractions only for current use.
- Use package vocabulary. Prefer one-word names. Use two words only for a useful distinction, with the action or qualifier first. Three words usually mean too many concepts.
- Prefer one ready constructor. Use options only for optional policy or tuning, and state every zero-option default.
- Group related cases in table tests. Put large inputs in `testdata` or fixture files.

## NEVER

- Add speculative dependencies, interfaces, options, configuration, or exports.
- Extend an exported interface for convenience, panic for an operational failure, or retain a request context after its call.
- Write a one-case boolean `switch`, a short declaration in a `case`, several switches for one guard table, or reorder checks in a way that changes which expressions run.
- Leave sentence-like `Test` names, orphan simple source or test files, large inline fixture blobs, or skipped verification without saying why.
