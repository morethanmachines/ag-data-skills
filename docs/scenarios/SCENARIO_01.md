# Scenario 01: three spray diaries, one season

**Dataset:** `test-data/spray-diary-01`
**Spec under test:** the reference spray diary spec (Wine Australia, 2023), at `https://github.com/morethanmachines/project-harmony`

## Purpose

This scenario checks the core loop. An agent is given messy data as it would find it in the wild, is asked to do ordinary work with it, and in doing so maps the data to a common spec and notices where the spec falls short. Nothing more.

There are two variants. Run variant A now, before any the Ag Data Skills pack exists, to get a baseline. Run variant B once the skill and the commons repository are in place. The difference between the two is the first measure of whether the commons adds anything.

## Setup

Give the agent a working folder containing only `exports/` and `reference/` from the dataset. (The ground truth and generator are held privately and are not in this repository.) Use a fresh session with no other context. The agent may use the web and may clone the reference spec repository.

## Variant A: baseline prompt (without the Ag Data Skills pack)

> I'm the agronomist for Synthetic Ridge Vineyard. This season's spray records are in three places: the VineDiary app export, the wine company's GrapeLog export and the grower's own spreadsheet. They're all in `exports/`. The block list, the product register and the wine company's planned harvest dates are in `reference/`.
>
> I need three things:
>
> 1. The total of each product applied to each block this vintage, using the wine company's block codes.
> 2. Whether any block will breach a withholding period at its planned harvest date.
> 3. The VineDiary records converted into GrapeLog's layout so the wine company can load them.
>
> Before you start, map everything to the common spray diary spec Wine Australia published (at https://github.com/morethanmachines/project-harmony) and work from the mapped data. If the spec can't hold something you need, don't just quietly work around it. Write down what's missing or wrong and draft the change you'd propose. For every field you map, note what evidence you used and how sure you are.
>
> Save the mapped data, your answers (tasks 1 and 2 as JSON as well as a short summary), the converted GrapeLog file and your proposed spec changes as files.

## Variant B: with the Ag Data Skills pack

Same setup, with the Ag Data Skills pack installed and the commons repository containing the shared spray diary spec and viticulture profile, plus crosswalks if any have been merged. The agent needs a shell with git.

> I'm the agronomist for Synthetic Ridge Vineyard. This season's spray records are in three places: the VineDiary app export, the wine company's GrapeLog export and the grower's own spreadsheet. They're all in `exports/`. The block list, the product register and the wine company's planned harvest dates are in `reference/`.
>
> I need three things:
>
> 1. The total of each product applied to each block this vintage, using the wine company's block codes.
> 2. Whether any block will breach a withholding period at its planned harvest date.
> 3. The VineDiary records converted into GrapeLog's layout so the wine company can load them.
>
> Save your answers (tasks 1 and 2 as JSON as well as a short summary) and the converted GrapeLog file.

Variant B deliberately says nothing about specs, mapping or evidence. The test is whether the skill makes the agent do all of that on its own.

## Scoring

Score against the privately held `ground_truth/expected_answers.json` for the dataset. Scoring should be automated from the agent's output files.

| Measure | How |
|---|---|
| Task 1 accuracy | Share of block and product totals within 1% of expected (28 values) |
| Task 2 accuracy | Both breaches found (Synth Myclo 200 EW on WC-102, Synth Phos 600 on WC-101), no false positives |
| Task 3 accuracy | Converted rows match `exports/grapelog_export_2025-26.csv` on date, block, product, rate per ha (within 0.5%), unit and total water (within 1 L) |
| Rough edges noticed | How many of the ten listed in the expected answers the agent explicitly identified |
| Spec gaps proposed | Overlap between the agent's proposed changes and the spec's starting issues (section 13 of the implementation plan), plus any new, valid gaps |
| Evidence discipline | Share of mapped fields with an evidence reference; whether weak inferences (such as wind being m/s) were flagged as such |
| Consistency | Whether the three exports, once mapped, give the same answers |
| Cost | Tokens and wall-clock time |

For variant B only, also check whether the skill triggered unprompted, whether a crosswalk was produced for each export format, and whether the evidence tiers recorded match what the evidence actually supports.

## Running it across models

Run each variant with at least two model families and keep the full session logs. The logs are research data. They show where agents struggle with real-looking farm data, and they become the first evidence for whether the commons converges.
