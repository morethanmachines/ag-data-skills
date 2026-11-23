---
name: combine-spray-records
description: Merge agrochemical spray records from several sources (spray diary app exports, a wine company or buyer system, contractor job histories, the grower's own spreadsheet) into one clean, de-duplicated set mapped to the shared spray diary spec, then report totals, conflicts and gaps. Use when the user has spray or chemical application records in more than one place or format, or wants totals of product applied per block or paddock.
---

# Combine spray records

Uses the `ag-data-commons` skill for all lookup and mapping. Follow it first.

## Steps

1. **Inventory the sources.** List every file the user gave you, what each seems to be (app export, buyer system, contractor jobs, own spreadsheet, reference list), and its layout (title rows, merged headers, one sheet per unit, JSON). Read reference files (block or paddock lists, product registers) as context, not as records.
2. **Map each source** to the shared spray diary (`shared/spray-diary`) with the industry's profile if there is one, following `ag-data-commons` and its `references/mapping.md`: one record per application event, management units by reference, products with registration numbers, rates with their basis, units as UCUM codes, evidence for every field. Keep one mapped document per source.
3. **Join management units** across sources with the reference list. Check for renames (a name used before a date and another after it) and for codes in one system and names in another.
4. **De-duplicate.** Sources usually describe the same applications. Match events on date, management unit and product (allowing for name and unit differences). Never add the same event twice. Keep the most complete source as primary for each event and note which sources agreed.
5. **Compare.** For every matched event, compare the amount of each product applied per unit (rate x area by basis, see `mapping.md`). Differences above 1% are conflicts: report them with the likely reason (rate basis, concentration factor, area, unit), and do not silently choose. Events in only one source are gaps: report them.
6. **Write the combined set** as one mapped document (validated against the profile or spec), and a short table for the user.

## What to give the user

- Totals of each product per block or paddock for the season, using the identifiers they asked for (for example the buyer's block codes), in kg for products measured by mass and L for liquids. Save as JSON and CSV if they want files.
- Whether the sources agree. If all sources give the same totals within 1%, say so; that is strong evidence the mapping is right.
- Conflicts, gaps and anything you were unsure of (unresolved units or rate bases, renamed blocks, timezone), briefly.
- Not: the evidence table, fingerprints or anything about the commons unless they ask.

Afterwards, if contribution is on, contribute as `ag-data-commons` describes (sightings for each source shape).
