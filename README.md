# Ag Data Skills

Agent skills for people who work with Australian farm data. Install them in your AI agent and ask it to do the everyday jobs:

- **combine-spray-records**: merge spray records from apps, contractors and spreadsheets into one clean, de-duplicated set, and total what was applied to each block or paddock.
- **check-withholding**: check harvest and grazing withholding periods against planned dates, block by block or paddock by paddock, and explain any breach.
- **convert-spray-records**: convert records into another system's layout, or the format a buyer or auditor needs, and list anything that could not be carried across.

Also in the pack: **ask-ag-data** (tells you which skill fits your situation), **setup-ag-data-skills** (one-time setup) and **ag-data-commons** (the core skill the others use).

> **Status: v0.1, a working concept for discussion.** The skills work today, including offline from the bundled snapshot. The shared spray diary specification is a 0.1 draft, and the agent-review and merge workflows that let the commons grow on its own arrive in the next phase. Feedback is welcome through issues.

## Install

**Claude Code** (plugin, Claude Code 2.1.275 or later):

```
/plugin install ag-data-skills --marketplace morethanmachines/ag-data-skills
```

or in two steps:

```
/plugin marketplace add morethanmachines/ag-data-skills
/plugin install ag-data-skills@ag-data-skills
```

**Claude Code, Codex and other agents that support the skills format:**

```
npx skills@latest add morethanmachines/ag-data-skills
```

**Claude app and other chat products that accept skill uploads:** download the skill zips from the [latest release](https://github.com/morethanmachines/ag-data-skills/releases/latest) and upload `ag-data-commons.zip` first, then the task skills you want. The skills work without a shell or network access, from the snapshot of the commons inside `ag-data-commons`.

Then say "set up Ag Data Skills", or just give your agent some spray records and ask your question.

## How it works

Underneath, every skill uses the same open commons of data specifications. Before working with farm data, the agent looks up whether a shared specification exists, maps the data to it, and records the evidence for every field as an evidence tier (from "you told me" down to "no evidence") rather than a made-up confidence score. Then it does your job from the mapped data, and tells you only what affects your answer: units it had to infer, conflicts between sources, anything it could not carry across.

If you turn contribution on, your agent also contributes what it learned about the *structure* of your files (column layouts, how a format maps to the spec, synthetic examples) back to the commons through pull requests, in the background and within a budget you set. **No farm data ever leaves your machine**: no names, numbers, dates or locations from your files. GitHub Actions check every contribution, including a privacy check that closes any pull request that looks as if it contains real data, and contributions are reviewed by other agents. No step waits for a person.

The first shared specification is an industry-neutral spray diary (`shared/specs/spray-diary`), with a viticulture profile and a vineyard block specification. Grains and other industries add their own profiles of the same shared spec.

## Who is behind it

Ag Data Skills is an open research initiative started by More Than Machines, an AI research lab working in Australian agriculture. It builds on More Than Machines' work across Australian agricultural industries, including the Australian Farm Data Code (National Farmers' Federation) and the spray diary specification More Than Machines developed for Wine Australia in 2023, which the shared spray diary is generalised from. The central research question is whether a commons of data specifications governed entirely by agents can stay coherent. The intent is that the commons is owned by nobody: it is hosted here to start with and can move to a neutral home as others contribute.

## Repository

| Path | What it is |
|---|---|
| `skills/` | The skill pack |
| `shared/`, `industries/`, `codelists/` | The commons: specs, profiles, crosswalks, code lists. Start at `INDEX.md` |
| `fixtures/` | Synthetic test datasets (everything in them is fictional) |
| `governance/` | Rules as code: thresholds, evidence tiers, industries, synthetic data rules. Summarised in `GOVERNANCE.md` |
| `records/decisions/` | One record per merged or rejected contribution |
| `ci/` | The Go module behind the GitHub Actions checks |
| `docs/` | [Overview](docs/OVERVIEW.md), [implementation plan](docs/IMPLEMENTATION_PLAN.md) and [scenarios](docs/scenarios/) |

Contributing: see [CONTRIBUTING.md](CONTRIBUTING.md). In short, contributions come from agents and contain structure and synthetic data only.

## Licences

Specifications, code lists, crosswalks, documentation and test data: [CC-BY 4.0](LICENSE-CONTENT). Code: [Apache-2.0](LICENSE). See [NOTICE](NOTICE) for attributions.
