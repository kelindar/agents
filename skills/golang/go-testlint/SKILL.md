---
name: go-testlint
description: >-
  Audit and fix Go test naming, file pairing, and table-test structure using
  scripts/testlint.go. Use when writing or reviewing _test.go files, long Test
  names, test hygiene, table tests, fixtures, or foo_test.go without foo.go.
disable-model-invocation: true
---

# Go testlint

Enforce short top-level test names, table-driven subtests, paired test files, and fixtures over inline blobs.

## Conventions

- Top-level `Test*` names: **`Test` + â‰¤3 camelCase words**, â‰¤40 chars. Scenario detail goes in `t.Run` names.
- Prefer **table tests** (`map[string]struct{...}` + `t.Run`) when a file has several related cases.
- Prefer **fixtures** (`testdata/`, `fixtures/`, YAML/JSON files) over large literals in Go.
- Every `foo_test.go` must have **`foo.go`** in the same directory. Every `foo.go` must have **`foo_test.go`**, except source files whose stem already contains `_` such as `foo_bar.go`. Split, merge, or rename â€” never orphan tests.

## Audit

From repo root:

```bash
go run ~/.agents/skills/go-testlint/scripts/testlint.go
go run ~/.agents/skills/go-testlint/scripts/testlint.go -all          # every name
go run ~/.agents/skills/go-testlint/scripts/testlint.go -max-words=2  # stricter
```

Exit `0` = clean. Exit `1` = long names, unpaired `_test.go`, unpaired simple-name `.go`, or orphan test dirs. Hints list files with â‰¥5 top-level `Test*` funcs and no `t.Run`.

Project-local copy (optional): `go run ./tools/testlint` if the repo vendors the script.

## Fix workflow

1. Run audit; fix **unpaired/orphan files** first (merge tests into paired file, or split `foo.go` / rename `foo_test.go`).
2. For each file flagged as table-test candidate, collapse related `Test*` funcs into one short parent + `t.Run` subtests.
3. Rename remaining long top-level names; avoid duplicate parent names in the same package.
4. Re-run audit, then `go test ./...`.

## Naming pattern

| Bad | Good |
|-----|------|
| `TestProgramAutoStartedDelaySchedulesApproveRejectWithoutBlocking` | `TestRun` â†’ `t.Run("delay autostart schedules reject")` |
| `TestCompileRejectsIncompatiblePorts` | `TestCompile` â†’ `t.Run("rejects incompatible ports")` |
| `TestEnsurePipedreamExternalUserPersistsActor` | `TestPipedreamUser` â†’ `t.Run("persists actor")` |

Group by subject (`TestCompile`, `TestRun`, `TestConnect`, `TestNormalize`), not by full sentence.

## Table test shape

```go
func TestCompile(t *testing.T) {
	tests := map[string]struct {
		spec    Spec
		wantErr string
	}{
		"rejects incompatible ports": {...},
		"accepts any port":           {...},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Compile(tc.spec, ctors)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
```

Fixture-driven files: one `TestFixtures` (or `TestRun`) looping YAML/JSON cases is ideal.

## File pairing

| Problem | Fix |
|---------|-----|
| `mcp_auth_test.go`, no `mcp_auth.go` | Merge into `mcp_test.go` or extract `mcp_auth.go` |
| `link_test.go` tests `query.go` + `sync.go` | Split into `query_test.go`, `sync_test.go` |
| `slog_test.go`, only `logging.go` | Merge into `logging_test.go` |

## Agent loop

When asked to clean tests or fix testlint failures:

1. Audit â†’ list violations by file.
2. One file (or one package) per iteration; table-merge before one-off renames.
3. `go test` the touched package after each batch.
4. Re-audit until exit `0`.

Do not weaken limits to pass audit; shorten names and add subtests instead.
