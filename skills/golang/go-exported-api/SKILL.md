---
name: go-exported-api
description: >-
  Maintains types.md — documents exported Go API (constants, functions, types,
  methods) with brief purpose notes. No naming review. Use for /go-exported-api,
  types.md updates, or validating exported API docs.
disable-model-invocation: true
---

# Go exported API

Maintain repo-root **`types.md`**. Document what exists — no rename suggestions.

## Format

```markdown
# Exported API

## `packagename` (`module/import/path`)

### Constants
- `Name` — what it is and why it exists

### Variables
- `Name` — what it is and why it exists

### Functions
- `Name` — what it is and why it exists

### `TypeName` — what it is and why it exists
- `MethodName` — what it is and why it exists
```

- One section per non-`main` package; omit empty subsections.
- Nest methods under their type; include interface method sets on the interface type.
- Descriptions: doc comment if present, else one factual line (what + why).

## Scope

- Include `internal/` (project API boundary here).
- Skip `main`, `*_test.go`, `*_gen.go`, `tools/`, `.agents/`.

## Workflow

1. Read existing `types.md` (if any).
2. Regenerate inventory:

```bash
go run ./.agents/skills/go-exported-api/scripts/goapi write -internal
```

3. Append branch delta (optional):

```bash
go run ./.agents/skills/go-exported-api/scripts/goapi diff --base main --internal --full --markdown >> types.md
```

4. Validate:

```bash
go run ./.agents/skills/go-exported-api/scripts/goapi eval --internal
```

Exit `0` = in sync. Exit `1` = missing or stale symbols.

For signature-level deltas in changed files only:

```bash
go run ./.agents/skills/go-exported-api/scripts/goapi diff --base main --internal --markdown
```

Optional config: copy [goapi.json.example](goapi.json.example) to repo-root `.goapi.json`.

## Do not include

Naming reviews, rename proposals, unexported symbols, test helpers.
