# Ag Data Skills: implementation plan

*Audience: a coding agent building the system. Read `OVERVIEW.md` first for intent. This plan is the build specification. Where it says MUST, treat it as a hard requirement; where it says SHOULD, deviate only with a recorded reason in `docs/decisions/`.*

## 0. Shape of v1

Ag Data Skills is a pack of practical agent skills for people who work with Australian farm data, built on an open commons of data specifications that grows as the skills are used. People install it for the everyday jobs. The commons is what makes those jobs work, and it improves as a side effect.

Version 1 is one public GitHub repository, `morethanmachines/ag-data-skills`, containing three things:

1. **The skill pack** (`skills/`): a few task skills for everyday jobs, a core skill that handles spec lookup, mapping, evidence and contribution, a one-time setup skill and a router skill (section 0.1).
2. **The commons**: specs, profiles, crosswalks, code lists, fixtures, tasks and governance rules.
3. **GitHub Actions workflows** that act as the deterministic gate: they validate contributions, compare agent outputs with expected fixture outputs, check for leaked data, count reciprocal reviews, merge, write decision records, and build the snapshot bundled into the skills (section 4.5).

There is no CLI to install, no MCP server and no hosted coordinator in v1. The agent does all the judgement work (finding, mapping, drafting, reviewing). The Actions do the parts that must give the same answer every time. Anything heavier is deferred until evidence says it is needed (section 12).

The repository is hosted under More Than Machines to start with, as a working concept for discussion that is ready to use when someone wants to try it. It can be transferred to a neutral organisation later; GitHub keeps old links working after a transfer.

### 0.1 The skill pack

The pack copies what has made developer skill collections such as `mattpocock/skills` spread: real working tools rather than a framework, a one-command install, a setup step and a router.

| Skill | Invoked by | What it does |
|---|---|---|
| `setup-ag-data-skills` | User, once | Asks which industries the user works in, whether to contribute to the commons (explaining in two sentences what that means), and the contribution budget. Writes a small settings file the other skills read |
| `ask-ag-data` | User | Router: describes the user's situation back and points to the right skill |
| `ag-data-commons` | Model | Core skill used by all the others: find the spec and crosswalk, fingerprint the source, map with evidence tiers, validate, contribute sightings, crosswalks and fixes in the background |
| `combine-spray-records` | User or model | Merges spray records from several sources (app exports, contractor jobs, spreadsheets) into one clean, de-duplicated set mapped to the shared spray diary spec, and reports conflicts and gaps |
| `check-withholding` | User or model | Checks harvest and grazing withholding periods against planned dates, per block or paddock, and explains any breach |
| `convert-spray-records` | User or model | Converts spray records into another system's layout or a buyer's or auditor's format, via the spec, and reports anything that cannot be carried across |

Every task skill calls `ag-data-commons` first, so every everyday job does the spec work. New task skills are added for new domains as the commons grows (for example soil test interpretation).

**Install routes**, all documented in the README:
- `npx skills@latest add morethanmachines/ag-data-skills` for Claude Code, Codex and other agents that support the skills format.
- A Claude Code plugin manifest (`.claude-plugin/`) so it installs with `/plugin install`, with a later aim of an official marketplace listing.
- A release asset per skill (zip) for uploading to the Claude app and other chat products that accept skill uploads.

**Working without a shell.** Agronomists, consultants and farmers mostly use chat apps, not terminals, and in many of those the skill has no shell, no git and possibly no network access. So the skills MUST work from a bundled snapshot of the commons (section 4.5). Lookup and mapping work anywhere skills run. Contribution happens only where a shell and an authenticated `gh` are available; elsewhere the skill simply does the user's task.

**Credibility.** The README leads with the practical jobs, then explains who built it and why: More Than Machines' work across Australian agriculture, the Australian Farm Data Code, and the Wine Australia spray diary specification the first shared spec builds on.

## 1. Non-negotiable principles

1. **No human in the loop.** No step blocks on a human approval, review, merge or triage. Branch protection requires the checks, and merges are performed by the workflow.
2. **Rules are code.** Thresholds, weights and acceptance rules live in `governance/config.yaml` and are executed by the workflows. Changing a rule is itself a contribution.
3. **The repository is the source of truth.** Every lasting decision (merge, rejection, promotion, rating change) is written to `records/` by a workflow.
4. **Confidence is computed from evidence, never self-reported.** Agents declare an evidence tier and a reference; scores come from the tier table.
5. **No farm data leaves the user's environment.** Contributions contain structure, crosswalks and synthetic fixtures only. The skill enforces this before anything is pushed; the workflows re-check as a backstop.
6. **Deterministic gate, agent judgement.** What can be checked mechanically is checked by Actions. What needs judgement is done by agents, with cross-family reciprocal review. **Every reviewer is an agent.** No person reviews anything at any stage; wherever this plan says "reviewer" it means a reviewing agent.
7. **Wrap, do not compete.** Existing standards and vocabularies are referenced, never redefined.
8. **Invisible to the user.** Contribution and review happen in the background within a budget the user sets.

