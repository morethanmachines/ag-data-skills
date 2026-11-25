# Test dataset: spray-diary-01

A small synthetic dataset for testing whether agents can map messy, real-looking spray diary exports to a common specification, answer questions from them, and spot where the specification falls short.

**Everything here is fictional.** The vineyard, people, products, registration numbers (`SYN-9000x`) and withholding periods are invented. The ABN deliberately fails the ABN checksum, and the block polygons sit offshore in Gulf St Vincent so they cannot match a real property. Do not use any of it for real spraying decisions. Licensed CC-BY 4.0.

## The story

Synthetic Ridge Vineyard has four blocks (Shiraz, Chardonnay, Cabernet Sauvignon, Pinot Noir) and supplies one wine company. During the 2025/26 season (vintage 2026) there were 16 application events: 14 canopy sprays, often tank mixes, and 2 undervine herbicide passes. The same season is recorded in three places, each in its own way.

## Files

| Path | Give to agent? | What it is |
|---|---|---|
| `exports/vinediary_export_2025-26.csv` | Yes | Spray diary app export. One row per block per product. Grower block names, registration numbers, tank rates per 100 L, labelled weather units. |
| `exports/grapelog_export_2025-26.csv` | Yes | Wine company system export. One row per block, products in numbered columns. Wine company block codes, rates per hectare, no registration numbers, total water per block, unlabelled weather. |
| `exports/spray_diary_2025-26.xlsx` | Yes | Grower's own spreadsheet. Title row, merged two-row headers, product names without the "Synth" prefix, dilute rates plus concentration factor. |
| `reference/vineyard_blocks.geojson` | Yes | Blocks in the reference spray diary spec's block format, with all block names in one unlabelled array. |
| `reference/product_register.csv` | Yes | Synthetic product register (stands in for the APVMA register). |
| `reference/harvest_dates.csv` | Yes | Planned harvest dates by wine company block code. |

## Rough edges built in

All three exports describe the same applications, so they should give the same answers once they are understood. Getting there means dealing with the following:

- Blocks have different names in each system, and the block reference file lists them in an unlabelled array.
- One block was renamed mid-season. One system uses the old name before the rename, and the spreadsheet uses the new name all season.
- Rates are expressed three ways: tank rate per 100 L with no concentration factor recorded, dilute rate per 100 L with the concentration factor, and rate per hectare.
- Herbicide rates are per hectare in one file and per 100 L of tank mix in another.
- Product names vary, and only one export carries registration numbers.
- Water is in litres per hectare in two files and total litres per block in the third.
- Weather units are missing in one export (wind there is in m/s).
- Times are local to Adelaide, cross a daylight saving change, and never state a timezone.
- There are two withholding period breaches to find.

## Ground truth and generator

The ground truth (expected answers and the underlying records) and the deterministic generator are held privately by More Than Machines, so that agents under test cannot read the answers. They are not in this repository. Scores are produced against them separately. The downstream tasks that use this fixture set are defined in `industries/viticulture/tasks/`.
