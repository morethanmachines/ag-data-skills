# Common jobs with mapped spray records

Do every job from records mapped to the shared spray diary (`shared/spray-diary`, with the industry profile), never from the raw files. The mapping step is what makes units, rate bases and identifiers explicit.

## Combining records from several sources

1. **Inventory the sources**: what each file seems to be (app export, buyer system, contractor jobs, own spreadsheet, reference list) and its layout. Read reference files (block or paddock lists, product registers) as context, not as records.
2. **Map each source** separately: one record per application event, management units by reference, products with registration numbers, rates with their basis, UCUM units, evidence for every field.
3. **Join management units** across sources with the reference list. Check for renames (one name used before a date and another after it) and for codes in one system and names in another.
4. **De-duplicate.** Sources usually describe the same applications. Match events on date, management unit and product, allowing for name and unit differences. Never count the same event twice. Keep the most complete source as primary for each event and note which sources agreed.
5. **Compare.** For every matched event, compare the amount of each product applied per unit (rate times area, by basis, see `mapping.md`). Differences above 1% are conflicts: report them with the likely reason (rate basis, concentration factor, area, unit), and don't silently choose. Events in only one source are gaps: report them.
6. **Give the user** totals of each product per block or paddock, using the identifiers they asked for, in kg for products measured by mass and L for liquids. If all sources agree within 1%, say so: that is strong evidence the mapping is right.

## Checking withholding periods

This checks the records the user supplied. It is not agronomic or legal advice: the product label or permit is the authority.

1. **Attach restrictions.** For each product application, take its withholding periods from the source the user gave (a product register, label extracts, a buyer's schedule). Record each as a restriction: `type: withholding_period`, `applies_to: harvest` or `grazing`, `period: {value, unit: "d"}`, `source` (`label`, `permit`, `buyer`) and `reference`. If a product has no known period, say so: never assume zero.
2. **Read planned dates** per block or paddock, joining identifiers as above.
3. **Compute**, per block or paddock and product:
   - last application date (local date of the application);
   - days before harvest = planned harvest date minus last application date, in whole days;
   - a breach when days before harvest is less than the harvest withholding period. Equal is compliant;
   - earliest grazing date = the latest, over every application to the unit, of application date plus that product's grazing withholding period.
4. **Explain each breach** in one line: block or paddock, product, last application, withholding period, days short, and the earliest compliant harvest date (last application plus the period). Mention any product without a known period, any record whose date or block could not be resolved, and any timezone doubt that could move an application across midnight.

## Converting to another layout

Always go source to mapped spray diary to target. Never convert column to column directly.

1. **Learn the target layout.** Best: a crosswalk for the target shape in the commons index, followed in reverse. Otherwise a sample export or template from the target system, or its documentation. Write down its shape: columns in order, one row per what (event, unit, product), date and time formats, whose identifiers it uses, units and rate bases, whether water is a rate or a total.
2. **Write each target row from the mapped records**, converting explicitly:
   - rates between bases (per 100 L tank or dilute to per hectare: rate times carrier rate (L/ha) / 100, times the concentration factor for dilute rates);
   - water between L/ha and total litres (multiply or divide by area treated);
   - areas (acres to hectares: times 0.404686), units (mL to L, g to kg) and wind speeds (knots to km/h: times 1.852; m/s to km/h: times 3.6);
   - identifiers (grower names to buyer codes, contractor codes to paddock names) with the reference list, respecting renames;
   - job totals to per-unit rates, by dividing by the total area treated.
3. **Check the result** against any rows the target system already has for the same events: same dates, units, products and rates within 0.5%.
4. **Report losses**: source fields with no target column, and anything the target needs that the source lacks. Give the user the file in the target layout (same columns and order as the target sample), a count of rows written, the losses, and any conversion that rested on an inference.
