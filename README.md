# Agent skills

This repository is the Git-managed source of truth for my personal Codex skills. Each skill lives under [`skills/`](skills/) and keeps its instructions in a `SKILL.md` file.

## Skill directory

| Area | Skill | Purpose |
| --- | --- | --- |
| Engineering | [`code-review`](skills/code-review/SKILL.md) | Run a deliberately harsh maintainability review focused on weak abstractions, oversized files, and tangled conditions. |
| Engineering | [`debug`](skills/debug/SKILL.md) | Debug difficult bugs and performance regressions with a reproducible test or measurement. |
| Design | [`frontend-design`](skills/design/frontend-design/SKILL.md) | Give new or reshaped interfaces a deliberate visual direction, including typography and layout choices. |
| Design | [`interface-design`](skills/design/interface-design/SKILL.md) | Design or audit product interfaces such as dashboards, admin tools, settings, and data-heavy applications. |
| Go | [`go-code`](skills/golang/go-code/SKILL.md) | Apply distilled contract, control-flow, and test rules whenever writing Go code. |
| Go | [`go-contract`](skills/golang/go-contract/SKILL.md) | Design and review package contracts before changing caller-visible Go APIs or behavior. |
| Go | [`go-rename`](skills/golang/go-rename/SKILL.md) | Plan and apply stutter-free package and API renames in verified waves. |
| Go | [`go-simplify`](skills/golang/go-simplify/SKILL.md) | Reduce complexity in Go code by removing dead concepts, duplication, and unnecessary APIs while preserving behavior. |
| Go | [`go-switch`](skills/golang/go-switch/SKILL.md) | Find branch-heavy Go functions and replace suitable `if` chains with clearer `switch` statements. |
| Go | [`go-testlint`](skills/golang/go-testlint/SKILL.md) | Audit and fix Go test naming, source-file pairing, table structure, and fixture use. |
| Planning | [`grilling`](skills/grilling/SKILL.md) | Stress-test a plan, decision, or idea through a focused interview. |
| Planning | [`to-domain`](skills/to-domain/SKILL.md) | Turn business-domain understanding into `specs/DOMAIN.yaml`. |
| Planning | [`to-spec`](skills/to-spec/SKILL.md) | Turn the current conversation into a standalone draft specification under `specs/todo`. |
| Planning | [`to-tickets`](skills/to-tickets/SKILL.md) | Convert an approved specification and plan into the smallest useful set of delivery tickets. |
| Writing | [`write-for-agents`](skills/writing/write-for-agents/SKILL.md) | Write instructions for coding agents, including skills, `AGENTS.md`, and `CLAUDE.md`. |
| Writing | [`write-for-humans`](skills/writing/write-for-humans/SKILL.md) | Remove common AI writing tells and make prose sound specific and human. |

`go-code` runs automatically for Go implementation, fixes, refactors, and tests. It applies `go-contract` to caller-visible changes, keeps performance visible, uses Testify in every test, and runs its bundled test linter before focused and full tests. Its adversarial eval set checks those rules under conflicting prompts and deadline pressure.

Folders such as `agents/`, `scripts/`, `evals/`, and `assets/` belong to the skill beside them. They are support files, not separate skills.

`grilling`, `to-spec`, `to-tickets`, `go-rename`, `go-switch`, and `go-testlint` require explicit invocation. Other skills follow their configured runtime defaults.
