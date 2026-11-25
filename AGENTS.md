# Ag Data Skills: notes for Claude Code

## What this is

Ag Data Skills is an agent skill (`skills/ag-data-commons`) that lets common data specifications for Australian agriculture emerge from agents doing real jobs with messy farm data (combining spray records, checking withholding periods, converting records between formats). Every time the skill handles farm data it looks up a shared spec, maps the data with evidence-based confidence, and contributes structure (never farm data) back through pull requests that GitHub Actions check and agents review. No step ever waits for a human.

It is being started by More Than Machines (Andrew Skinner), an AI research lab working in Australian agriculture, and will be hosted at `morethanmachines/ag-data-skills` as a public repository. It is a concept for discussion, but it must be ready to use when someone wants to try it.

## Read first

`docs/` is local only: it is git-ignored and must not be committed or pushed, because it holds plans, decisions and questions that are not public.

1. `docs/OVERVIEW.md`: intent, design and how success is measured.
2. `docs/IMPLEMENTATION_PLAN.md`: the build specification. This is the source of truth for what to build. Where it says MUST, it is a hard requirement.
3. `docs/scenarios/`: Scenario 01 (viticulture) and Scenario 02 (grains).
4. `skills/ag-data-commons/SKILL.md`: the skill.
5. `fixtures/`: two synthetic datasets (agent-visible exports and reference files only).

## Current state

- Public repository: `morethanmachines/ag-data-skills`. Phase 0 of the plan is built: commons layout, governance, the shared spray diary 0.1.0 draft, the viticulture profile and vineyard block draft, code lists, fixtures and tasks, the `ci/` Go module, the `checks`, `scrub`, `build-index` and `release` workflows, and the skill. Since v0.2 there is one skill, `ag-data-commons`, not a pack (`docs/decisions/0002-single-skill.md`). Decisions taken while building it are in `docs/decisions/`; deferred work is in GitHub issues.
- `private/` holds the ground truth and generators for both datasets. It is git-ignored and MUST never be committed or pushed to the public repository, because agents under test must not be able to read the answers. A local pre-commit hook and the `hidden-fixtures` check both guard against it. In the plan these belong in a separate private repository (`morethanmachines/ag-data-skills-hidden`); create that when the hidden-fixture workflows need it.
- The generators write `exports/`, `reference/` and `ground_truth/` next to themselves. The public `fixtures/` copies were produced by them. To run a `check_exports.py`, run its generator in a scratch copy first, or point it at the matching public `exports/` and `reference/`.
- The reference spec the shared spray diary is generalised from is `https://github.com/morethanmachines/project-harmony` (spray diary JSON Schema, block format, grapevine variety list, E-L growth stages). Use it loosely, as described in section 13 of the plan. Do not call this project, or anything in it, "Project Harmony". Its licence is unconfirmed, so nothing is copied from it yet.
- Run the checks locally with `go -C ci test ./...`, `go -C ci run . checks` and `go -C ci run . scrub --all`. Regenerate indexes with `go -C ci run . build-index` and the snapshot with `go -C ci run . snapshot`.

## Phase 0 build order (done)

Phase 0 of the plan ("Foundations, ready to use") was built in this order:

1. `git init -b main`, first commit of the current files (not `private/`), and create the public repository with `gh repo create morethanmachines/ag-data-skills --public --source . --push`. Confirm with Andrew before creating it.
2. Build the commons layout (plan section 2), governance files, `governance/industries.yaml`.
3. Seed the shared spray diary 0.1.0 draft, the viticulture profile, the vineyard block 0.1.0 draft and the code lists (plan section 13), recording bootstrap decision records. Keep the original reference schema under `shared/specs/spray-diary/reference/` for provenance. Log the open questions in `KNOWN_ISSUES.md`.
4. Move the datasets into `fixtures/` and their tasks into `industries/<industry>/tasks/` as the plan describes.
5. The `ci/` Go module with unit tests, and the `checks`, `scrub`, `build-index` and `release` workflows.
6. The skill pack v0.1 (plan section 0.1), the Claude Code plugin manifest, and the README install instructions.
7. Meet the Phase 0 acceptance criteria, including: a fresh agent with the pack installed and no network access can map the spray-diary-01 exports from the bundled snapshot and answer the three tasks.

Work in small commits and push as you go. Open issues for anything you defer.

## Rules

- No step in the system may require a human to approve, review or merge. All reviewers are agents.
- Never put real farm data, or values copied from a user's files, into the repository. Fixtures are synthetic: ABNs that fail the checksum, fictional names, coordinates in the designated offshore test areas.
- Agents never produce numeric confidence; they record an evidence tier (E0 to E5) and a reference.
- Specs are JSON Schema 2020-12 with `x-cp-` annotations. Crosswalks are declarative instructions checked by fixtures, not code.
- Licences: content CC-BY 4.0 (`LICENSE-CONTENT`), code Apache-2.0 (`LICENSE`). Keep `NOTICE` attributions to Wine Australia and More Than Machines.
- Stack: Go for `ci/` (`github.com/santhosh-tekuri/jsonschema/v6` for validation, `gopkg.in/yaml.v3`). No AWS, GCP or Azure dependencies.
- Writing: British English throughout, no em dashes, no fabricated references or citations. More Than Machines is an AI research lab working in agriculture, never a consultancy.
