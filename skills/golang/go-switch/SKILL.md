---
name: go-switch
description: >-
  Find Go functions with 3+ branch-style if statements and simplify them into
  switch statements when this improves readability without changing behavior.
  Use when simplifying repeated branching, replacing if chains with switch,
  reducing repetitive guard clauses, or making Go control flow more idiomatic.
---

# Go Switch

Use this skill when working in a Go repository and the user asks to simplify repeated branching, replace chains of `if` statements with `switch`, reduce repetitive guard clauses, or make Go control flow more idiomatic.

The goal is to find functions containing several branch checks, usually 3 or more, and rewrite them to `switch { case ...: ... }` when it makes the function shorter, clearer, and equally correct.

## Style preference

Prefer compact, readable Go.

A sequence like this:

```go
if c == nil {
	return fmt.Errorf("catalog is nil")
}
if tile == nil {
	return fmt.Errorf("tile is nil")
}
if tile.Name == "" {
	return fmt.Errorf("tile name is required")
}
```

should usually become:

```go
switch {
case c == nil:
	return fmt.Errorf("catalog is nil")
case tile == nil:
	return fmt.Errorf("tile is nil")
case tile.Name == "":
	return fmt.Errorf("tile name is required")
}
```

Use `switch {}` for boolean guard branches — but **only when there are two or more cases**.

A switch with a single `case` should stay (or be reverted to) a plain `if`. One case does not buy readability over `if`:

```go
// Prefer this
if c == nil {
	return fmt.Errorf("catalog is nil")
}

// Not this — one case is not a decision table
switch {
case c == nil:
	return fmt.Errorf("catalog is nil")
}
```

The same rule applies when comma-separated conditions collapse to one case (`case w <= 0, h <= 0, fn == nil:`) — keep a single `if` with `||` instead.

After splitting around init statements or side effects, if a boolean switch would have only one remaining case, leave that branch as `if` rather than introducing a one-case switch.

Do not force the rewrite when the resulting code is less clear.

## One switch per guard table

Produce **one** boolean `switch { case ... }` per related guard or validation ladder. Do not stack several consecutive boolean switches when the cases belong to the same decision table.

Bad — three boolean switches for one validation function:

```go
switch {
case p.ID == "":
	return fmt.Errorf("primitive id is required")
}
switch p.Kind {
case assets.PrimitiveBox, ...:
default:
	return fmt.Errorf("invalid kind")
}
switch {
case p.Width <= 0:
	return fmt.Errorf("dimensions must be positive")
}
```

Good — one boolean switch:

```go
switch {
case p.ID == "":
	return fmt.Errorf("primitive id is required")
case p.Kind != assets.PrimitiveBox && p.Kind != assets.PrimitivePrism && ...:
	return fmt.Errorf("invalid kind")
case p.Kind == assets.PrimitivePrism && p.RidgeAxis != assets.AxisX && p.RidgeAxis != assets.AxisY:
	return fmt.Errorf("invalid ridge axis")
case p.Width <= 0 || p.Depth <= 0 || p.Height <= 0:
	return fmt.Errorf("dimensions must be positive")
}
```

Use a separate `switch p.Kind` (or `switch x`) only for **value dispatch** — routing to different handlers — not duplicated alongside a boolean switch that already validates the same field.

When statements must run between guards (assignments, calls, mutations), prefer:

```go
if c == nil || worldRect.Empty() {
	return image.Rectangle{}, false
}
vw := viewportWidth(c)
vh := viewportHeight(c)
if vw <= 0 || vh <= 0 {
	return image.Rectangle{}, false
}
```

Do **not** split that into `switch` → assignment → `switch` → assignment → `switch`. Either keep the `if` chain or use one switch only where no intervening work is required.

A second switch in the same function is fine when it starts a **different phase** (e.g. guard checks, then later `switch { case full: ... default: ... }` upload strategy dispatch).

## Important Go syntax rule

Do not write this:

```go
switch {
case _, ok := c.tiles[tile.Name]; ok:
	return fmt.Errorf("duplicate tile name %q", tile.Name)
}
```

This is invalid Go.

A `case` clause cannot contain a short variable declaration. When an `if` branch has an init statement, such as:

```go
if _, ok := c.tiles[tile.Name]; ok {
	return fmt.Errorf("duplicate tile name %q", tile.Name)
}
```

handle it conservatively.

Allowed options:

1. Leave that branch as an `if`.
2. Hoist the init before the switch only if it is safe, does not change evaluation order, and does not panic before earlier guards.
3. Split into a switch for the simple branches, followed by the init `if`.

For example, this is safe and readable:

