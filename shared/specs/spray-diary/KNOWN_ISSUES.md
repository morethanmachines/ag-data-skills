# Shared spray diary: known issues

These questions need evidence from real use. They are deliberately not settled in the 0.1.0
seed. Agents work on them through pull requests (`spec_change`, with evidence at E3 or
above for every semantic attribute touched); nobody settles them by editing the spec
directly.

## Open questions

1. **Product type taxonomy and product identity.** `codelists/product-types` is a minimal
   starting list. How product identity should align with the APVMA register is open: one
   product with several active constituents, permits as opposed to registrations, the same
   product under several trade names, and products named without their registration number
   (as in two of the three spray-diary-01 exports).
2. **Season, vintage and timezone rules.** `season` has a scheme and a label, but the
   boundaries of a vintage or crop year are not defined. Times are recorded as local time
   with an optional IANA timezone and UTC offset. When an export states no timezone (all six
   fixture exports), what may an agent infer from the business's location, and at what
   tier? How should times near a daylight saving change be handled?
3. **Partial-area applications and job totals.** `management_units[].area_treated` records
   the area actually sprayed, and `total_quantity` is a total for the whole record.
   Contractors record product and water as job totals across several paddocks
   (spray-diary-02). Should a record carry per-unit quantities, and how should a job total
   be apportioned?
4. **Crop stage.** `crop_stage` is one object with a scale reference (`E-L`, `Zadoks`,
   `BBCH`, `text`). Is that enough, or does it need more (several stages for a mixed
   block, stage ranges, the crop the stage refers to)?
5. **Management unit identity across organisations and over time.** Units are referenced by
   `id` and `id_scheme`. Renames (spray-diary-01 has one mid-season), splits, merges and the
   same unit known by different names in different systems are not modelled in the shared
   spec. The viticulture vineyard-block spec holds names by party with validity dates.

## Seed limitations

6. **Nothing copied from the reference spec.** This draft follows the reference spec
   loosely. It includes no verbatim reference schema or example, no grapevine variety list
   and no E-L stage descriptions.
7. **Units use UCUM codes.** UCUM writes hectare as `har` (`ha` would be a hecto-annum).
   `codelists/units` lists the codes the spec expects with plain labels. Whether agents
   reliably use UCUM codes rather than familiar labels is a question for the
   consistency test across model families.
8. **Rate basis.** `codelists/rate-basis` fixes the denominators (per hectare, per 100 L
   dilute, per 100 L tank, per plant, per 100 m of row). Rates per acre must be converted.
   Whether a fixed list is better than a free denominator quantity is open.

## Differences from the reference spec

Recorded for provenance (decision record `bootstrap-0001`). This draft is industry-neutral:
viticulture fields (vineyard, variety, E-L stage, vintage, wine company comment) move to
the viticulture profile. It uses JSON Schema 2020-12, types quantities as numbers with
UCUM units, keeps optional fields optional, and refers to parties rather than repeating
their details in every record.
