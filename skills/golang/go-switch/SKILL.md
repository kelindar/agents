---
name: go-switch
description: Simplify Go branching with switch, cmp.Or, and built-in min/max without changing behavior.
disable-model-invocation: true
---

# Go Switch

Prefer compact, flat control flow. Rewrite only when readability improves and behavior stays the same.

## Choose the rewrite

- SHOULD remove `else` by moving work after a terminating guard or using an ordered switch for alternatives. Preserve scope; avoid closures or duplicated work just to remove `else`.
- SHOULD use boolean switches for related guard ladders with at least two cases. Keep a single condition as `if`, including conditions joined with `||`.
- SHOULD keep one switch per decision table. Separate parsing steps and distinct phases may have separate switches. Use value switches for dispatch or enum validation when clearer than boolean comparisons.
- SHOULD retain shared initializers on the switch when converting `if` / `else if`. Preserve case order; a final `else` can become `default`:

```go
switch value, ok, err := decodeString(object, "response_format"); {
case err != nil:
	return audio.SpeechRequest{}, err
case ok:
	request.ResponseFormat = strings.ToLower(value)
}
```

- SHOULD prefer `cmp.Or(a, b, fallback)` for first-non-zero selection when zero means unset. Requires Go 1.22+ and comparable operands.
- SHOULD prefer built-in `min(a, b)` / `max(a, b)` for minimum or maximum selection. Requires Go 1.21+ and ordered operands.
- SHOULD use `min(max(value, lower), upper)` for clamping only when `lower <= upper` is established and the original code clamps rather than rejects invalid input.
- MUST respect the project's supported Go version; do not raise it just for a rewrite.

## Preserve behavior

- MUST preserve side effects, panic behavior, branch precedence, and control-flow targets. Moving an unlabeled `break` into a switch changes its target; retain the branch unless its target can be preserved clearly.
- MUST keep independent checks when multiple bodies must run. A switch selects only one case.
- MUST preserve parsing and validation order. Keep calls and mutations between guards in place. A case cannot declare variables; use a switch init or retain the `if`. Hoist work only when safe before every preceding guard.
- MUST account for eager argument evaluation in `cmp.Or`, `min`, and `max`. Keep branches when additional evaluation could cause side effects, panics, or expensive work. `cmp.Or(value, loadDefault())` always calls `loadDefault()`.
- MUST preserve absence versus explicit zero values. An absent format may keep a default while an explicitly empty format must fail validation; `cmp.Or` would erase that distinction.
- MUST preserve floating-point NaN and signed-zero behavior; comparison branches are not automatically equivalent to `min` / `max`.
- MUST preserve comments beside the code they explain and leave generated code unchanged.

## Workflow

1. Inspect repository status and establish the requested scope. Preserve unrelated changes.
2. From this skill's `branchswitch` directory, run the helper with absolute paths. It is a separate Go module:

   ```sh
   go run . -path <absolute-target-directory> -out <absolute-checklist-path>
   ```

   The helper lists functions by branch count, highest first. `-min` defaults to 3. Pass `-out` explicitly to avoid its legacy `.cursor` default.
3. Supplement the checklist with an in-scope manual pass for value selection and init-based `if` / `else if` chains. The helper skips `if` statements with `else` and misses functions below its threshold. Do not expand a scoped request into a repository-wide scan.
4. Review one function at a time, considering value selection before switches. Mark each candidate converted or skipped with a reason. Preserve notes if regenerating the checklist; the helper overwrites its output.
5. After each coherent batch, run `gofmt` on changed files and affected-package tests from the target module. Investigate failures; revert only a faulty rewrite, preserving user changes. Report unrelated failures.
6. Finish when every in-scope candidate is reviewed. Run required repository checks, inspect the diff, and summarize changes, skips, and verification limits.
