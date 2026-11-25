# Acceptance tests

*2026-11-23, tag v0.1.0. Scores only: the agent's answers would reveal the hidden expected answers, so they are not reproduced here.*

| Criterion | How it was tested | Result |
|---|---|---|
| Specs, profile and code lists validate in CI | `checks` workflow on every push to main runs `go -C ci run . checks` (schemas, x-cp annotations, profile constraint rules, code lists, examples) | Pass |
| Scrub unit tests pass | `go -C ci test ./...` in the `checks` workflow, including property tests (2,000 generated checksum-valid ABNs always rejected, 2,000 failing ones always accepted; likewise ACNs and coordinates) | Pass |
| A test PR containing a valid ABN is closed automatically | PR #1 added a CSV with the Australian Business Register's worked-example ABN. The `scrub` workflow commented with the value masked, closed the PR and deleted its branch; the required `checks` job also failed | Pass |
| The pack installs with `npx skills add` | `npx skills@latest add morethanmachines/ag-data-skills --skill '*' --agent claude-code -y --copy` in an empty project: six skills installed, snapshot included | Pass |
| The pack installs as a Claude Code plugin | `claude plugin marketplace add morethanmachines/ag-data-skills` then `claude plugin install ag-data-skills@ag-data-skills` in an isolated config: six skills listed; `claude plugin validate .` passes | Pass |
| A fresh agent with the pack and no network can map the spray-diary-01 exports from the snapshot and answer the three tasks | Headless Claude Code (claude-opus-5-5), pack built from the v0.1.0 release zips, Scenario 01 variant B prompt verbatim, working folder with only `exports/` and `reference/`. Web tools denied, Bash sandboxed with no network, the repository, transcripts and private ground truth unreadable. Scored against the private expected answers | Pass: see below |

## Offline agent run

| Measure | Result |
|---|---|
| Skills triggered unprompted | combine-spray-records, ag-data-commons, check-withholding, convert-spray-records |
| Read from the bundled snapshot | root, viticulture and shared indexes; viticulture profile; spray diary README and KNOWN_ISSUES; evidence tiers; registry |
| Web or network calls | 0 |
| Mapped documents | one per export plus a combined set; combined set validates against the viticulture profile |
| Task 1 | 28 of 28 block and product totals within 1% |
| Task 2 | both breaches found, no false positives |
| Task 3 | 49 rows; 85 of 85 product rows match on date, block, product, rate, unit and water |
| Consistency | the agent reported that all three exports give the same amounts within 1% |
| Cost | 27 turns, about 4 minutes |

One run with one model family. A baseline across model families is tracked in issue #7.

## Rerun with v0.2 (one skill)

*2026-11-24.* After the pack was replaced by the single `ag-data-commons` skill, the offline run was repeated with the same prompt, the same restrictions and a fresh pack built from the v0.2.0 release zip, so only that one skill was installed.

| Measure | Result |
|---|---|
| Skill triggered unprompted | ag-data-commons (the prompt does not name it) |
| Web or network calls | 0; no permission denials |
| Mapped documents | one per export; each validates against the viticulture profile |
| Task 1 | 28 of 28 block and product totals within 1%, units correct |
| Task 2 | both breaches found, no false positives |
| Task 3 | 49 rows; 85 of 85 product rows match on date, block, product, rate, unit and water |
| Consistency | the agent reported that all three exports agree on every application, and that the converted file duplicates the existing GrapeLog export |
| Cost | 19 turns, about 3 minutes |

Again one run with one model family.
