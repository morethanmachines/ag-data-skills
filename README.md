# Ag Data Skills

Agent skills for people who work with Australian farm data, built on an open commons of data specifications that improves every time the skills are used.

> **Status: concept for discussion.** This repository sets out the design, the first test scenarios and synthetic datasets, and a draft of the core skill. The working skill pack, the shared spray diary specification and the GitHub Actions that govern contributions are being built now (Phase 0 in the implementation plan). Feedback is welcome through issues.

## What it will do

Agronomists, consultants, farmers and researchers increasingly use AI agents to clean, combine, convert and query farm data. Ag Data Skills gives those agents practical skills for everyday jobs:

- **combine-spray-records**: merge spray records from apps, contractors and spreadsheets into one clean, de-duplicated set.
- **check-withholding**: check harvest and grazing withholding periods against planned dates, block by block or paddock by paddock.
- **convert-spray-records**: convert records into another system's layout, or the format a buyer or auditor needs.

Underneath, every skill uses the same commons. Before working with a piece of farm data, the agent checks whether a shared specification exists, maps the data to it with evidence-based confidence, and records how that file format maps, so the next agent with the same kind of file does not have to work it out again. Where a specification falls short, the agent proposes an improvement. Contributions are checked and merged by GitHub Actions and reviewed by other agents, with no step waiting on a person. No farm data ever leaves the user's machine: only structure and synthetic examples are contributed.

The first shared specification is a general, industry-neutral spray diary, developed from the spray diary specification More Than Machines produced for Wine Australia in 2023, with viticulture and grains as the first industry profiles.

## Read more

- [Overview](docs/OVERVIEW.md): the idea, why it is difficult, why it is important, and how success will be measured.
- [Implementation plan](docs/IMPLEMENTATION_PLAN.md): the build specification.
- [Scenarios](docs/scenarios/): test scenarios for viticulture and grains.
- [Fixtures](fixtures/): synthetic datasets. Everything in them is fictional.

## Installing (once Phase 0 is released)

```
npx skills@latest add morethanmachines/ag-data-skills
```

A Claude Code plugin and downloadable skill zips for chat apps will follow with the first release.

## Who is behind it

Ag Data Skills is an open research initiative started by More Than Machines, an AI research lab working in Australian agriculture. It builds on More Than Machines' work across Australian agricultural industries, including the Australian Farm Data Code (National Farmers' Federation) and the Wine Australia spray diary specification. The intent is that the commons is owned by nobody: it is hosted here to start with and can move to a neutral home as others contribute.

## Licences

Specifications, code lists, crosswalks, documentation and test data: [CC-BY 4.0](LICENSE-CONTENT). Code: [Apache-2.0](LICENSE). See [NOTICE](NOTICE) for attributions.
