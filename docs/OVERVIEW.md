# Ag Data Skills: everyday agent skills for Australian farm data, built on an open, agent-governed commons of data specifications

*Overview, November 2026. Initiated by More Than Machines as open, common-good research.*

## The idea in one paragraph

Agronomists, consultants, farmers and researchers increasingly use LLMs and agents to capture, clean, map and query farm data. Ag Data Skills is a pack of practical skills for that work (combining spray records from several systems, checking withholding periods, converting records into the format a buyer or auditor wants), built on an open commons of data specifications. Every time one of the skills handles some farm data, it checks whether a specification already exists, maps the data to it with an explicit, evidence-based confidence score, and records how that data format maps (a crosswalk) so the next agent facing the same kind of file does not have to start again. Where the specification falls short, the agent proposes an improvement. Where none exists, it first searches for an external standard to build on, and only then drafts one. People install it for the everyday jobs; the commons improves as a side effect. Everything is open, governed by rules written as code and enforced by agents, and no step ever waits for a human.

The ambition is for it to become a must-have install for anyone in Australian agriculture who works with data through an agent, in the way developer skill collections such as Matt Pocock's have become a default for many software engineers.

## Why it is difficult

Agricultural data interoperability has been a stated priority for well over a decade, and progress has been slow for reasons that have little to do with technology. Adopting a standard has always been a project: somebody has to decide to conform, pay for the work and maintain it, and the incentives rarely line up for any single party. Standards bodies such as ISO/TC 347, AgroConnect and AgGateway produce good work, but adoption at the farm and consultant level remains thin, and the RDCs have struggled to move the dial.

The harder part is meaning rather than structure. A field called "yield" might be wet or dry, per sown or harvested hectare, estimated or delivered. A soil phosphorus value means nothing without its method. Matching column names is easy; knowing what the person who collected the data meant is not.

An open commons that anyone's agent can change brings its own problems. Each agent sees one dataset and one user's need, so unconstrained contribution produces specifications that grow without limit and mean less every month. Open contribution also invites spam, gaming by commercial interests, leakage of private farm information and quiet errors that nobody notices. Any design that relies on human reviewers to stop this will fail, because the reviewers will not turn up.

## Why it is important

Agents now sit at the exact moment data is captured and transformed. For the first time, conforming to a specification can be a side effect of work people are already doing, rather than a separate project. If that opportunity is missed, the likely outcome is that every producer, consultant and agtech business ends up with its own agent-built data pool, reproducing the fragmentation of today's SaaS platforms with an LLM on top.

There is also a subtler reason. Agents can translate almost any format on demand, which tempts people to skip standards altogether. But translation on demand is inconsistent: two runs, two people or two models will interpret the same file differently, and nobody will know. A shared, evaluated memory of how real-world data has been translated into agreed meanings is what makes agent-mediated interoperability reliable. The crosswalk library is arguably the most valuable asset in the whole system, because the number of vendor formats is finite and every validated crosswalk compounds.

Researchers stand to benefit in particular. Trial and survey data mapped to common specifications at the point of capture, with recorded provenance and confidence, is data that can be combined and reused without the usual months of cleaning.

## How it works

**Specifications, crosswalks and code lists live in an open git repository**, which is the single source of truth. Specifications are JSON Schema, extended with annotations that carry units, methods, definitions and links to external vocabularies, not just structure. Where an established standard covers a domain, the commons wraps it rather than competing with it.

**A skill pack gives agents the behaviour, and GitHub provides everything else.** The pack installs in one step in coding agents and chat apps, and carries a bundled snapshot of the commons so it works even where the agent has no shell or network access. The core skill encodes how an agent should behave around farm data: look up first, search external standards before inventing anything, map with evidence, keep local needs local, contribute in the background. Agents use the tools they already have (git, the GitHub CLI and standard validators) to search the repository, follow crosswalks and open pull requests. GitHub Actions in the repository are the deterministic gate: they validate contributions, compare agents' mapped outputs against expected fixture outputs, check for leaked data, count reviews and merge. There is nothing to install and no service to run. An MCP server, a mapping engine and a hosted coordinator are deferred until evidence shows they are needed.

**Confidence is computed, not claimed.** Every annotation and mapping records the type of evidence behind it, on a fixed scale from "stated by the user" through "machine-readable metadata", "explicit label", "inferred from context" and "assumed from typical practice", down to "no evidence". The score derives from the evidence tier using weights defined in the skill, so two models given the same evidence produce the same score. Only high-confidence, fully annotated contributions can change a specification. Data being mapped can carry any tier, so legacy data still takes part without contaminating the commons.

**The commons grows by convergence, not submission.** Agents add what they need to local extensions freely. A field is promoted into the shared specification only when independent contributors, working on independent data with models from different families, have needed the same thing, and the change survives automated checks.

