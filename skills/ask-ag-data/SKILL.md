---
name: ask-ag-data
description: Router for the Ag Data Skills pack. Use when the user has Australian farm data (spray records, chemical application diaries, contractor job sheets, paddock or block lists, product registers, harvest dates) and is not sure what to do with it or which skill to use, or asks what Ag Data Skills can do.
---

# Which Ag Data Skills skill?

Describe the user's situation back to them in one or two sentences (what data they have, from how many systems, and what they seem to need), then point to the skill that fits and offer to start it:

| The user wants to... | Skill |
|---|---|
| Put spray records from several apps, contractors or spreadsheets into one clean list, or total what was applied where | `combine-spray-records` |
| Check withholding periods (harvest, grazing) against planned harvest or grazing dates | `check-withholding` |
| Turn spray records into another system's layout, or a buyer's or auditor's format | `convert-spray-records` |
| Set up the pack, or change whether their agent contributes to the commons | `setup-ag-data-skills` |
| Anything else with farm data (cleaning, mapping, querying a file) | `ag-data-commons` directly |

If the job needs more than one (for example combine, then check withholding), say so and run them in order. If the data is not spray records, say that the commons starts with spray records, and use `ag-data-commons`, which looks up whether a spec exists for the data and maps it with evidence anyway.

If the settings file `~/.config/ag-data-skills/settings.yaml` does not exist and the user seems to be setting up, suggest `setup-ag-data-skills` once.