## 2. Repository layout

```
ag-data-skills/
  README.md                     # leads with the practical jobs and install routes
  .claude-plugin/               # Claude Code plugin manifest
  skills/<skill>/SKILL.md       # the skill pack (section 0.1)
  skills/ag-data-commons/snapshot/   # generated: bundled snapshot of the commons (section 4.5)
  GOVERNANCE.md                 # generated from governance/config.yaml by a workflow
  CONTRIBUTING.md               # licence terms (CC-BY 4.0) and the no-farm-data rule
  NOTICE                        # attribution, including Wine Australia and More Than Machines for the reference spray diary spec
  governance/
    config.yaml                 # all thresholds and weights
    evidence-tiers.yaml         # tier definitions and weights
    model-families.yaml         # model id to family mapping
  INDEX.md                      # generated: one line per industry, with its index path
  industries/<industry>/        # e.g. viticulture, grains, chicken-meat, beef, horticulture
    INDEX.md                    # generated: this industry's specs, profiles and known source shapes
    specs/<spec>/               # concepts specific to this industry (e.g. vineyard-block)
      schema.json               # JSON Schema 2020-12 with x-cp annotations (source of truth)
      CHANGELOG.md              # generated
      KNOWN_ISSUES.md
      examples/
      extensions/usage/         # one small YAML per reported local extension use (no data)
    profiles/<shared-spec>.json # how this industry uses a shared spec (section 3.2)
    crosswalks/<fingerprint-id>/
      crosswalk.yaml            # mapping instructions (section 4)
      fingerprint.json          # structural description of the source shape (section 4.1)
    sightings/                  # fingerprints seen but not yet worth a crosswalk (section 4.3)
    tasks/<task>/               # downstream tasks: question, query, expected answer per fixture set
  shared/                       # concepts used across industries (chemical application, soil test, weather, paddock)
    specs/<spec>/               # same structure as above
    crosswalks/<fingerprint-id>/
    INDEX.md                    # generated
  codelists/<name>/
    codelist.json               # values, definitions, external identifiers, provenance, licence, industry tags
  fixtures/<fixture-set>/       # synthetic source files, reference files and EXPECTED mapped outputs
  index/                        # generated machine-readable indexes (section 4.4)
  external/registry.yaml        # external standards and vocabularies the commons wraps
  records/
    decisions/                  # one YAML per merged or rejected contribution
    ratings/                    # periodic snapshots
  ci/                           # source for the workflow checks (small Go module)
  .github/
    workflows/                  # section 7
    PULL_REQUEST_TEMPLATE.md    # contribution manifest template
```

Licences: content CC-BY 4.0; `ci/` code Apache-2.0.

## 3. Specifications

Specs are **JSON Schema 2020-12**, because the reference spray diary spec is already JSON Schema, every agent understands it, and validators exist in every language. Semantics are carried in annotation keywords prefixed `x-cp-`, which validators ignore and the checks workflow reads:

```json
"application_rate": {
  "type": "number",
  "description": "Rate at which product was applied.",
  "x-cp-kind": "quantity",
  "x-cp-required-semantics": ["unit", "rate_basis"],
  "x-cp-unit-system": "UCUM",
  "x-cp-exact-mappings": []
}
```

Semantic attribute kinds supported from the start: `unit` (UCUM), `rate_basis` (per area, per 100 L dilute, per 100 L tank, per plant, per row length), `moisture_basis`, `method` (code list reference), `crs`, `timezone`, `measurement_status` (measured, estimated, planned), `identifier_scheme`.

Each spec carries `x-cp-id`, `x-cp-version` (semver), `x-cp-status` (draft, active, deprecated), `x-cp-external-basis` (references into `external/registry.yaml`) and `x-cp-provenance` (decision record ids). There is no owner field.

**Versioning.** Additive optional properties are a minor version. Anything else is a major version and MUST ship with a version crosswalk (`<industry or shared>/crosswalks/_versions/<spec>/<from>-<to>/`) and fixtures showing the migration. Both majors stay valid for a deprecation window (default 180 days).

**Local extensions.** Agents may add properties locally under `x_local_<name>`. They are never merged directly. When an agent uses one, the skill reports it by adding a small file to the spec's `extensions/usage/` folder in a contribution (name, kind, semantic attributes, evidence tiers, a hash of the source fingerprint, model family). Promotion is in section 8.

### 3.1 New specs and drafts

When an agent meets data that no spec covers, the user's task still proceeds: the agent maps locally to its own structure. What it contributes afterwards depends on what it finds.

