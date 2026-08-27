---
name: debug
description: Debug hard bugs and performance regressions with a reproducible test or measurement.
---

# Debug

Use this workflow for hard bugs and performance regressions. Skip a phase only when you can explain why.

## Read relevant specs

Locate the project's spec directory. Prefer a path named by repository instructions; otherwise search for an `AGENTS.md` inside a `spec` or `specs` directory. Read it as the index, then open only the specs tied to the reported behavior. If no spec index exists, continue with the code.

## Protect secrets

This workflow shows commands, output, and captured artifacts. Redact every secret first by replacing it with `<REDACTED>`. Keep credentials in environment variables instead of command output. Captured artifacts may contain authentication headers, so show only the lines needed for diagnosis.

If redaction removes the evidence needed to diagnose the problem, say so and ask for a safer artifact.

## Phase 1: Build a feedback loop

Start with one pass/fail command that exercises the reported bug and can fail on its exact symptom. Without that command, code-based theories are speculation.

Build the loop in roughly this order:

1. A failing unit, integration, or end-to-end test at the seam that reaches the bug.
2. A curl or HTTP script against a running development server.
3. A CLI invocation with fixture input and a comparison against known-good output.
4. A headless browser script using Playwright or Puppeteer that checks the DOM, console, or network.
5. A replay of a captured trace, request, payload, or event log.
6. A throwaway harness that exercises the path with a minimal set of dependencies.
7. A property or fuzz loop for incorrect output that appears only for some inputs.
8. A bisection harness that boots two known states and checks each candidate automatically.
9. A differential loop that compares old and new versions or configurations.
10. A HITL helper when a human must perform the final interaction. Use `scripts/hitl-loop.template.ps1` on Windows and `scripts/hitl-loop.template.sh` on Unix-like systems.

Once the loop exists, tighten it. Cache setup, narrow the test, assert on the reported symptom, and control time, randomness, filesystem state, and network access where possible. A deterministic loop that runs in seconds is the target.

For non-deterministic bugs, increase the reproduction rate. Repeat the trigger, run it in parallel, add stress, or narrow the timing window. A one-percent failure rate is not a useful debugging loop.

If you cannot build a loop, stop and say what you tried. Ask for one of these:

- Access to the environment that reproduces the problem.
- A redacted log dump, HAR file, core dump, or screen recording with timestamps.
- Permission to add temporary production instrumentation.

Do not continue with unsupported hypotheses.

The phase is complete when you have already run one command that is:

- Red-capable. It drives the real bug path and checks the user's exact symptom.
- Deterministic, or repeatable at a useful rate for a flaky bug.
- Fast enough to run repeatedly.
- Agent-runnable without a human, except through the HITL script.

## Phase 2: Reproduce and minimize

Run the loop and confirm that it fails for the reason the user reported. Capture the exact error, wrong output, or timing result.

Then minimize the case. Remove one input, caller, configuration value, data item, or step at a time. Rerun the loop after each change. Keep only elements required for the failure. Removing any remaining element should make the loop pass.

Do not proceed until the failure is reproduced and minimized.

## Phase 3: Form hypotheses

Write 3 to 5 ranked hypotheses before testing any of them. Each one must make a prediction.

Use this form:

> If `<X>` is the cause, changing `<Y>` will make the bug disappear, or changing `<Z>` will make it worse.

Show the ranked list to the user before testing. Their context may change the order. If they are unavailable, continue with your ranking.

The phase is complete when every hypothesis has a rank and a testable prediction, and the list has been shown or recorded for the user.

## Phase 4: Instrument

Each probe should test one prediction at a time.

Prefer these tools in order:

1. A debugger or REPL inspection.
2. Targeted logs at the boundaries that distinguish the hypotheses.
3. A narrow trace or measurement that isolates the suspected change.

Tag temporary logs with a unique prefix such as `[DEBUG-a4f2]`. Search for the prefix during cleanup and remove every tagged log.

For performance regressions, establish a baseline with a timing harness, profiler, query plan, or benchmark before changing code. Then bisect or compare one variable at a time.

The phase is complete when each probe has a result tied to a hypothesis. For performance work, record the baseline and each comparison.

## Phase 5: Fix and add regression coverage

Write the regression test before the fix when a correct test seam exists. The seam must exercise the real bug pattern at the relevant call site. A shallow test that cannot reproduce the triggering chain gives false confidence.

If no correct seam exists, record that as an architecture finding.

When a seam exists:

1. Turn the minimized case into a failing test at that seam.
2. Run it and confirm that it fails.
3. Apply the fix.
4. Run the test and confirm that it passes.
5. Rerun the original, unminimized scenario through the Phase 1 loop.

## Phase 6: Clean up

Before declaring the work complete:

- Rerun the original loop.
- Rerun the regression test, or document why no correct seam exists.
- Search for and remove all temporary debug instrumentation.
- Delete throwaway prototypes or move them to a clearly marked debug location.
- State the confirmed cause in the commit or pull request message.
