# Shared spray diary (`shared/spray-diary`)

Status: **draft 0.1.0**. Industry-neutral record of agrochemical applications.

- `schema.json`: JSON Schema 2020-12 with `x-cp-` annotations (source of truth).
- `KNOWN_ISSUES.md`: open questions left for agents to settle with evidence.
- `examples/`: synthetic examples that validate against the schema.
- `extensions/usage/`: reports of local extensions in use (no data).

A document is `{ "spec": "shared/spray-diary", "spec_version": "0.1.0", "records": [...] }`.
Each record is one application event: one tank mix applied in one session to one or more
management units. Industries constrain it with a profile, for example
`industries/viticulture/profiles/spray-diary.json`.

Mapping conventions that matter for answering questions correctly:

- **Quantities** are `{ "value": number, "unit": "<UCUM code>" }` with codes from
  `codelists/units`. Hectare is `har`.
- **Rates** add a `basis` from `codelists/rate-basis`. The amount applied to a unit is:
  - `per_hectare`: rate x area treated (ha)
  - `per_100L_tank`: rate x carrier rate (L/ha) x area treated / 100
  - `per_100L_dilute`: rate x concentration factor x carrier rate (L/ha) x area treated / 100
- **Times** are local wall-clock time with an optional IANA `timezone` and `utc_offset`.
  If the source does not settle the timezone, leave it out and list it in
  `mapping.sources[].unresolved_semantics`.
- **Evidence** for every mapped field goes in `mapping.sources[].fields[]` as a tier and a
  reference. Never a numeric confidence.

Derived loosely from the spray diary specification More Than Machines developed for Wine
Australia in 2023. See `NOTICE`.
