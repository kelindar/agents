---
name: go-rename
description: >-
  Proposes stutter-free renames (package-context names like light.Point,
  assets.Source) and applies them in safe waves with go test verification.
  Use when renaming packages or exported types, removing name stutter, or doing
  a systematic API naming pass. For API inventory only, use go-exported-api.
disable-model-invocation: true
---

# Go exported API rename pass

Systematic workflow for review → phased renames → verify. Target style: **package context carries the domain** (`light.Point`, `light.Activate`, `assets.Source`, `scene.Preview`), not repeated prefixes (`PointLight`, `ActivatePointLights`, `LightSource`).

## When to use

- User asks for a naming pass or stutter cleanup
- After splitting/renaming packages (e.g. `structural` → `geometry`)
- Before a breaking public API cleanup across `internal/`

Not for API inventory — use `.agents/skills/go-exported-api` (`goapi write`, `goapi eval`, `goapi diff`).

## Principles

1. **Package context is enough** — inside `assets`, `Source` not `LightSource`; inside `light`, `Point` not `PointLight`.
2. **Verb-first package functions** — `Activate`, `Validate`, `LoadSources` not `ActivatePointLights`, `ValidateLightSource`.
3. **Stable wire format** — keep `json:"..."` tags when renaming Go fields (`Sources` with `json:"lights"`).
4. **Resolve homonyms by role** — e.g. `render.View` embeds `camera.Camera`; `edit/view.BuildPreview` vs `scene.Preview`.
5. **Do not rename generated code** (`*_gen.go`, msgpack helpers) unless regenerating from templates.
6. **Leave domain jargon in comments** — e.g. frame `structural []uint32` lane buffer ≠ subpackage name.

## Workflow

If no current API inventory exists, run the `go-exported-api` skill first.

Copy and track:

```
- [ ] 1. Naming review (P0–P3) — append to types.md or types_naming.md
- [ ] 2. Plan waves (A–D); get scope sign-off if huge
- [ ] 3. Apply renames (ordered replacements)
- [ ] 4. go build ./... && go test ./...
- [ ] 5. Fix homonym/collateral damage; gofmt
- [ ] 6. Update README / wasm / editor call sites
- [ ] 7. Refresh types.md via go-exported-api
```

### 1. Naming review

Append to the same `types.md` (or `types_naming.md` merged in):

| Priority | Meaning | Examples |
|----------|---------|----------|
| **P0** | Broken or misleading names | `NormalizePointsource` → `NormalizeSource` |
| **P1** | Clear stutter / homonyms | `LightSource` → `Source`, `Lights()` → `Points()` |
| **P2** | Consistency polish | `GetHexPalette` → `HexPalette`, `HostTileGrids` → `HostGrids` |
| **P3** | Keep as-is | `PackSurface`, generated `Create1`, enum `MaterialAsphalt` |

Include **homonym table** and **rename waves** (see [reference.md](reference.md)).

### 2. Rename waves (default order)

| Wave | Scope | Examples |
|------|-------|----------|
| **A** | Prefab/data package + parse | `Source`, `NormalizeSource`, `parse.LoadSources` |
| **B** | Runtime scene types | `scene.New`, `Points()`, `Enrich`, `Preview` |
| **C** | Cross-package disambiguation | `render.View`, `edit/view.BuildPreview` |
| **D** | Constants, debug helpers, display enums | `DisplayBottom`, `ApplyGeometryCount` |

### 3. Apply renames safely

**Order matters.** Longest / most specific tokens first. Never blanket-replace short substrings before long ones.

Suggested sequence:

1. P0 corrupted names (`NormalizePointsource`, func/type collisions like `parse.LightAnimation` → `parse.ParseAnimation`)
2. `ValidateLight*` / `ParseLight*` / `NormalizeLight*` before bare `Light*`
3. `LightSource` → `Source` (not before step 2)
4. `LightActivation` / `LightAnimation` → `Activation` / `Animation`
5. Scene/runtime: `EnrichLights`, `ToggleLightAt` before `ToggleLight`
6. Homonym-specific renames **before** global `Camera` or `TilePreview` sweeps
7. Field renames: `Object.Lights` → `Object.Sources` with **regex** `\bLights\b` only where it is the field/method — **not** `.Lights()` → `.Sources()` (that should become `.Points()` for resolved runtime lights)
8. `scene.Open` / package-local `Open(` → `New` only in the target package (not `os.Open`)
9. `go build ./...` then `go test ./...`

**Protect packages that must keep a name:** e.g. only `render` wrapper becomes `View`; `scene/camera.Camera` stays `Camera`.

### 4. Verify

```bash
go build ./...
go test ./...
gofmt -w <changed packages>
```

Grep for stale symbols and collateral:

```bash
rg 'LightSource|PointLight|NormalizePoint|EnrichLights|TilePreview|scene\.Open|NewCamera' --glob '*.go'
```

Update `README.md`, WASM/editor JS only if they call renamed Go exports.

### 5. Shader / wire collateral

- Kage/uniform names must match Go `Uniforms` map keys.
- JSON wire structs may keep `Lights` field name while Go uses `Sources`.

## xrender conventions (this repo)

| Area | Pattern |
|------|---------|
| Prefab lights | `assets.Source`, `Object.Sources`, `parse.LoadSources`, `parse.ValidateSource` |
| Runtime lights | `light.Point`, `light.Activate`, `frame.Points`, `Scene.Points()` |
| Scene lifecycle | `scene.New`, `scene.NewPreview`, `scene.Preview` |
| Viewport | `scene/camera.Camera`, `render.View`, `render.NewView` |
| Geometry subpackage | `geometry.Rasterize`, `geometry.TopAt` (import as `geom` in `scene`) |
| Composition | `assets.Mode`, `CompositionCached` consts unchanged |

## Additional resources

- Pitfalls, homonym patterns, and wave templates: [reference.md](reference.md)
- Historical snapshot (may drift): repo root `types.md`