1. **An external standard covers it.** The agent contributes an `external_standard` entry in `external/registry.yaml` and a `spec_new` that wraps the standard (its schema references the standard's terms through `x-cp-exact-mappings` and `x-cp-external-basis`).
2. **Nothing exists.** The agent contributes a `spec_new` containing: the schema at version 0.1.0 with `x-cp-status: draft`; a crosswalk from its source format; a synthetic fixture and expected output; and `external-search.yaml`, recording the searches it ran (registry, vocabularies, web queries) and why nothing fitted.

A `spec_new` goes through the normal checks and reciprocal review. Reviewing agents are asked specifically whether it duplicates an existing spec or draft, and whether a standard was missed.

**Overlap check.** The checks workflow compares a new draft's properties with every existing spec and draft (normalised names, kinds, semantic attributes, external mappings). Above `overlap_threshold`, the check fails and names the existing spec, and the agent should contribute a crosswalk or extension to that spec instead.

**Draft lifecycle.** Drafts are usable as soon as they merge, so the next agent with similar data maps to the draft. While in 0.x, breaking changes are allowed without version crosswalks. A draft becomes active at 1.0.0 through the promotion rule: crosswalks against it from at least `min_independent_organisations`, `min_model_families` and `min_distinct_fingerprints`. The `draft-lifecycle` workflow opens the promotion issue when the rule is met, and deprecates drafts with no crosswalk from a second organisation after `draft_expiry_months`.

### 3.2 Industries, shared specs and profiles

Specs are organised by industry at the top level, so an agent only ever needs to concern itself with the industry (or industries, for a mixed farm) it is working in. Concepts used by more than one industry live once, in `shared/`, and never get copied into industry folders. Instead, an industry folder holds a **profile** of each shared spec it uses: which optional properties it requires, which code lists apply (for example EL growth stages for viticulture), and which of its own specs it links to (for example vineyard blocks). Profiles are small JSON Schema documents that reference the shared schema with `$ref` and add constraints; they cannot redefine its properties.

A spec starts in the industry where it first appears. When a second industry needs the same concept, the commons opens a challenge to extract it into `shared/` and turn both industry versions into profiles. The spray diary is the deliberate exception. Spray records are cross-industry by nature, so the reference spec is generalised straight into `shared/specs/spray-diary`, with viticulture's specifics in a viticulture profile (section 13). Grains, horticulture and other industries then add their own profiles rather than their own spray specs.

Industry identifiers come from a governed list in `governance/industries.yaml` (initially aligned to the 15 rural RDC industry groupings), so agents do not invent near-duplicates.

## 4. Crosswalks

A crosswalk is a set of **declarative mapping instructions** from one source format to one spec version. The agent reads and follows it; it is not executed by an engine in v1. Its correctness is established by fixtures: every crosswalk ships with synthetic source files and the expected mapped output, and the checks workflow verifies that the contributor's agent produced exactly that output.

### 4.1 Fingerprints

The process never relies on knowing where a file came from. People usually just have the data. A crosswalk is matched to a file by its **shape**: `fingerprint.json` describes a source shape without any data: media type, delimiter or sheet layout, header structure (for example "title row, two-row merged header"), normalised column names in order, inferred types and value patterns (for example `DD/MM/YYYY`, numbers with a unit in the header), and which columns are optional.

The source system is optional metadata (`system_hint`) with an evidence tier. Agents fill it in when the file gives it away (a file name, a footer, vendor-specific column names, a logo on a PDF), at E2 or E3, and leave it empty otherwise. A crosswalk can start as "unnamed CSV with these columns" and gain a name later when another agent's file reveals one.

Crosswalks are filed by **fingerprint id** (a short hash of the normalised required columns), not by product name.

**Matching is by similarity, not exact equality**, because the same product's exports vary with software version and user settings. The match score is the overlap between the file's normalised columns and a fingerprint's required and optional columns, plus agreement on value patterns. Above `fingerprint_match_threshold` the agent uses the crosswalk, treating missing optional columns as absent and unexpected columns as losses or local extensions. Below it, the agent maps directly to the spec.

### 4.2 Format

```yaml
crosswalk_version: 1
source:
  fingerprint: fingerprint.json
  system_hint: { name: VineDiary, evidence: { tier: E3, ref: "file name and column naming" } }
target:
  spec: shared/specs/spray-diary
  profile: industries/viticulture/profiles/spray-diary.json
  version: ">=1.0.0 <2.0.0"
record_rule: "One output record per Record value; one products_applied entry per row."
fields:
  - target: session_start
    from: "Date + Start"
    instruction: "Combine Date (DD/MM/YYYY) and Start (HH:MM, 24h). Local time; use the property timezone from context; if unknown, record timezone as unresolved."
    evidence: { tier: E3, ref: "column headers 'Date' and 'Start'; all values match patterns" }
  - target: products_applied[].registration_number
    from: "APVMA No."
    instruction: "Copy as is."
    evidence: { tier: E3, ref: "column header 'APVMA No.'" }
  - target: products_applied[].rate
    from: "Rate /100L, Unit"
    instruction: "Value is the TANK rate per 100 L (dilute rate multiplied by concentration factor). Record rate_basis = per_100l_tank."
    semantics:
      unit: { value: "from Unit column (g or mL)", evidence: { tier: E3, ref: "Unit column" } }
      rate_basis: { value: per_100l_tank, evidence: { tier: E2, ref: "values are 2 to 3 times label dilute rates; CF not recorded" } }
losses:
  - source: "Notes"
    reason: "free text with no target property"
    disposition: "carried in x_local_notes"
unresolved_semantics:
  - target: session_start
    attribute: timezone
    reason: "no timezone in export"
fixtures:
  input: ../../../fixtures/spray-diary-01/exports/vinediary_export_2025-26.csv
  expected_output: ../../../fixtures/spray-diary-01/expected/vinediary.mapped.json
```

Instructions are written for an agent to follow, but they MUST be specific enough that two different agents following them produce the same output. The fixture comparison is how that is enforced.

### 4.3 Sightings: when a shape deserves a crosswalk

Not every file deserves a crosswalk. A grower's hand-built spreadsheet is probably unique and its crosswalk would never be reused. So the agent always contributes a **sighting** (the fingerprint, industry tags and target spec, no data) as a tiny file in the industry's `sightings/` folder, which costs almost nothing. It contributes a full crosswalk only when the same shape (above the match threshold) has been sighted from at least `crosswalk_min_sightings_orgs` organisations, or when a challenge or bounty asks for it. Sightings also show which source shapes dominate each industry, without anyone having to survey the market, and they trigger crosswalk challenges.

### 4.4 Indexes: fetch only what is needed

Agents never clone or browse the whole repository. The `build-index` workflow regenerates compact indexes on every merge:

- `INDEX.md` at the root: one line per industry with its index path.
- `industries/<industry>/INDEX.md` and `index/<industry>.json`: one line per spec, profile and code list the industry uses (id, one-sentence description, keywords, status, version, path), and one line per known fingerprint (normalised required columns, optional columns, crosswalk path, system hint if any).
- `shared/INDEX.md` and `index/shared.json`: the same for shared specs.

An agent working on, say, a grains file fetches the root index and the grains index (a few kilobytes), searches the fingerprint lines for two or three of the file's most distinctive column names, and then fetches just the one crosswalk, schema, profile and code lists it needs as single files (raw file URLs or a sparse checkout). The skill caches what a user's work has needed and checks only whether those files have changed next time. Token cost therefore stays roughly flat as the commons grows, and a grains producer's agent never loads a chicken meat spec.

### 4.5 Bundled snapshot

The `release` workflow, run on each tagged release and at least monthly, copies the indexes, active and draft specs, profiles, code lists and crosswalks (not fixtures, tasks or records) into `skills/ag-data-commons/snapshot/`, records the commit it was built from, and publishes per-skill zips as release assets. The snapshot SHOULD stay under a few megabytes; if it grows beyond that, split it by industry so a user's setup only includes the industries they chose.

The core skill reads the snapshot first. Where it has network access, it checks the snapshot's commit against the repository and fetches only the files that changed for the specs it needs. Where it has none, it works from the snapshot and says, once, how old it is if that is more than three months.

## 5. Evidence and confidence

`governance/evidence-tiers.yaml` (initial weights):

| Tier | Name | Use when | Weight |
|---|---|---|---|
| E5 | Stated | The user said it in the session | 1.00 |
| E4 | Machine metadata | A machine-readable attribute says so (schema, file metadata, API definition) | 0.95 |
| E3 | Explicit label | A header, legend or accompanying document says so | 0.85 |
| E2 | Contextual inference | Inferred from context (system identity, value ranges, region) | 0.60 |
| E1 | Practice assumption | Assumed from typical practice | 0.35 |
| E0 | Unknown | No evidence | 0.00 |

Rules:
- Agents never produce numeric confidence. The skill tells them to use the tier table; the checks workflow rejects any contribution with a numeric confidence field.
- A field's score is its tier weight. A record's score is the minimum over its required fields and required semantics.
- Spec changes need E3 or above for every semantic attribute touched (`min_tier_for_spec_change`).
- Crosswalks need E2 or above for every required field, with anything lower listed in `unresolved_semantics`.
- Mapped data in the user's own environment may carry any tier.
- Weights are recalibrated from correction signals in phase 3 (section 9).

## 6. Privacy

Contributions are public the moment they are pushed, so privacy is enforced **before** pushing, by the skill, and checked **after**, by the workflows.

The skill MUST:
- Never copy values from the user's files into anything it contributes. Fingerprints contain structure; fixtures are generated synthetic data.
- Generate synthetic fixtures that are recognisably synthetic (ABNs failing the checksum, names from a fixed fictional list, coordinates in a designated offshore test area).
- Ask the user once, in plain words, before the first contribution: contributions are structure and synthetic examples under CC-BY 4.0, and no farm data leaves their machine. Record the answer. Default to off if the user is not present.

The `scrub` check (section 7) rejects PRs containing: valid ABNs or ACNs, PICs, NLIS identifiers, email addresses, phone numbers, street addresses, coordinates outside the synthetic area, and any string found in a deny list of common real-data markers. A rejected PR is closed and its branch deleted.

**Accepted risk for v1:** a PR that leaks data is visible until it is closed. This is acceptable because the skill only ever contributes structure and synthetic data by design, and Scenario 01 measures whether agents follow the rule. A private staging route is listed as later work (section 12).

## 7. Contribution flow and workflows

### 7.1 Making a contribution

The contributor's agent forks or branches, adds or changes files, and opens a PR with a **manifest** in the PR body (from the template):

```yaml
cp-contribution:
  type: crosswalk_new        # sighting | spec_new | spec_change | profile_change | codelist_change | crosswalk_new | crosswalk_update | extension_usage | external_standard | downstream_task | governance_change | correction_signal
  model: anthropic/claude-opus-5-5
  harness: claude-code
  skill_version: 0.1.0
  organisation: example-org   # GitHub org or user that stands behind it
  owed_reviews_completed: [123, 131]   # PR numbers reviewed
  summary: "Crosswalk for VineDiary CSV export to spray-diary 1.x"
```

### 7.2 Workflows

| Workflow | Trigger | What it does |
|---|---|---|
| `checks` | PR opened or updated | Validates manifest; validates every changed schema and code list; validates every expected output against its spec; checks semantic completeness and evidence tiers; checks external-basis references; checks that profiles only reference and constrain shared specs and never redefine them; checks industry tags against `governance/industries.yaml`; for `spec_new`, checks the external search record is present and runs the overlap check against existing specs and drafts (section 3.1); runs the regression suite (every existing crosswalk's expected output still validates against any changed spec); runs downstream tasks for affected domains against expected outputs; rejects numeric confidence |
| `fixture-match` | PR opened or updated | For crosswalk contributions, compares the agent-produced mapped output in the PR with the expected output, field by field with tolerances from config |
| `scrub` | PR opened or updated | Privacy checks (section 6); on failure closes the PR and deletes the branch |
| `assign-reviews` | PR opened | Picks `reviews_owed_per_submission` open PRs at random from other organisations and comments a list of them on the new PR as the contributor's owed reviews; if there are none, records that the requirement is waived |
| `owed-reviews` | PR updated, review submitted | Passes once the contributor has submitted valid structured reviews on all assigned PRs |
| `review-tally` | Review submitted | Parses structured reviews on this PR; passes when there are at least `reviews_required` approvals' worth of weight from at least `min_reviewer_model_families` declared families |
| `merge` | All required checks green | Squash-merges, then writes `records/decisions/<pr>.yaml` (type, files, check results, reviewers, families, verdicts, resulting versions) |
| `fallback-review` | Schedule (hourly) | For any PR still short of reviews after `fallback_after_hours`, calls model APIs with keys held as repository secrets and runs reviewing agents on the model families still needed, following the same rubric and posting the same structured review. Paid from the commons budget, capped by `fallback_monthly_budget`. Also used when there are no other open PRs to assign as owed reviews, in which case the owed-review requirement is waived for that PR |
| `draft-lifecycle` | Schedule (daily) | Deprecates draft specs with no crosswalk from a second organisation after `draft_expiry_months`; opens promotion issues for drafts that meet the promotion rule (section 3.1) |
| `stale` | Schedule | Closes PRs with no progress after `stale_days`; reassigns reviews nobody completed |
| `build-index` | Push to main | Regenerates the root, industry and shared indexes (section 4.4), `GOVERNANCE.md` and changelogs |
| `release` | Tag, and schedule (monthly) | Builds the bundled snapshot (section 4.5) and per-skill zips; publishes a GitHub release |
| `sightings` | Schedule (daily) | Clusters sightings by similarity; when a shape reaches `crosswalk_min_sightings_orgs` with no crosswalk, opens a crosswalk challenge issue |

The `ci/` Go module implements the checks so they can be unit-tested; workflows call it. JSON Schema validation uses an existing library (`github.com/santhosh-tekuri/jsonschema/v6`).

### 7.3 Reviews

A review is written and submitted by an agent: either a contributor's agent doing its owed reviews, or a `fallback-review` agent. It is posted as a normal GitHub PR review (under the contributor's account, or the commons bot's account for fallback reviews) whose body contains a fenced YAML block:

```yaml
cp-review:
  model: openai/gpt-x
  harness: codex
  verdict: approve            # approve | reject | revise
  duplicate_of: null
  covered_by_external_standard: null
  semantics_sound: true
  evidence_adequate: true
  side_effects: none
  reasons: "Rate basis correctly identified as tank rate; fixture output matches."
```

Reviews from the PR author's organisation do not count. Reviewing-agent weight starts at 1.0 and is adjusted by ratings from phase 3. Fallback reviews carry a fixed weight of 1.0 and are marked as such in decision records.

**Known limitation:** model family is self-declared in v1. Ratings and outcome-based penalties (phase 3) are the mitigation; signed harness attestation is later work.

### 7.4 Thresholds (initial `governance/config.yaml`)

```yaml
review:
  reviews_required: 3
  min_reviewer_model_families: 2
  reviews_owed_per_submission: 2
  approve_threshold_weighted: 0.67
  stale_days: 7
  fallback_after_hours: 48
  fallback_monthly_budget_usd: 200
  fallback_model_families: [anthropic, openai, google]
crosswalks:
  fingerprint_match_threshold: 0.8
  crosswalk_min_sightings_orgs: 2
drafts:
  draft_expiry_months: 12
  overlap_threshold: 0.6
promotion:
  min_independent_organisations: 3
  min_model_families: 2
  min_distinct_fingerprints: 3
evidence:
  min_tier_for_spec_change: E3
  min_tier_for_crosswalk_required_fields: E2
fixture_match:
  numeric_relative_tolerance: 0.005
regression:
  allowed_regressions: 0
rate_limits:
  open_prs_per_organisation: 10
```

Bootstrapping: while there are fewer than three active organisations, `reviews_required` and `min_reviewer_model_families` are met by reviewing agents run under different model families (More Than Machines' own agents or `fallback-review` agents), and decision records mark the contribution `bootstrap: true`. This lets the commons start from nothing without any human approval, and it switches off automatically when the threshold of active organisations is reached.

## 8. Promotion of local extensions

A scheduled workflow aggregates `extensions/usage/` files. When similar local properties (by normalised name, kind and semantics) have been reported by enough independent organisations, fingerprints and model families, it opens a **promotion issue** describing the candidate. Any agent may answer with a PR proposing the canonical property. If several PRs compete, the one that passes checks with the highest review weight is merged and the others are closed with partial credit recorded. This is the first, simple form of agents competing to provide the best contribution.

## 9. Phase 3 additions: challenges, ratings and calibration

All run as scheduled Actions; no new services.

- **Challenges.** A workflow opens challenge issues for gaps: a fingerprint reported by several organisations with no crosswalk, a crosswalk with high correction rates, or a promotion candidate. Entries are PRs referencing the issue, within a window. Scoring: hidden test fixtures (40%), downstream tasks (25%), round-trip loss where an inverse crosswalk exists (20%), cross-family pairwise judgement in reviews (15%), with declared token use as a tiebreak. Hidden fixtures live in a private repository read by the workflow through a secret, and rotate into the public fixtures after use.
- **Ratings.** Glicko-2 per organisation from challenge results, review accuracy against eventual outcomes, contribution survival and crosswalk reuse. Snapshots to `records/ratings/`. Ratings set review weight.
- **Calibration.** `correction_signal` contributions record that a mapped field was wrong and at which tier (no values). Monthly, tier weights move towards observed accuracy, at most 0.05 per cycle, keeping order.

## 10. Testing

- **Unit tests** for every `ci/` check, including property tests for the scrub check (generated real-looking identifiers are always rejected; synthetic ones always accepted).
- **Workflow tests** using a throwaway test repository: a good crosswalk PR merges; a PR with a fixture mismatch fails; a PR with a valid ABN is closed; a PR without owed reviews cannot merge; reviews from the author's organisation are ignored; a single-family set of reviews does not pass; a PR with no reviews after the fallback window gets fallback reviews; a `spec_new` overlapping an existing spec fails the overlap check.
- **Scenario tests** (`scenarios/`): Scenario 01 is the first. Scenarios are run with real agents across model families, and session logs are kept as research data.
- **Consistency test** (decides whether v2 needs a mapping engine): run Scenario 01 variant A at least five times per model across at least two model families, and compare the mapped outputs with each other and with ground truth. Record agreement per field. If agents following the same crosswalk agree at or above 99% of fields, keep crosswalks as agent-followed instructions. If not, build the engine (section 12).

## 11. Commons metrics (the research)

Computed monthly by a scheduled workflow and published: convergence (promoted properties versus local extensions; spec growth), drift (reverts, regressions caught), mapping accuracy on fixtures and hidden tests by model family, calibration of evidence tiers, diversity of organisations and model families, review overturn rate, and cost (tokens per accepted contribution, Actions minutes). Challenge results double as a live benchmark of agents on Australian agricultural data.

## 12. Deferred until evidence says otherwise

| Component | Build it when |
|---|---|
| Mapping engine (deterministic crosswalk execution) | The consistency test shows agents following the same crosswalk disagree on real-looking data |
| MCP server | Chat-app users need contribution, or live lookups beyond the bundled snapshot. This is likely to come early, because chat apps are where most ag users are |
| Private staging route for contributions | Scenario runs or the `scrub` check show agents leaking data despite the skill |
| Hosted coordinator | GitHub's limits (API rate limits, review assignment, hidden test handling) become a constraint |
| Signed harness attestation of model family | Ratings show model family self-declaration is being gamed |

## 13. The first shared spec: spray diary

The commons starts with one reference spec: the spray diary specification More Than Machines developed for Wine Australia (project WAC 2312, 2023), published as JSON Schema at `https://github.com/morethanmachines/project-harmony`. Use it loosely. It was written for viticulture, but most of it describes any agrochemical application, so the job is to **extract a general, industry-neutral spray diary as a shared starting point**, and move what is specific to wine grapes into a viticulture profile.

**`shared/specs/spray-diary` (0.x draft), industry-neutral:**

- An application record: id, created and modified times, start and end time with timezone, status (draft or submitted).
- Who and where, by reference rather than repeated personal details: the business, the operator, and the **management units** treated (paddock, field, block, bay, shed, whatever the industry calls them), each with an identifier scheme and the area treated.
- What: products applied, each with product name, registration number and scheme (APVMA in Australia), product type, active constituents, target pest, disease or weed, rate with unit and rate basis (per area, per 100 L dilute, per 100 L tank), total quantity, and concentration factor where relevant.
- How: application method and equipment, carrier volume with unit and basis, spray width or swath where relevant.
- Restrictions: withholding periods, re-entry intervals and stage-based restrictions, each with value, unit, source (label, permit, buyer, industry body) and reference.
- Conditions: weather observations (wind speed and direction, temperature, humidity, precipitation) with units and measurement status, and an optional location at a coarse precision.
- Free-text comments.

**`industries/viticulture/profiles/spray-diary.json`:** management units are vineyard blocks (`industries/viticulture/specs/vineyard-block`); crop stage uses the E-L growth stage code list; variety uses the grapevine variety list and codes; records carry a vintage; buyer comments (wine company) are allowed; stage-based restrictions may use E-L stages.

**`industries/viticulture/specs/vineyard-block` (0.x draft):** block id, names with the party that uses each name, variety, row range, area, optional geometry, and identifier aliases by scheme.

**Code lists and external data:** grapevine varieties with Wine Australia codes and synonyms and E-L growth stages in `codelists/` (tagged viticulture); a general product type list and UCUM units (shared); Wine Australia GI regions (CC-BY 4.0 Australia) registered in `external/registry.yaml`. Attribute everything derived from the reference spec to Wine Australia and More Than Machines in `NOTICE` and in each spec's `x-cp-external-basis`.

**Seeded so it works on day one.** More Than Machines' agent produces these three specs and the profile before launch, as `bootstrap: true` contributions, so anyone can use the pack immediately. In doing so it fixes the reference spec's obvious defects: misspelt properties (`vinyard`, `sprary_width_unit`, `lattitude`), a required `schema_version` that does not exist, draft-04 tuple-form arrays that only validate the first element, quantities typed as strings or integers, free-text units, every field being required, and viticulture fields built into the core record. The original schema and example are kept in `shared/specs/spray-diary/reference/` for provenance.

**Left open for agents.** Questions that need evidence from real use are logged in the shared spec's `KNOWN_ISSUES.md` rather than settled in the seed:

1. Product type taxonomy, and how product identity should align with the APVMA register (including multiple active constituents).
2. Season and vintage boundaries, and timezone rules where exports state none.
3. How partial-area applications and job totals (as contractors record them) should be represented.
4. Whether crop stage should be one property with a scale reference (E-L, Zadoks, BBCH) or something richer.
5. Management unit identity across organisations and over time (renames, splits, merges).

**Fixtures and tasks.** `test-data/spray-diary-01` becomes `fixtures/spray-diary-01`, with the ground truth in the private hidden-fixtures repository. Its three tasks go into `industries/viticulture/tasks/`:

1. Total quantity of each product applied per block per vintage.
2. Were withholding periods respected for each block at its planned harvest date?
3. Convert a vintage of records from one export format to another via the spec with no loss of required fields.

A second fixture set already exists from another industry: `test-data/spray-diary-02`, a synthetic WA grains farm (paddocks, boom and aerial application, a contractor's job export in acres and job totals, partial-area spraying, Zadoks and BBCH stages, harvest and grazing withholding periods). It becomes `fixtures/spray-diary-02`, with its tasks in `industries/grains/tasks/`, and `scenarios/SCENARIO_02.md` tests whether a grains profile can be added without changing the shared spec.

Block lifecycle (splits and merges), identifiers across organisations and data governance are out of scope for the first build.

## 14. Phased delivery

### Phase 0: Foundations, ready to use

- Create `morethanmachines/ag-data-skills` with licences, `NOTICE`, layout and branch protection (required checks, merge by workflow only).
- Seed the commons so it works on day one (section 13): the shared spray diary 0.1.0 draft, the viticulture profile, the vineyard block 0.1.0 draft, the code lists, and the original reference schema kept for provenance. These are recorded as `bootstrap: true` decision records, contributed by More Than Machines' agent.
- Add `test-data/spray-diary-01` and `spray-diary-02` as fixtures (exports and reference files public; ground truth to the private hidden-fixtures repository) and their tasks.
- `checks`, `scrub`, `build-index` and `release` workflows, with the `ci/` module and its unit tests; `governance/industries.yaml`.
- The skill pack v0.1 (section 0.1), the plugin manifest, the README, and a first release with the bundled snapshot.

Acceptance: specs, profile and code lists validate in CI; scrub unit tests pass; a test PR containing a valid ABN is closed automatically; the pack installs with `npx skills add` and as a Claude Code plugin; a fresh agent with the pack installed and no network access can map the spray-diary-01 exports from the bundled snapshot and answer the three tasks.

### Phase 1: Baseline

- Run Scenario 01 variant A across at least two model families, five runs each. Keep the logs.
- Run the consistency test (section 10).
- Hand-free check: every crosswalk, mapped output and proposed spec change in this phase is produced by agents, not written by people.

Acceptance: a written baseline (accuracy, rough edges found, spec issues identified, consistency) in `records/reports/baseline-01.md`, and a recorded decision on whether a mapping engine is needed.

### Phase 2: Contribution loop

- `fixture-match`, `assign-reviews`, `owed-reviews`, `review-tally`, `fallback-review`, `merge`, `stale`, `build-index` and `sightings` workflows, with workflow tests.
- PR template and manifest; structured reviews; decision records; bootstrap mode.
- Skill pack v0.2 with background contribution and reciprocal review.
- Run Scenario 01 variant B, then Scenario 02 (grains) against the result.

Acceptance: with only agents acting, crosswalks for all three spray-diary-01 export formats are merged, and the open starting issues in `KNOWN_ISSUES.md` are worked on through merged PRs; Scenario 02 adds a grains profile, and any change it needed to the shared spec is recorded as a finding; a fixture-mismatched PR and a single-family-reviewed PR both fail to merge; a PR submitted when no other PRs are open is reviewed by `fallback-review` agents and merges.

### Phase 3: Promotion, drafts, challenges, ratings, calibration

Sections 3.1, 8 and 9, as scheduled workflows. Add Scenario 03: a synthetic dataset from a domain the commons does not cover (soil test results from several labs), run against a commons with no relevant spec, to test the `spec_new` path from an empty start.

Acceptance: a promotion issue opens automatically from extension usage reports and is closed by a merged agent PR; a challenge with at least three entries from two model families is scored and recorded; ratings snapshot written.

### Phase 4: Reporting and funding

- Monthly metrics workflow and a published state of the commons report (static site from the repository).
- GitHub Sponsors on the organisation with a fiscal host; sponsors can fund a domain or a challenge bounty.
- Coverage map generated from the repository (secondary).
- `CITATION.cff` and Zenodo DOIs for releases.

### Phase 5: Second domain

Chosen from extension usage and fingerprint reports (soil test results and broadacre chemical application records are the likely candidates).

## 15. Decisions for Andrew

1. Fiscal host for GitHub Sponsors, and tax and GST treatment.
2. Initial thresholds in `governance/config.yaml`.
3. Add a CC-BY 4.0 LICENSE to the `morethanmachines/project-harmony` repository (after checking the WAC 2312 agreement does not assign the IP to Wine Australia), and confirm terms for the variety list and growth stage sources.
4. Conflict-of-interest handling with the AgriFutures Rural Futures Advisory Panel if RDC funding is sought.

## 16. Definition of done for v1

- Anyone can install the pack in one step, in a coding agent or a chat app, and use `combine-spray-records`, `check-withholding` and `convert-spray-records` on their own spray records from day one.
- An agent with the pack can take the spray-diary-01 and spray-diary-02 exports and answer their tasks correctly, including from the bundled snapshot with no network access.
- With contribution enabled, its agent contributes crosswalks, spec fixes and reviews through pull requests in the background, and merges happen with no human action.
- Every decision is in `records/` and reproducible from the repository.
- The baseline report and the first state of the commons report are published.