```go
switch {
case c == nil:
	return fmt.Errorf("catalog is nil")
case tile == nil:
	return fmt.Errorf("tile is nil")
case tile.Name == "":
	return fmt.Errorf("tile name is required")
}

if _, ok := c.tiles[tile.Name]; ok {
	return fmt.Errorf("duplicate tile name %q", tile.Name)
}
```

Do not hoist `_, ok := c.tiles[tile.Name]` above `case c == nil` or `case tile == nil` unless you have proved it cannot panic.

## Detection workflow

1. Run the helper tool bundled with this skill:

```bash
go run ./.cursor/skills/go-switch/branchswitch -path .
```

This writes `.cursor/skills/go-switch/branches.md` — a checklist sorted by branch count (highest first). Each line looks like:

```markdown
[ ] 7 branches, function handleInput(), file cmd/wasm/main.go
```

The tool counts guard-style `if` branches per function. Existing `switch` statements are not counted.

Optional flags:

* `-min 3` — minimum if branches to list a function (default 3)
* `-out path/to/branches.md` — output path (default `.cursor/skills/go-switch/branches.md`)

2. Open `branches.md` and work **top to bottom, one function at a time**.

3. For each unchecked line:

   * Read the function and the **Details** section for that entry.
   * Decide whether related `if` branches should become `switch { case ... }`.
   * Convert when it improves readability without changing behavior.
   * If not worth converting, leave the code as-is.
   * Mark the line `[x]` when done (converted or reviewed and skipped).
   * Re-run the tool after a batch if you want an updated checklist.

4. After every change, run:

```bash
gofmt -w <changed-files>
go test ./...
```

5. If tests fail or the rewrite looks worse, revert that function and still mark the todo `[x]` with a skip note in the Details section if helpful.

## Rewrite rules

### Good candidates

Convert consecutive or near-consecutive simple branches:

```go
if x == nil {
	return errX
}
if y == nil {
	return errY
}
if name == "" {
	return errName
}
```

to:

```go
switch {
case x == nil:
	return errX
case y == nil:
	return errY
case name == "":
	return errName
}
```

The `if` statements do not have to be literally adjacent, but they should be part of the same logical decision block.

### Acceptable branch bodies

These are usually safe:

```go
return ...
continue
break
panic(...)
log.Fatal(...)
```

Multi-statement bodies may be converted only when they are still clear:

```go
case invalid:
	log.Warn("invalid input")
	return errInvalid
```

### Avoid rewriting

Do not rewrite when:

* There is only one guard branch (or only one case would remain after splitting).
* The `if` has an `else`.
* The condition or body has subtle side effects.
* The branch depends on variables mutated by earlier non-branch statements.
* The branch has an init statement that cannot be safely preserved.
* The resulting switch is longer or less readable.
* The rewrite would create **multiple back-to-back boolean switches** that should be one guard table (see above).
* The code is generated.
* The function already has a clearer domain-specific structure.
* The branches are unrelated checks scattered across unrelated phases of a long function.
* `"side_effect_only": true` and multiple branches may all run in one call.

### Preserve comments

Move comments with the branch they describe.

Before:

```go
// Catalog must be present.
if c == nil {
	return fmt.Errorf("catalog is nil")
}
```

After:

```go
switch {
case c == nil:
	// Catalog must be present.
	return fmt.Errorf("catalog is nil")
}
```

Prefer comment placement that `gofmt` keeps clean.

## Iterative agent behavior

When asked to apply this skill:

1. Inspect the repository status first.
2. Run the helper tool to generate or refresh `branches.md`.
3. Read the checklist; start with the highest branch count.
4. Process **exactly one function per iteration** (convert or consciously skip).
5. Mark that checklist line `[x]` before moving to the next.
6. Run `gofmt` and `go test ./...` after each conversion.
7. Continue until every checklist line is `[x]`.
8. Do not rewrite questionable functions just to reduce the branch count.
9. Summarize converted functions and skipped functions at the end.

## Checklist output

The helper writes markdown like:

```markdown
[ ] 7 branches, function handleInput(), file cmd/wasm/main.go
[ ] 4 branches, function transformedRect(), file internal/render/camera.go
```

A **Details** section lists branch conditions, init statements, and side-effect-only bodies to review.

## Quality bar

A good rewrite should:

* Reduce visual repetition.
* Keep the same evaluation order.
* Preserve exact behavior.
* Avoid cleverness.
* Compile cleanly.
* Make the next reader understand the decision table faster.
* Use **one boolean switch per guard table**, not several stacked switches.
* Have **at least two cases**; otherwise keep a plain `if`.
