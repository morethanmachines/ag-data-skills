---
name: convert-spray-records
description: Convert spray records into another system's layout (for example a spray diary app export into a wine company's or farm software's import format), or into a buyer's or auditor's format, by going through the shared spray diary spec, and report anything that cannot be carried across. Use when the user wants records reformatted for loading into another system or for an audit.
---

# Convert spray records

Uses the `ag-data-commons` skill. Always convert *via the spec*: source to mapped spray diary, then mapped spray diary to the target. Never convert column to column directly; the spec step is what makes units, rate bases and identifiers explicit.

## Steps

1. **Map the source** to the shared spray diary (with the industry profile) following `ag-data-commons`.
2. **Learn the target layout.** Best: a crosswalk for the target shape in the commons index (follow it in reverse). Otherwise: a sample export or template from the target system (the user often has one), or the target's documentation. Write down its shape: columns in order, one row per what (event, unit, product), date and time formats, identifiers used (whose block or paddock codes), units and rate bases, whether water is a rate or a total.
3. **Write each target row from the mapped records**, converting explicitly:
   - rates between bases (per 100 L tank or dilute to per hectare: rate x carrier rate (L/ha) / 100, times concentration factor for dilute rates);
   - water between L/ha and total litres (multiply or divide by area treated);
   - areas (acres to hectares: x 0.404686), units (mL to L, g to kg) and wind speeds (knots to km/h: x 1.852; m/s to km/h: x 3.6);
   - identifiers (grower names to buyer codes, contractor codes to paddock names) with the reference list, respecting renames;
   - job totals to per-unit rates by dividing by the total area treated.
4. **Check the result** against any rows the target system already has for the same events, if the user has them: same dates, units, products and rates within 0.5%.
5. **Report losses**: source fields with no target column (for example registration numbers, operator names, weather units), and anything the target needs that the source lacks.

## What to give the user

The converted file in the target layout (same columns and order as the target sample), a one-line count of rows written, the list of losses, and any conversion that rested on an inference (for example an unlabelled unit).
