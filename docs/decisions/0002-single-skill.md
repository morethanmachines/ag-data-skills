# 0002: One skill instead of a pack

*Date: 2026-11-24.*

Version 0.1 shipped six skills, as section 0.1 of the implementation plan then specified: three task skills (`combine-spray-records`, `check-withholding`, `convert-spray-records`), a setup skill, a router and the core `ag-data-commons` skill. More Than Machines decided to replace them with the one core skill.

**Why.** The idea is that common specifications emerge from agents doing real jobs with messy data, and the core skill is what tells agents how to take part. The task skills hid that idea behind a set of tools, and they added little: a capable model already does the arithmetic of combining records or counting days to harvest. What it lacks is the behaviour of mapping to a shared spec with evidence and contributing back, which is all in the core skill.

**What changed.**

- The core skill's description names the everyday jobs, so agents still use it whenever farm data is involved without a task skill to trigger it.
- The useful job knowledge (de-duplicating and comparing sources, the withholding rules, converting via the spec) moved into `skills/ag-data-commons/references/jobs.md`.
- Setup became a first-run step in the core skill: ask once about industries and contribution, and write `~/.config/ag-data-skills/settings.yaml`.
- The router was removed, since there is nothing left to route between.
- The skill now opens with how the commons grows from agents' work (sightings, crosswalks, extension promotion, evidence-backed fixes), not with looking up a fixed spec.
- The plan (sections 0, 0.1, Phase 0 and 16), the overview, the README, `CONTRIBUTING.md`, the plugin manifests and the release notes were updated to match. The repository and plugin keep the name Ag Data Skills.

The Phase 0 acceptance test was run again with only the single skill installed; see `records/reports/phase-0-acceptance.md`.
