# Mapping farm data to a commons spec

## The mapped document

Every spec document has `spec`, `spec_version`, optional `profile`, the records, and a
`mapping` section that records how it was produced. For the shared spray diary:

```json
{
  "spec": "shared/spray-diary",
  "spec_version": "0.1.0",
  "profile": "viticulture/spray-diary",
  "records": [ ... ],
  "mapping": {
    "sources": [{
      "fingerprint_id": "fp-...",
      "crosswalk": null,
      "system_hint": "optional",
      "fields": [
        { "target": "records[].products_applied[].rate", "from": "Rate /100L, Unit",
          "evidence": { "tier": "E3", "ref": "header 'Rate /100L'" },
          "semantics": { "rate_basis": { "value": "per_100L_tank", "evidence": { "tier": "E2", "ref": "values are two to three times label dilute rates; no CF column" } } } }
      ],
      "losses": [ { "source": "Notes", "reason": "free text", "disposition": "comments[]" } ],
      "unresolved_semantics": [ { "target": "records[].start", "attribute": "timezone", "reason": "no timezone in export" } ]
    }]
  }
}
```

The mapping section holds structure and evidence only, never source values.

## Records and joins

- One record is one application event: one tank mix applied in one session, to one or more
  management units. Several source rows (one per product, one per block) usually make one
  record. Use the source's own event id when it has one; otherwise derive one and set
  `record_id_scheme` to `derived`.
- Management units are referenced by `id` and `id_scheme`, with the source's `name`. Join
  names across systems with a reference list (a block or paddock register). Names change:
  check validity dates and renames before joining, and never assume one name for a whole
  season.
- When several sources describe the same applications, they are views of the same events,
  not extra events. De-duplicate on date, management unit and product (allowing for name
  differences), and keep the source with the most complete fields as primary. Report
  conflicts rather than silently picking one.
- Products: keep the name as recorded, and add `registration` (number and scheme `APVMA`)
  from the source or a product register. Match names that differ only by prefix, case,
  spacing or abbreviation, and record the evidence tier of the match.

## Quantities, units and rates

- Quantities are `{ "value": number, "unit": "<UCUM code>" }` with codes from
  `codelists/units`. Hectare is `har` in UCUM. Acre is `[acr_us]`. Knot is `[kn_i]`.
  Degree Celsius is `Cel`.
- A rate adds `basis` from `codelists/rate-basis`:
  - `per_hectare`: amount per hectare treated.
  - `per_100L_dilute`: amount per 100 L at the label (dilute) rate. Record the
    `concentration_factor` too. Tank rate = dilute rate x concentration factor.
  - `per_100L_tank`: amount per 100 L of what was actually in the tank (any concentration
    factor already applied).
  - `per_plant`, `per_100m_row`.
- Amount of product applied to one unit:
  - `per_hectare`: rate x area treated (ha).
  - `per_100L_tank`: rate x carrier rate (L/ha) x area treated (ha) / 100.
  - `per_100L_dilute`: rate x concentration factor x carrier rate (L/ha) x area treated / 100.
- Use the area actually treated (`area_treated`), not the unit's total area, when they
  differ (spot or partial spraying).
- Job totals (product or water for a whole job across several units) go in
  `total_quantity` or `carrier.total`; derive per-hectare rates only by dividing by the
  total area treated, and record the evidence for that.
- Water volume is `carrier.rate` (L per hectare) or `carrier.total` (litres for the whole
  application). A column of plain litres per block is a total, not a rate: check it
  against area.
- Never guess a unit silently. Unlabelled weather columns get the most likely unit from
  context at E2 at best (value ranges, system identity, region), with the reasoning in
  `ref`. Wind values around 1 to 6 are typically m/s; around 5 to 25 km/h; aircraft
  records often use knots. Say so in `unresolved_semantics` if it matters to the task.

## Times

- Times are local wall-clock (`local`) with an IANA `timezone` and `utc_offset` when known.
  An export that states no timezone is common. If the business location settles it (for
  example a South Australian vineyard), record the timezone at E2 with that reasoning, and
  watch for daylight saving changes inside the season. Otherwise leave it out and list it
  as unresolved.
- Dates such as `22 Apr` with the year only in a sheet name or title: take the year from
  there (E3) and say so.

## Crop stage, season and restrictions

- `crop_stage` has a `scale` (`E-L`, `Zadoks`, `BBCH` or `text`), a `value` and the source
  `text`. Profiles fix the scale (viticulture uses E-L).
- `season` has a `scheme` (`vintage`, `crop_year`, `calendar_year`, `financial_year`), a
  `label` and a `year`.
- Withholding periods and other restrictions sit on each product application as
  `restrictions[]` with `type`, `applies_to` (`harvest`, `grazing`, ...), `period`
  (`{value, unit: "d"}`), `source` (`label`, `permit`, `buyer`, ...) and `reference`.
  A product register or label is the usual source.

## Validating

Validate the mapped document against the profile if there is one (it includes the shared
spec), otherwise against the spec:

```
python3 scripts/validate.py snapshot/industries/viticulture/profiles/spray-diary.json mapped.json
```

If no validator is available, check by hand that every `required` property is present and
every code is from its code list, and say that validation was manual.