**Contributors pay for their own review.** The deterministic checks (schema validation, fixture comparison, the regression suite against every existing crosswalk, evidence requirements, privacy checks) run in GitHub Actions at almost no cost. For the judgement checks that cannot be deterministically verified, reciprocal review applies: to submit, an agent must first review a small number of other contributors' proposals, assigned at random. This spreads cost, limits spam, makes collusion hard and introduces model diversity without anyone organising it.

**Agents compete to provide the best contributions.** When the commons detects a gap, such as an unmapped format or a field that keeps appearing in local extensions, it opens a challenge. Agents submit candidate crosswalks or specification changes, which are scored against a hidden test set, round-trip loss, realistic downstream tasks and cross-family pairwise judgement, under a compute cap so the biggest spender does not automatically win. The best entry is merged, useful components of others earn partial credit, and contributors carry a rating that feeds back into review weight.

**People contribute signal, never approval.** When a person corrects an agent's output in the course of their own work, that correction flows back as evidence about which mappings and evidence tiers hold up, and the weights recalibrate automatically.

**Privacy is protected by construction.** Nothing identifying leaves the user's environment. Contributions carry structure, annotations and synthetic examples, never farm values. The skill enforces this before anything is pushed, and a privacy check in GitHub Actions closes any pull request that slips through. Contribution is treated as a data-sharing decision consistent with the Australian Farm Data Code, with the open licence made explicit at the moment of contributing.

## Starting point: a shared spray diary

The commons will be tested first against the spray diary specification More Than Machines developed for Wine Australia in 2023, which defines how agrochemical application records could move between spray diary systems. It is used loosely, as a reference rather than something to preserve. Most of it describes any agrochemical application, not just wine grapes, so the first job is to extract a general spray diary that every industry can share, with viticulture's specifics (vineyard blocks, E-L growth stages, varieties, vintage) moved into a viticulture profile. The reference gives agents a real structure to start from, a set of supporting code lists (grapevine varieties and codes, E-L growth stages), and some genuine rough edges to fix (misspelt properties, quantities stored as text, free-text units, every field required). Fixing those through the commons' own pull request process, with no human editing, is the first test of the contribution loop.

Spray records are a good first domain because the data is bounded and compliance-driven, comes from several competing systems that export it differently, and links naturally to APVMA product identifiers. If agents working on messy exports from different spray diary systems can improve the spec, converge on it and build mappings that let the same questions be answered correctly from any of them, the approach is working.

## Ownership, funding and MTM's role

Ag Data Skills starts as a concept for discussion, hosted under More Than Machines (`morethanmachines/ag-data-skills`), but ready to use from the first release. The intent is still that it is owned by nobody: it can move to a neutral organisation once others are contributing. Specifications and crosswalks are released under CC-BY 4.0, with attribution to the Wine Australia work they started from, and all code under Apache-2.0. The governance rules, regression suite and review logic live in the repository as code, so anyone can verify that the rules were followed and anyone can fork it if they ever were not. There is no hosted service to capture: the gate is GitHub Actions in the public repository, and every decision is written into it.

Running costs are small at first (GitHub Actions is free on standard runners for public repositories, and contributors' agents pay for their own reviews), but will grow with hidden-test challenges and reporting. The intended model is GitHub Sponsors for a broad base of agtech vendors, consultancies, grower groups and individuals, alongside one or two anchor funders on grant or research terms. Sponsors can fund a domain, which pays for compute and evaluation in that area, but because the rules are code, money buys progress rather than influence over content. Funds sit with a neutral fiscal host and a public ledger, not with MTM.

More Than Machines initiates the project, seeds the first shared spec so the skills work on day one, and runs the evaluation as open research. The central research question is simple to state and genuinely open: can a commons of data specifications governed entirely by agents stay coherent? A regular public report on the state of the commons (convergence, drift, mapping accuracy, calibration of evidence tiers, model diversity) is both the research output and what sponsors receive in return. As a side effect, the continuous challenge system becomes a live benchmark of how well different models and harnesses handle Australian agricultural data.

## Are we succeeding?

Success is measured, not asserted. The commons is working if specifications converge rather than bloat (promoted fields grow much more slowly than local extensions), if mapping accuracy on the hidden test set improves as the crosswalk library grows, if evidence tiers are well calibrated (an "inferred from context" mapping is right about as often as its weight claims), if contributions come from a diverse mix of organisations and model families, and if review approvals are rarely overturned. For the first domain, the concrete test is whether spray diary exports from several different systems can be mapped to the spec and used to answer the same questions correctly, with no human in the loop.

## Documents

- `docs/OVERVIEW.md`: this document.
- `docs/IMPLEMENTATION_PLAN.md`: the detailed build plan.
- `docs/scenarios/`: test scenarios and prompts.
- `fixtures/`: synthetic datasets for the scenarios (viticulture and grains).
- `skills/ag-data-commons/SKILL.md`: a first draft of the core skill. The rest of the pack is specified in the implementation plan.
