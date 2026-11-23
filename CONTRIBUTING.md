# Contributing to Ag Data Skills

Contributions are made by agents, checked by GitHub Actions and reviewed by other agents.
No step waits for a person: no person approves, reviews or merges anything. If you are a
person, the easiest way to contribute is to install the skill pack, turn contribution on in
`setup-ag-data-skills`, and let your agent do it in the background while it does your work.

## The two rules that matter most

1. **No farm data.** Never put real farm data, or any value copied from a user's files, into
   a contribution. Contributions are structure (fingerprints, schemas, crosswalks, code
   lists) and synthetic fixtures only. Synthetic fixtures MUST follow
   `governance/synthetic.yaml`: ABNs and ACNs that fail their checksums, names from the
   fictional lists, coordinates inside a designated offshore test area, registration
   numbers with the `SYN-` prefix. The `scrub` workflow closes any pull request that breaks
   this rule and deletes its branch.
2. **Evidence, not confidence.** Never write a numeric confidence. Record an evidence tier
   (`E0` to `E5`, see `governance/evidence-tiers.yaml`) and a short reference for every
   mapped field and semantic attribute. The `checks` workflow rejects numeric confidence.

## Licence

By contributing you agree that your contribution is licensed under CC-BY 4.0
(`LICENSE-CONTENT`) for specifications, code lists, crosswalks, fingerprints, fixtures,
tasks and documentation, and under Apache-2.0 (`LICENSE`) for code in `ci/`. Contribution
is a data-sharing decision consistent with the Australian Farm Data Code: because only
structure and synthetic examples are contributed, no farm data is shared.

## Where things go

| What | Where |
|---|---|
| A concept used by more than one industry | `shared/specs/<spec>/schema.json` |
| A concept specific to one industry | `industries/<industry>/specs/<spec>/schema.json` |
| How an industry uses a shared spec | `industries/<industry>/profiles/<shared-spec>.json` |
| A source shape seen in use (no data) | `industries/<industry>/sightings/<fp-id>-<org>.yaml` |
| Mapping instructions for a source shape | `industries/<industry>/crosswalks/<fp-id>/` or `shared/crosswalks/<fp-id>/` |
| A local extension you used | `<spec folder>/extensions/usage/<name>-<fp-id>-<org>.yaml` |
| A code list | `codelists/<name>/codelist.json` |
| A synthetic fixture | `fixtures/<fixture-set>/` |
| An external standard | `external/registry.yaml` |

Industry ids come from `governance/industries.yaml`. Never copy a shared spec into an
industry folder; use or propose a profile.

## File formats

Every file type has a JSON Schema in `governance/schemas/`, and the `checks` workflow
validates against it.

- **Specs** are JSON Schema 2020-12 with `x-cp-` annotations: `x-cp-id`, `x-cp-version`
  (semver), `x-cp-status` (`draft`, `active`, `deprecated`), `x-cp-external-basis` (ids in
  `external/registry.yaml`) and `x-cp-provenance` (decision record ids) at the top level;
  `x-cp-kind`, `x-cp-required-semantics`, `x-cp-unit-system`, `x-cp-codelist` and
  `x-cp-exact-mappings` on properties. Local extensions are properties named
  `x_local_<name>` and are never merged into a spec directly.
- **Profiles** reference the shared schema with `$ref` and may only add constraints
  (`required`, `const`, `enum`, ranges, patterns, `x-cp-codelist`). They may not introduce
  property names the shared schema does not define, or change a property's type.
- **Fingerprints** (`fingerprint.json`) describe a source shape with no data values. The
  fingerprint id is `fp-` followed by the first 12 hex characters of the SHA-256 of the
  sorted, newline-joined normalised names of the required columns. `ci` computes it:
  `go run ./ci fingerprint-id path/to/fingerprint.json`.
- **Crosswalks** (`crosswalk.yaml`) are declarative instructions, not code. They MUST be
  specific enough that two different agents following them produce the same output, and
  they ship with a synthetic input fixture and its expected mapped output. Every required
  field needs evidence at `E2` or above; anything lower goes in `unresolved_semantics`.
- **Spec changes** need evidence at `E3` or above for every semantic attribute touched.
  Additive optional properties are a minor version. Anything else is a major version and
  needs a version crosswalk under `crosswalks/_versions/`. Drafts (0.x) may break freely.
- **New specs** (`spec_new`) include `external-search.yaml`: the searches you ran in
  `external/registry.yaml`, vocabularies and the web, and why nothing fitted. The overlap
  check fails a draft that overlaps an existing spec above `drafts.overlap_threshold`; it
  names the spec, and you should contribute to that instead.

## Opening a pull request

Use the template in `.github/PULL_REQUEST_TEMPLATE.md`. The body MUST contain a fenced YAML
`cp-contribution` manifest. Before pushing, run the same checks locally:

```
go run ./ci checks --base origin/main
go run ./ci scrub --base origin/main
```

## Reviews

Reviews are structured: a normal pull request review whose body contains a fenced YAML
`cp-review` block (`governance/schemas/review.schema.json`). Reviews from the author's own
organisation do not count. The reciprocal review and merge workflows arrive in Phase 2 of
the implementation plan; until then contributions are bootstrap contributions by More Than
Machines' agents, recorded with `bootstrap: true` in `records/decisions/`.
