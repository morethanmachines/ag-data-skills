# Scenario 02: grains spray records against the shared spec

**Dataset:** `fixtures/spray-diary-02`
**Spec under test:** the shared spray diary spec, after Scenario 01 has produced it and its viticulture profile

## Purpose

Scenario 01 builds the shared spray diary from a viticulture starting point. This scenario checks whether it really is shared. An agent with grains data should be able to map it to the shared spec by adding a grains profile, without changing the shared spec itself. If it has to change the shared spec, that is the most useful result this scenario can produce: either the shared spec is not yet general, or the change genuinely belongs in it.

Run it after Scenario 01 variant B, against a commons that contains the shared spray diary, the viticulture profile and the spray-diary-01 crosswalks, but nothing for grains.

## Setup

Give the agent a working folder containing only `exports/` and `reference/` from the dataset. (The ground truth and generator are held privately and are not in this repository.) Fresh session, the Ag Data Skills pack installed, a shell with git, and the commons repository as above.

## Prompt

> I'm the agronomist for Synthetic Plains, a grains farm in WA. This season's spray records are in three places: the farm software export, the spray contractor's job history and the farmer's own spreadsheet. They're all in `exports/`. The paddock list, product register and planned harvest dates are in `reference/`.
>
> I need three things:
>
> 1. The total of each product applied to each paddock this season.
> 2. Whether any paddock will breach a harvest withholding period at its planned harvest date, and the earliest date stock can go onto each paddock after its last spray.
> 3. The contractor's jobs converted into the farm software's layout so they can be checked against what's already recorded there.
>
> Save your answers (tasks 1 and 2 as JSON as well as a short summary) and the converted file.

As in Scenario 01 variant B, the prompt says nothing about specs. The test is whether the skill makes the agent find the shared spec, recognise that grains needs a profile rather than a new spec, and contribute it.

## Scoring

Score against the privately held `ground_truth/expected_answers.json` for the dataset.

| Measure | How |
|---|---|
| Task 1 accuracy | Share of paddock and product totals within 1% of expected (16 values), with no double counting of contractor jobs |
| Task 2 accuracy | Both harvest breaches found (Synth Tebucon 430 on Creek, Synth Glypho 540 on Long), no false positives, and earliest grazing date correct for all four paddocks |
| Task 3 accuracy | Converted rows match the contractor rows of the farm software export (14 rows) on date, time, paddock, area (within 0.1 ha), product, rate (within 0.5%) and water |
| Shared spec untouched | Whether the agent mapped the data using a new grains profile only. Any change it proposes to the shared spec is recorded and assessed separately: is it a genuine gap in the shared spec, or something that belongs in the profile? |
| Reuse | Whether the agent found and used the shared spec rather than creating a new grains spray spec |
| Rough edges noticed | How many of the eleven listed in the expected answers the agent explicitly identified |
| Evidence discipline | Share of mapped fields with an evidence reference; whether weak inferences (wind in knots for the aerial job, per hectare in the spreadsheet) were flagged as such |
| Contributions | Sightings for all three shapes; a grains profile; no crosswalks (each shape has only been seen once) unless a challenge asked for them |
| Cost | Tokens and wall-clock time |
