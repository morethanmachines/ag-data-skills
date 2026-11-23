# 0001: Decisions taken while building Phase 0

*Date: 2026-11-23. Recorded under the implementation plan's rule that a deviation from a
SHOULD, or a resolution of something the plan leaves open, is written down with a reason.*

1. **No verbatim copy of the reference spec, and no variety or E-L code lists yet.** The
   reference repository has no licence (plan decision 15.3 is open). The plan asks for the
   original schema to be kept under `shared/specs/spray-diary/reference/` and for variety
   and E-L code lists. Both are held back until reuse terms are confirmed. The shared spec
   follows the reference loosely; the viticulture profile constrains E-L stages to integers
   1 to 47 and leaves variety as free text. Tracked in an issue and in `KNOWN_ISSUES.md`.
2. **Bootstrap decision records are ids `bootstrap-NNNN`, not PR numbers,** because the
   seeds were committed directly, before any merge workflow existed. They state that no
   reviews took place: the review workflows arrive in Phase 2.
3. **No crosswalks are seeded.** Phases 1 and 2 require crosswalks to come from agents
   through the contribution loop, so Phase 0 leaves `crosswalks/` empty and the acceptance
   test maps the fixtures directly to the spec.
4. **Fixtures carry exports and reference files only.** Expected mapped outputs only exist
   once crosswalks do; expected task answers stay in the private store.
5. **Downstream tasks are a question, a method and an answer schema.** The plan does not
   define a query language. Each task's `method` is a list of steps over spec properties,
   and `answer.schema.json` fixes the shape of the answer. The checks workflow validates
   task definitions; executing tasks against expected answers needs the hidden store and is
   deferred (issue).
6. **Industries list.** "Aligned to the 15 rural RDC industry groupings" is read as: farm
   production industries covered by the 15 RDCs, each tagged with its RDC. Post-farm-gate
   remits (meat processing, live export) are not industries in this sense.
7. **Units are UCUM codes** (`har` for hectare), with a `codelists/units` list giving plain
   labels. Rate basis is a fixed list with fixed denominators.
8. **Times are local wall-clock plus optional IANA timezone and offset,** because every
   fixture export omits the timezone and a mapped record must be able to say so.
9. **Generated files are committed by workflows** (`build-index`, `release`) using the
   workflow token, which needs a bypass in the branch ruleset for GitHub Actions. The
   ruleset also keeps a repository admin bypass until the Phase 2 merge workflow exists,
   because otherwise nothing could be merged at all; an issue tracks removing it.
10. **Scrub uses `pull_request_target`** so it can close pull requests from forks. It runs
    only base-branch code and rules and reads the pull request as data. Values in its
    findings are masked so its comment never repeats leaked data. One path is exempt
    (`ci/internal/scrub/scrub_test.go`, which holds deliberate real-looking identifiers).
11. **Deny-list markers and unlabelled PIC and NLIS patterns apply to data files only**
    (CSV, JSON, YAML, xlsx and similar outside `governance/`), because prose such as this
    document mentions them.
12. **Skill v0.1 contribution** opens pull requests that the `checks` and `scrub` workflows
    test; they wait for the Phase 2 review and merge workflows before they can merge.
