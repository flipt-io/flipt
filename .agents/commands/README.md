# Agent Commands

This directory contains reusable prompts for release-related agent workflows.

They are written to be usable with Codex or any other AI agent that can read a
Markdown prompt file and follow repository instructions from `AGENTS.md`.

## Usage

Give your agent one of these files and provide the required inputs explicitly.

Example:

```text
Use .agents/commands/changelog.md for version 2.5.0.
```

```text
Use .agents/commands/release.md for version 2.5.0 and follow its confirmation checkpoints.
```

## Conventions

- Versions are semver strings without a `v` prefix unless the prompt says
  otherwise.
- These prompts are procedural and include explicit stop points where the agent
  should wait for user confirmation.
