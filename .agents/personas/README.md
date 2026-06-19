# Agent Personas

This directory contains reusable reviewer personas for agent workflows.

They are written to be usable with Codex or any other AI agent that can read a
Markdown prompt file and follow repository instructions from `AGENTS.md`.

Unlike the procedural prompts in `.agents/commands/`, a persona defines a *role*
(what to look for, how to flag findings, what output to produce) rather than a
step-by-step task.

## Usage

Give your agent one of these files along with the code to review.

Example:

```text
Use .agents/personas/go-reviewer.md to review the staged Go changes.
```

```text
Use .agents/personas/go-reviewer.md to review internal/server/evaluation.
```

## Available personas

- `go-reviewer.md` — senior Go reviewer; expands the "Go conventions" and
  "Testing" sections of `AGENTS.md` into a full review checklist with examples.
- `react-reviewer.md` — senior frontend reviewer for the UI; expands the
  "Conventions" and "Testing" sections of `ui/AGENTS.md` into a full review
  checklist with examples.

## Conventions

- Personas defer mechanical formatting and lint findings to `goimports` and
  `golangci-lint`; they flag only what the linter won't catch.
- Findings should cite `file:line` and separate blocking issues from
  suggestions.
