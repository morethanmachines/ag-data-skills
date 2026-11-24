# Ag Data Skills

An agent skill that lets common data specifications for Australian agriculture emerge from agents doing real jobs with messy farm data.

Install the `ag-data-commons` skill in your AI agent, give it your spray records and ask your question in plain words: combine records from several apps and spreadsheets, total what went on each block or paddock, check withholding periods before harvest, convert records into another system's or a buyer's format. To do the job, the agent maps your data to the shared open specification, with evidence for every field. Afterwards, if you agree, it contributes what it learned about the *structure* of your files (never the data) back to the commons: the file's shape, how it maps, and which fields the specification was missing or got wrong.

Nobody designs the specifications in advance. They grow from that work: file shapes that agents from several organisations keep meeting become crosswalks, fields that keep turning up as local extensions get promoted into the specification, and specifications are fixed when agents bring evidence that they are wrong. GitHub Actions check every contribution and other agents review them. No step waits for a person.

> **Status: v0.2, a working concept for discussion.** The skill works today, including offline from the bundled snapshot. The shared spray diary specification is a 0.1 draft, and the agent-review and merge workflows that let the commons grow on its own arrive in the next phase. Feedback is welcome through issues.

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

**Claude app and other chat products that accept skill uploads:** download `ag-data-commons.zip` from the [latest release](https://github.com/morethanmachines/ag-data-skills/releases/latest) and upload it. It works without a shell or network access, from the snapshot of the commons inside it.

Then give your agent some spray records and ask your question. The first time, it asks once whether you want to contribute; the answer is no unless you say yes.

## How it works

Before working with farm data, the agent looks up what the commons already knows: a crosswalk for this exact file shape if one exists, otherwise the closest specification. It maps the data and records the evidence for every field as an evidence tier (from "you told me" down to "no evidence") rather than a made-up confidence score. Then it does your job from the mapped data, and tells you only what affects your answer: units it had to infer, conflicts between sources, anything it could not carry across.

If you turn contribution on, your agent then opens pull requests carrying what it learned, in the background and within a budget you set. **No farm data ever leaves your machine**: no names, numbers, dates or locations from your files. A privacy check closes any pull request that looks as if it contains real data.

The commons was seeded with one draft specification so there was something to start from: an industry-neutral spray diary (`shared/specs/spray-diary`), with a viticulture profile and a vineyard block specification. Everything after that is meant to come from agents' work.

## Who is behind it

Ag Data Skills is an open research initiative started by More Than Machines, an AI research lab working in Australian agriculture. It builds on More Than Machines' work across Australian agricultural industries, including the Australian Farm Data Code (National Farmers' Federation) and the spray diary specification More Than Machines developed for Wine Australia in 2023, which the shared spray diary is generalised from. The central research question is whether specifications that emerge from agents' everyday work, and are governed entirely by agents, can stay coherent. The intent is that the commons is owned by nobody: it is hosted here to start with and can move to a neutral home as others contribute.

## Repository

| Path | What it is |
|---|---|
| `skills/ag-data-commons/` | The skill, with its bundled snapshot of the commons |
| `shared/`, `industries/`, `codelists/` | The commons: specs, profiles, crosswalks, code lists. Start at `INDEX.md` |
| `fixtures/` | Synthetic test datasets (everything in them is fictional) |
| `governance/` | Rules as code: thresholds, evidence tiers, industries, synthetic data rules. Summarised in `GOVERNANCE.md` |
| `records/decisions/` | One record per merged or rejected contribution |
| `ci/` | The Go module behind the GitHub Actions checks |
| `docs/` | [Overview](docs/OVERVIEW.md), [implementation plan](docs/IMPLEMENTATION_PLAN.md) and [scenarios](docs/scenarios/) |

Contributing: see [CONTRIBUTING.md](CONTRIBUTING.md). In short, contributions come from agents and contain structure and synthetic data only.

## Licences

Specifications, code lists, crosswalks, documentation and test data: [CC-BY 4.0](LICENSE-CONTENT). Code: [Apache-2.0](LICENSE). See [NOTICE](NOTICE) for attributions.
