---
name: go-code
description: Write Go code. Use whenever implementing, fixing, refactoring, or adding tests in Go.
---

# Go code

Make the smallest idiomatic change. Apply every relevant rule before declaring the work done.

## MUST

- When delegating, explicitly select GPT Luna with `max` reasoning. Give workers a fresh context with the task, relevant paths, constraints, required skills, and acceptance criteria.
- Keep contract decisions, planning, integration, and final review with the parent. Inspect delegated diffs and independently verify the integrated result.
- Have workers complete their assigned work directly and return blockers to the parent. Request concise results with findings or changed files, verification evidence, and unresolved issues.
- Apply `$go-contract` before changing caller-visible API, behavior, or names. Inspect real callers; settle ownership, defaults, lifecycle, and compatibility; test the contract. State ownership when slices, maps, or pointers cross the API.
- Use package context for caller-visible names. For one primary type and constructor, prefer `client.Client` and `client.New`.
- Validate before side effects. Wrap operational errors with the caller's action. Put `context.Context` first on blocking or cancellable calls and keep it scoped to that call.
- Preserve evaluation order, short-circuiting, and side effects when changing control flow. Keep independent checks as separate `if` statements when every matching body must run.
- Use `switch` for related guard ladders and alternative branches. Use a value switch for one discriminant and a boolean switch for different predicates. Keep single checks and local error handling as `if` statements; use an `if` chain otherwise only when a switch would obscure scope or change behavior.
- Keep performance visible. Think Casey Muratori. Avoid added allocations, copying, indirection, interface dispatch, contention, and poor data locality.
- Pair `foo.go` with `foo_test.go`; source stems containing `_` are exempt. Name top-level tests `Test` plus at most three camel-case words; put scenarios in `t.Run`.
- Use Testify in every test. Use `assert` for checks and `require` only when the test cannot continue.
- When Go code changes, run `gofmt` on changed Go files, focused package tests, then `go test ./...`. When tests change, also run exactly `go run ~/.agents/skills/golang/go-code/scripts/testlint.go`; use `-root` for another repository. Fix its findings before completing verification.

## SHOULD

- Delegate substantial, bounded research, code summarization, implementation, and test-writing to subagents.
- Use one worker by default. Parallelize only independent tasks with separate file ownership.
- Handle trivial edits directly. Escalate worker tasks only after a concrete blocker; when delegation is unavailable, continue locally.
- Think Rob Pike. Clear is better than clever. Make the zero value useful.
- Separate logical blocks with one blank line. Keep an operation and its immediate error check together; keep short switch cases compact.
- Put a blank line before a standalone explanatory comment that introduces a new block, except at the start of a block. Keep the comment adjacent to the code it explains. Explain intent or non-obvious constraints rather than narrating statements.
- Prefer the standard library, concrete types, and narrow interfaces at real substitution seams. Add abstractions only for current use.
- Use package vocabulary. Prefer one-word names. Use two words only for a useful distinction, with the action or qualifier first. Three words usually mean too many concepts.
- Prefer one ready constructor. Use options only for optional policy or tuning, and state every zero-option default.
- Group related cases in table tests. Put large inputs in `testdata` or fixture files.

## NEVER

- Use vague catch-all names such as `helper`, `helpers`, `util`, or `utils` for packages, files, types, or functions, including prefixes and suffixes. These are an anti-pattern under all circumstances. Name code for its concrete responsibility and keep it with the domain that owns it.
- Have workers delegate their assignments again.
- Add speculative dependencies, interfaces, options, configuration, or exports.
- Extend an exported interface for convenience, panic for an operational failure, or retain a request context after its call.
- Write a one-case boolean `switch`, write a short declaration in a `case`, split one contiguous guard ladder across several switches, or reorder checks in a way that changes which expressions run.
- Leave sentence-like `Test` names, orphan simple source or test files, large inline fixture blobs, or skipped verification without saying why.
