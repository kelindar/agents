# Go API rename — reference

## Homonym resolution patterns

| Conflict | Resolution |
|----------|------------|
| `scene.Preview` struct vs `edit/view` builder func | Struct: `scene.Preview`; func: `view.BuildPreview` |
| `render.Camera` vs `scene/camera.Camera` | Wrapper: `render.View` embedding `camera.Camera` |
| Prefab `assets.Source` vs runtime `light.Point` | `Object.Sources` / `SetSources`; scene method `Points()` not `Lights()` |
| JSON `"lights"` vs Go field | `Sources` + `` `json:"lights,omitempty"` `` |
| `parse` wire struct | Keep `Lights []sourceWire` with `json:"lights"`; map to `o.Sources` |

## Bulk-replace pitfalls (learned)

| Mistake | Effect | Fix |
|---------|--------|-----|
| `CompositionMode` → `Mode` globally | Broke unrelated `Camera` if combined with other replaces | Scope replaces per package; revert `scene/camera.Camera` |
| `.Lights\b` before `.Lights()` → `.Points()` | `frame.Lights()` became `frame.Sources()` | Use `.Lights()` → `.Points()` first, or only rename struct fields `Lights` → `Sources` |
| `lightSources` → `sources` | `LoadSources` → `Sources` (broken call) | Replace `assetparse.Sources` → `assetparse.LoadSources` |
| `TilePreview` global replace | `BuildPreview` → `Preview` in tests | Rename func to `BuildPreview` **before** struct `TilePreview` → `Preview` |
| `func Open(` → `func New(` only on `scene.Open(` | Left `Open(` inside `scene` package tests | Also replace package-local `\bOpen(` → `New(` |
| `ToggleLight` before `ToggleLightAt` | Corrupted `ToggleAt` | Rename `ToggleLightAt` first |

## Rename wave template (paste into types.md)

```markdown
## P0 — fix now

| Current | Proposed | Package | Rationale |
|---------|----------|---------|-----------|

## P1 — high value

### `internal/<pkg>`

| Current | Proposed | Notes |
|---------|----------|-------|

## P2 — medium

| Current | Proposed | Package | Notes |
|---------|----------|---------|-------|

## P3 — keep as-is

- ...

## Homonyms

| Symbol A | Symbol B | Recommendation |
|----------|----------|----------------|

## Suggested waves

**Wave A** — ...
**Wave B** — ...
**Wave C** — ...
**Wave D** — ...
```

## Stutter heuristics (Go)

Drop repeated package/domain token from exported names when:

- The package import already disambiguates (`light.Point`, `geometry.Field`)
- Only one primary type exists in the package (`assets.Source`)
- Method receiver supplies context (`(*Scene).Enrich` not `EnrichLights`)

Keep the prefix when:

- Name is a widely recognized enum constant (`MaterialAsphalt`, `SurfaceGround`)
- Shortening loses meaning across packages (`PackSurface`, `PackCaster`)
- Symbol is generated or part of an external contract

## Optional AST inventory snippet

Throwaway `go:build ignore` or `tools/` script: walk `**/*.go`, skip `*_test.go` and `main`, emit sorted exported decls. Delete the tool after the pass unless the team wants to keep it.

## Verification checklist

- [ ] `go test ./...` green
- [ ] No stale grep hits for old names (except intentional JSON/wire/comments)
- [ ] README / skill docs updated
- [ ] Shader uniform keys match Go maps
- [ ] WASM/editor bridges use new export names
