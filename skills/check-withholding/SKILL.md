---
name: check-withholding
description: Check harvest and grazing withholding periods for agrochemical applications against planned harvest or grazing dates, block by block or paddock by paddock, and explain any breach. Also gives the earliest date stock can graze each paddock after its last spray. Use when the user asks whether spraying is compliant with withholding periods (WHP), export slaughter intervals or buyer restrictions before harvest or grazing.
---

# Check withholding periods

Uses the `ag-data-commons` skill for lookup and mapping, and `combine-spray-records` when records come from more than one source. This is a check of the records the user supplied, not agronomic or legal advice: the product label or permit is the authority.

## Steps

1. **Get mapped, de-duplicated records** for the season (run `combine-spray-records` first if there are several sources).
2. **Attach restrictions.** For each product application, find its withholding periods from the source the user gave (a product register, label extracts, a buyer's schedule). Record each as a restriction: `type: withholding_period`, `applies_to: harvest` or `grazing`, `period: {value, unit: "d"}`, `source` (`label`, `permit`, `buyer`) and `reference`. If a product has no known period, say so: never assume zero.
3. **Planned dates.** Read planned harvest (and grazing) dates per block or paddock, joining identifiers as in `combine-spray-records`.
4. **Compute, per block or paddock and product:**
   - last application date (local date of the application);
   - days before harvest = planned harvest date minus last application date, in whole days;
   - a breach when days before harvest is less than the harvest withholding period. Equal is compliant.
   - earliest grazing date = the latest, over every application to the unit, of application date plus that product's grazing withholding period.
5. **Explain each breach** in one line: block or paddock, product, last application, withholding period, days short, and the earliest compliant harvest date (last application plus the period).

## What to give the user

A table per block or paddock (compliant or not, breaches, earliest grazing date if asked), and JSON if they want a file, shaped like:

```json
[{"block": "...", "harvest_date": "YYYY-MM-DD", "compliant": false,
  "violations": [{"registration_number": "...", "product_name": "...", "last_application": "YYYY-MM-DD",
                  "withholding_period_days": 21, "days_before_harvest": 17}]}]
```

Mention any product without a known withholding period, any record whose date or block could not be resolved, and any timezone doubt that could move an application across midnight.
