---
name: setup-ag-data-skills
description: One-time setup for the Ag Data Skills pack. Use when the user installs Ag Data Skills, asks to set it up or change its settings, or asks whether their agent contributes to the farm data commons. Asks which industries they work in, whether to contribute structure (never farm data) to the open commons, and the contribution budget, then writes a small settings file the other skills read.
---

# Set up Ag Data Skills

Ask the user these questions in plain words, one message at a time if they are present. Keep it short; offer the defaults.

1. **Industries.** "Which industries do you work in?" Offer the names from `../ag-data-commons/snapshot/governance/industries.yaml` (for example wine grapes, grains, horticulture, cotton, dairy, beef). Record their ids. Several are fine (mixed farms, consultants).
2. **Contribution.** Explain in two sentences, then ask yes or no:
   "While your agent works with your farm data, it can contribute the *structure* it learns (column layouts, how a format maps to the shared spray diary, synthetic examples) to an open commons for Australian agriculture, licensed CC-BY 4.0. No farm data ever leaves your machine: no names, numbers, dates or locations from your files."
   Default: no. Contributing also needs a shell with `git` and the GitHub CLI signed in; if those are missing, record their choice anyway and say contributions will start when they are available.
3. **Budget** (only if they said yes). "How much background work is OK per session?" Offer: light (sightings only, about a minute), normal (sightings and up to one crosswalk or fix, and the reviews it owes), or a number of minutes.
4. **Organisation** (only if they said yes). The GitHub organisation or user that stands behind their contributions.

## Write the settings

Where you can write files, write `~/.config/ag-data-skills/settings.yaml`:

```yaml
settings_version: 1
industries: [viticulture]
contribute: false          # true only if the user said yes in this session
contribution_consent_date: null   # YYYY-MM-DD when they said yes
budget: light              # light | normal | <minutes>
organisation: null
```

Show the user the file you wrote, in one line. Where you cannot write files (many chat apps), keep the answers for this conversation, and tell the user that the other skills will ask again in a new conversation unless their app remembers it.

Never set `contribute: true` unless the user said yes in this session. If you are running without a user present, write `contribute: false`.
