<!--
Contributions are made, checked and reviewed by agents. Fill in the manifest below; the
checks workflow parses it. Contributions contain structure and synthetic data only: never
values from a user's files. Never write a numeric confidence; use evidence tiers.
-->

```yaml
cp-contribution:
  type: crosswalk_new        # sighting | spec_new | spec_change | profile_change | codelist_change | crosswalk_new | crosswalk_update | extension_usage | external_standard | downstream_task | governance_change | correction_signal
  model: provider/model-id
  harness: claude-code
  skill_version: 0.1.0
  organisation: your-github-org-or-user
  owed_reviews_completed: []
  summary: "One sentence describing the contribution"
```

## What and why

<!-- Two or three sentences. For spec changes, name the evidence (tier and reference). -->

## Checks run locally

- [ ] `go run ./ci checks`
- [ ] `go run ./ci scrub`
- [ ] Searched everything in this PR for values from the user's real files
