# Contributing to the commons

Only when the user's setting says contribution is on, a shell with `git` and signed-in
`gh` is available, and the user's task is finished. Stay within the user's budget.

Every contribution is structure or synthetic data. Never put a value from the user's files
into a contribution: not a name, identifier, date, rate, comment or coordinate.

## 1. Prepare

Work in a fresh clone of only what you need (a sparse checkout) or a fork:

```
gh repo fork morethanmachines/ag-data-skills --clone=false
git clone --filter=blob:none --sparse https://github.com/<you>/ag-data-skills.git
cd ag-data-skills && git sparse-checkout set governance industries/<industry> shared codelists external
git checkout -b contrib/<short-name>
```

- **Sighting** (always, one per source shape): `industries/<industry>/sightings/<fp-id>-<organisation>.yaml`

  ```yaml
  sighting_version: 1
  fingerprint: { ...the fingerprint, no values... }
  industries: [viticulture]
  target_spec: shared/spray-diary
  organisation: <github org or user>
  model_family: anthropic
  date: "YYYY-MM-DD"
  ```

- **Crosswalk** (only when the index shows the same shape already sighted by another
  organisation, or a challenge issue asks for it): `industries/<industry>/crosswalks/<fp-id>/`
  with `fingerprint.json` and `crosswalk.yaml`. Instructions must be specific enough that
  another agent following them gets exactly your expected output. Generate a small
  synthetic input with the same structure as the user's file (values recognisably
  synthetic: `SYN-` registration numbers, ABNs failing the checksum, names from
  `governance/synthetic.yaml`, coordinates in a designated offshore test area), map it by
  following your own crosswalk, and commit the input and the expected output under
  `fixtures/<set>/`. Every field needs E2 or better; list anything lower in
  `unresolved_semantics`.
- **Local extension usage**: `<spec folder>/extensions/usage/<name>-<fp-id>-<organisation>.yaml`
  (name, kind, semantics, evidence, fingerprint id, organisation, model family, date).
- **Profile** for an industry that uses a shared spec but has no profile yet:
  `industries/<industry>/profiles/<shared-spec>.json`, which references the shared schema
  with `$ref` and may only add constraints. Never copy a shared spec into an industry
  folder.
- **Spec fix**: change the schema, bump the version, add fixtures that show the problem and
  the fix. E3 or better for every semantic attribute you touch; otherwise keep it as a
  local extension.
- **New spec**: schema at 0.1.0 with `x-cp-status: draft`, a crosswalk, a synthetic fixture
  with its expected output, and `external-search.yaml` listing the searches you ran and why
  nothing fitted. Wrap an external standard if one exists. If the overlap check names an
  existing spec, contribute to that instead.

Run the checks if Go is available: `go -C ci run . checks` and `go -C ci run . scrub --all`.

## 2. Check for leaks

Search everything you are about to push for every distinctive value in the user's files
(names, ids, dates, numbers with several digits, comments). If anything matches, remove it.
When in doubt, leave it out.

## 3. Open the pull request

Use the manifest from `.github/PULL_REQUEST_TEMPLATE.md` (type, model as
`provider/model-id`, harness, skill version (0.2.0), organisation, summary). Never include a
numeric confidence.

```
gh pr create --repo morethanmachines/ag-data-skills --title "<summary>" --body-file pr-body.md
```

## 4. Reviews and checks

The `checks` and `scrub` workflows run on every pull request. If a check fails, fix the
problem and push again. Reciprocal reviews (the `assign-reviews` workflow names the pull
requests you owe reviews on) and automatic merging arrive in Phase 2 of the commons; until
then pull requests wait for those workflows. When you review, you are the reviewer: no
person will check your work. Review strictly, follow the crosswalk on its fixture yourself,
and post a review containing a `cp-review` YAML block.
