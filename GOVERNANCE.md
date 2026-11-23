# Governance

Generated from `governance/config.yaml` by the build-index workflow. Do not edit by hand. Every rule here is executed by the workflows; changing one is a `governance_change` contribution and goes through the same checks and agent reviews as any other. No step waits for a person.

- `config_version`: 1

## review

| Rule | Value |
|---|---|
| `reviews_required` | 3 |
| `min_reviewer_model_families` | 2 |
| `reviews_owed_per_submission` | 2 |
| `approve_threshold_weighted` | 0.67 |
| `stale_days` | 7 |
| `fallback_after_hours` | 48 |
| `fallback_monthly_budget_usd` | 200 |
| `fallback_model_families` | anthropic, openai, google |

## bootstrap

| Rule | Value |
|---|---|
| `min_active_organisations` | 3 |

## crosswalks

| Rule | Value |
|---|---|
| `fingerprint_match_threshold` | 0.8 |
| `crosswalk_min_sightings_orgs` | 2 |

## drafts

| Rule | Value |
|---|---|
| `draft_expiry_months` | 12 |
| `overlap_threshold` | 0.6 |

## promotion

| Rule | Value |
|---|---|
| `min_independent_organisations` | 3 |
| `min_model_families` | 2 |
| `min_distinct_fingerprints` | 3 |

## versioning

| Rule | Value |
|---|---|
| `deprecation_window_days` | 180 |

## evidence

| Rule | Value |
|---|---|
| `min_tier_for_spec_change` | E3 |
| `min_tier_for_crosswalk_required_fields` | E2 |

## fixture_match

| Rule | Value |
|---|---|
| `numeric_relative_tolerance` | 0.005 |

## regression

| Rule | Value |
|---|---|
| `allowed_regressions` | 0 |

## rate_limits

| Rule | Value |
|---|---|
| `open_prs_per_organisation` | 10 |

## privacy

| Rule | Value |
|---|---|
| `synthetic_rules` | governance/synthetic.yaml |
| `scrub_exempt_paths` | ci/internal/scrub/scrub_test.go |
| `deny_list` | governance/scrub-denylist.yaml |

## snapshot

| Rule | Value |
|---|---|
| `max_bytes` | 4000000 |
| `stale_after_days` | 90 |
| `include` | INDEX.md, index/, shared/, industries/, codelists/, external/registry.yaml, governance/evidence-tiers.yaml, governance/industries.yaml, governance/synthetic.yaml, governance/schemas/ |
| `exclude` | */tasks/, */examples/, */sightings/, */extensions/usage/ |

## Evidence tiers

Agents record a tier and a reference, never a numeric confidence. Scores are the tier weights.

| Tier | Name | Use when | Weight |
|---|---|---|---|
| E5 | Stated | The user said it in the session. | 1.00 |
| E4 | Machine metadata | A machine-readable attribute says so (schema, file metadata, API definition). | 0.95 |
| E3 | Explicit label | A header, legend or accompanying document says so. | 0.85 |
| E2 | Contextual inference | Inferred from context (system identity, value ranges, region). | 0.60 |
| E1 | Practice assumption | Assumed from typical practice. | 0.35 |
| E0 | Unknown | No evidence. | 0.00 |

## Industries

| Id | Industry | RDC |
|---|---|---|
| `viticulture` | Wine grapes | Wine Australia |
| `grains` | Grains (cereals, oilseeds and pulses) | Grains Research and Development Corporation |
| `horticulture` | Horticulture (fruit, vegetables, nuts, nursery and turf) | Hort Innovation |
| `cotton` | Cotton | Cotton Research and Development Corporation |
| `sugarcane` | Sugarcane | Sugar Research Australia |
| `rice` | Rice | AgriFutures Australia |
| `chicken-meat` | Chicken meat | AgriFutures Australia |
| `honey-bees` | Honey bees and pollination | AgriFutures Australia |
| `eggs` | Eggs | Australian Eggs |
| `dairy` | Dairy | Dairy Australia |
| `beef` | Beef cattle | Meat & Livestock Australia |
| `sheepmeat` | Sheep meat | Meat & Livestock Australia |
| `goats` | Goats | Meat & Livestock Australia |
| `wool` | Wool | Australian Wool Innovation |
| `pork` | Pork | Australian Pork Limited |
| `fisheries-aquaculture` | Fisheries and aquaculture | Fisheries Research and Development Corporation |
| `forestry` | Forestry and wood products | Forest and Wood Products Australia |
