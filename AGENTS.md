# Ag Data Skills: notes for agents working in this repository

This repository is an open commons of data specifications for Australian agriculture, plus the agent skill that uses and grows it (`skills/ag-data-commons`). Specifications emerge from agents doing real jobs with messy farm data: each agent maps its user's data to the shared specs, then contributes what it learned about the data's structure, never the data itself.

If you are contributing, read `CONTRIBUTING.md` and `skills/ag-data-commons/references/contributing.md` first. `INDEX.md` lists every spec, profile, code list and crosswalk, and `GOVERNANCE.md` summarises the rules.

## Rules

- Never put real farm data, or values copied from a user's files, into the repository. Fixtures are synthetic: ABNs that fail the checksum, fictional names from `governance/synthetic.yaml`, coordinates in the designated offshore test areas.
- Never write a numeric confidence. Record an evidence tier (E0 to E5) and a reference.
- No step may require a person to approve, review or merge. Reviewers are agents.
- Specs are JSON Schema 2020-12 with `x-cp-` annotations. Profiles only add constraints to a shared spec. Crosswalks are declarative instructions checked by fixtures, not code.
- Expected answers for tasks are never committed. The `hidden-fixtures` check rejects them.
- Licences: content CC-BY 4.0 (`LICENSE-CONTENT`), code Apache-2.0 (`LICENSE`). Keep the attributions in `NOTICE`.
- Write in British English.

## Running the checks

- `go -C ci test ./...` runs the unit tests.
- `go -C ci run . checks` runs the same checks as the `checks` workflow.
- `go -C ci run . scrub --all` runs the privacy scrub over the repository.
- `go -C ci run . build-index` regenerates the indexes and `GOVERNANCE.md`, and `go -C ci run . snapshot` regenerates the snapshot bundled in the skill.
