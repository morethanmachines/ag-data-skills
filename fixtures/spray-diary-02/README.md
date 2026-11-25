# Test dataset: spray-diary-02 (grains)

A synthetic grains dataset that tests whether the **shared** spray diary spec is genuinely industry-neutral. It is deliberately different from `spray-diary-01` (viticulture) in almost every way that matters: paddocks instead of blocks, broadacre boom and aerial application instead of canopy sprays, rates per hectare instead of per 100 L, a spray contractor alongside the farmer, partial-area spraying, different crop stage scales, and grazing withholding periods as well as harvest ones.

**Everything here is fictional.** The farm, people, contractor, products, registration numbers (`SYN-9100x`) and withholding periods are invented, and the paddock polygons sit offshore in the Indian Ocean west of Geraldton so they cannot match a real property. Do not use any of it for real spraying or grazing decisions. Licensed CC-BY 4.0.

## The story

Synthetic Plains is a Western Australian grains farm with four paddocks in the 2025 season: wheat, barley, canola and lentils (361 ha in total). There were 10 application events: a knockdown before sowing, pre-emergent herbicides, post-emergent and spot spraying, fungicides, an aerial insecticide on canola, and a desiccation spray on lentils. A contractor did four of them; the farmer did the rest.

## Files

| Path | Give to agent? | What it is |
|---|---|---|
| `exports/paddockbook_spray_records_2025.csv` | Yes | Farm management software export. All 10 events. One row per paddock per product, rates in L/ha, mL/ha or g/ha, area actually sprayed, crop stage, Delta T. |
| `exports/sprayjob_contractor_jobs_2025.json` | Yes | The contractor's job history. Contractor jobs only. Paddock codes instead of names, areas in acres, product as job totals, water as a job total, aerial wind in knots. |
| `exports/spray_records_by_paddock_2025.xlsx` | Yes | The farmer's spreadsheet. One sheet per paddock, year only in the sheet name, dates such as "22 Apr", units typed as "ml", "g" or "L" (per hectare, unstated). |
| `reference/paddock_register.csv` | Yes | Paddock names, contractor codes, crops, areas and planned harvest dates. |
| `reference/product_register.csv` | Yes | Synthetic product register with harvest and grazing withholding periods. |
| `reference/paddocks.geojson` | Yes | Synthetic paddock boundaries. |

## Rough edges built in

- The contractor export covers only four of the ten events, so combining sources means de-duplicating rather than adding.
- Paddocks have names in two sources and codes in the third.
- Areas are in acres, and product and water are job totals rather than rates, in the contractor export.
- Wind is in knots for the aerial job and km/h for ground jobs, with no unit given.
- Rate units vary (L/ha, mL/ha, g/ha), and the spreadsheet leaves out "per hectare".
- One spot spray covered 12 ha of an 85 ha paddock, so totals must use the area treated.
- Crop stage uses Zadoks for cereals, BBCH for canola, and words for other timings.
- There are harvest and grazing withholding periods, with two harvest breaches to find.
- Times are local to Western Australia (UTC+08:00) and no export states a timezone.

## Ground truth and generator

The ground truth (expected answers and the underlying records) and the deterministic generator are held privately by More Than Machines, so that agents under test cannot read the answers. They are not in this repository. Scores are produced against them separately. The downstream tasks that use this fixture set are defined in `industries/grains/tasks/`.
