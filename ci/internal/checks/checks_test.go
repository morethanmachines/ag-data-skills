package checks

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/morethanmachines/ag-data-skills/ci/internal/fingerprint"
	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

// copyRepo copies the real repository (without .git, private/ and generated output) to a
// temporary directory so a test can break it.
func copyRepo(t *testing.T) string {
	t.Helper()
	src, err := repo.Find(".")
	if err != nil {
		t.Fatal(err)
	}
	dst := t.TempDir()
	err = filepath.WalkDir(src.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src.Root, p)
		if rel == "." {
			return nil
		}
		if repo.IsSkipped(filepath.ToSlash(rel)) || strings.HasPrefix(rel, "ci") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, b, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
	return dst
}

func run(t *testing.T, root string) []Finding {
	t.Helper()
	r, err := repo.Find(root)
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(r)
	if err != nil {
		t.Fatal(err)
	}
	return c.RunAll()
}

func has(fs []Finding, check, substr string) bool {
	for _, f := range fs {
		if f.Check == check && strings.Contains(f.Path+" "+f.Msg, substr) {
			return true
		}
	}
	return false
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, rel)
	os.MkdirAll(filepath.Dir(p), 0o755)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, root, rel string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func writeJSON(t *testing.T, root, rel string, v any) {
	t.Helper()
	b, _ := json.MarshalIndent(v, "", "  ")
	write(t, root, rel, string(b))
}

func TestRealRepositoryPasses(t *testing.T) {
	r, err := repo.Find(".")
	if err != nil {
		t.Fatal(err)
	}
	if f := run(t, r.Root); len(f) != 0 {
		t.Fatalf("the repository should pass its own checks: %v", f)
	}
}

const profile = "industries/viticulture/profiles/spray-diary.json"

func mutateProfile(t *testing.T, fn func(records map[string]any)) []Finding {
	root := copyRepo(t)
	p := readJSON(t, root, profile)
	items := p["properties"].(map[string]any)["records"].(map[string]any)["items"].(map[string]any)
	fn(items["properties"].(map[string]any))
	writeJSON(t, root, profile, p)
	return run(t, root)
}

func TestProfileCannotIntroduceProperties(t *testing.T) {
	f := mutateProfile(t, func(props map[string]any) {
		props["vineyard_name"] = map[string]any{"type": "string"}
	})
	if !has(f, "profile", "vineyard_name") {
		t.Fatalf("expected a profile finding: %v", f)
	}
}

func TestProfileCannotChangeTypes(t *testing.T) {
	f := mutateProfile(t, func(props map[string]any) {
		props["record_id"] = map[string]any{"type": "integer"}
	})
	if !has(f, "profile", "type \"integer\"") {
		t.Fatalf("expected a type finding: %v", f)
	}
}

func TestProfileMayNarrowNumberToInteger(t *testing.T) {
	f := mutateProfile(t, func(props map[string]any) {
		props["crop_stage"].(map[string]any)["properties"].(map[string]any)["value"] = map[string]any{"type": "integer", "minimum": 1}
	})
	if has(f, "profile", "") {
		t.Fatalf("narrowing should be allowed: %v", f)
	}
}

func TestProfileCannotUseOtherKeywords(t *testing.T) {
	f := mutateProfile(t, func(props map[string]any) {
		props["season"].(map[string]any)["additionalProperties"] = true
	})
	if !has(f, "profile", "additionalProperties") {
		t.Fatalf("expected a keyword finding: %v", f)
	}
}

func TestProfileMustReferenceSharedSpec(t *testing.T) {
	root := copyRepo(t)
	p := readJSON(t, root, profile)
	p["allOf"] = []any{map[string]any{"type": "object"}}
	writeJSON(t, root, profile, p)
	if f := run(t, root); !has(f, "profile", "allOf") {
		t.Fatalf("expected an allOf finding: %v", f)
	}
}

func TestExampleMustValidate(t *testing.T) {
	root := copyRepo(t)
	ex := "shared/specs/spray-diary/examples/viticulture-example.json"
	p := readJSON(t, root, ex)
	rec := p["records"].([]any)[0].(map[string]any)
	delete(rec, "management_units")
	writeJSON(t, root, ex, p)
	if f := run(t, root); !has(f, "example", "management_units") {
		t.Fatalf("expected an example finding: %v", f)
	}
}

func TestProfileRulesApplyToExamples(t *testing.T) {
	root := copyRepo(t)
	ex := "shared/specs/spray-diary/examples/viticulture-example.json"
	p := readJSON(t, root, ex)
	rec := p["records"].([]any)[0].(map[string]any)
	rec["crop_stage"] = map[string]any{"scale": "Zadoks", "value": 23}
	writeJSON(t, root, ex, p)
	f := run(t, root)
	if !has(f, "example", "profiles/spray-diary.json") {
		t.Fatalf("a Zadoks stage should fail the viticulture profile: %v", f)
	}
}

func TestNumericConfidenceRejected(t *testing.T) {
	root := copyRepo(t)
	write(t, root, "industries/viticulture/sightings/x.yaml", "mapping:\n  confidence: 0.9\n")
	if f := run(t, root); !has(f, "numeric-confidence", "confidence") {
		t.Fatalf("expected a numeric-confidence finding: %v", f)
	}
}

func TestCodelistReferenceMustExist(t *testing.T) {
	root := copyRepo(t)
	sp := "industries/viticulture/specs/vineyard-block/schema.json"
	s := readJSON(t, root, sp)
	s["$defs"].(map[string]any)["block"].(map[string]any)["properties"].(map[string]any)["rootstock"] = map[string]any{"type": "string", "x-cp-codelist": "codelists/rootstocks"}
	writeJSON(t, root, sp, s)
	if f := run(t, root); !has(f, "codelist-ref", "rootstocks") {
		t.Fatalf("expected a codelist-ref finding: %v", f)
	}
}

func TestSpecStatusFollowsVersion(t *testing.T) {
	root := copyRepo(t)
	sp := "shared/specs/spray-diary/schema.json"
	s := readJSON(t, root, sp)
	s["x-cp-status"] = "active"
	writeJSON(t, root, sp, s)
	if f := run(t, root); !has(f, "spec", "draft") {
		t.Fatalf("expected a status finding: %v", f)
	}
}

func TestUnknownIndustryFolder(t *testing.T) {
	root := copyRepo(t)
	write(t, root, "industries/wine/INDEX.md", "x")
	if f := run(t, root); !has(f, "industries", "wine") {
		t.Fatalf("expected an industries finding: %v", f)
	}
}

func TestGroundTruthNeverCommitted(t *testing.T) {
	root := copyRepo(t)
	write(t, root, "fixtures/spray-diary-01/ground_truth/expected_answers.json", "{}")
	write(t, root, "industries/viticulture/tasks/withholding-at-harvest/expected.json", "{}")
	f := run(t, root)
	if !has(f, "hidden-fixtures", "ground_truth") || !has(f, "hidden-fixtures", "expected.json") {
		t.Fatalf("expected hidden-fixtures findings: %v", f)
	}
}

const fpJSON = `{"fingerprint_version": 1, "media_type": "text/csv", "layout": {"delimiter": ",", "header_rows": 1},
 "columns": [{"name": "job", "type": "string"}, {"name": "date", "type": "date", "pattern": "YYYY-MM-DD"},
  {"name": "block", "type": "string"}, {"name": "product", "type": "string"}, {"name": "rate_l_ha", "type": "number"}]}`

func crosswalkYAML(tier string) string {
	return `crosswalk_version: 1
source:
  fingerprint: fingerprint.json
target:
  spec: shared/specs/spray-diary
  profile: industries/viticulture/profiles/spray-diary.json
  version: ">=0.1.0 <0.2.0"
record_rule: "One record per job."
fields:
  - target: records[].record_id
    from: job
    instruction: "Copy as is."
    evidence: { tier: E3, ref: "column header job" }
  - target: records[].products_applied[].rate
    from: rate_l_ha
    instruction: "Rate per hectare in L."
    evidence: { tier: ` + tier + `, ref: "header" }
fixtures:
  input: ../../../../fixtures/test-cw/input.csv
  expected_output: ../../../../fixtures/test-cw/expected/out.mapped.json
`
}

func addCrosswalk(t *testing.T, root, dirName, tier string, expected string) {
	dir := "industries/viticulture/crosswalks/" + dirName
	write(t, root, dir+"/fingerprint.json", fpJSON)
	write(t, root, dir+"/crosswalk.yaml", crosswalkYAML(tier))
	write(t, root, "fixtures/test-cw/README.md", "Synthetic.")
	write(t, root, "fixtures/test-cw/input.csv", "job,date,block,product,rate_l_ha\n")
	write(t, root, "fixtures/test-cw/expected/out.mapped.json", expected)
}

const goodOutput = `{"spec": "shared/spray-diary", "spec_version": "0.1.0", "profile": "viticulture/spray-diary", "records": [
 {"record_id": "J1", "start": {"local": "2025-10-01"}, "season": {"scheme": "vintage", "label": "2026", "year": 2026},
  "management_units": [{"unit_type": "vineyard_block", "id": "B1", "id_scheme": "grower"}],
  "products_applied": [{"product_name": "Synth X", "rate": {"value": 1, "unit": "L", "basis": "per_hectare"}}]}]}`

func fpID(t *testing.T, _ string) string {
	t.Helper()
	var fp fingerprint.Fingerprint
	if err := json.Unmarshal([]byte(fpJSON), &fp); err != nil {
		t.Fatal(err)
	}
	return fingerprint.ID(fp)
}

func TestCrosswalkChecks(t *testing.T) {
	root := copyRepo(t)
	id := fpID(t, root)
	addCrosswalk(t, root, id, "E2", goodOutput)
	if f := run(t, root); len(f) != 0 {
		t.Fatalf("a good crosswalk should pass: %v", f)
	}

	root = copyRepo(t)
	addCrosswalk(t, root, id, "E1", goodOutput)
	if f := run(t, root); !has(f, "evidence", "E1") {
		t.Fatalf("expected an evidence finding: %v", f)
	}

	root = copyRepo(t)
	addCrosswalk(t, root, "fp-000000000000", "E2", goodOutput)
	if f := run(t, root); !has(f, "crosswalk", "fingerprint id") {
		t.Fatalf("expected a folder name finding: %v", f)
	}

	root = copyRepo(t)
	addCrosswalk(t, root, id, "E2", strings.Replace(goodOutput, `"vintage"`, `"crop_year"`, 1))
	if f := run(t, root); !has(f, "regression", "profiles/spray-diary.json") {
		t.Fatalf("expected a regression finding against the profile: %v", f)
	}
}

func gitInit(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{{"init", "-q", "-b", "main"}, {"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "base"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
}

func gitCommit(t *testing.T, root string) {
	t.Helper()
	for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "change"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
}

const manifest = "```yaml\ncp-contribution:\n  type: spec_new\n  model: anthropic/claude-opus-5-5\n  harness: claude-code\n  skill_version: 0.1.0\n  organisation: example-org\n  summary: \"A grains spray spec\"\n```\n"

func runPR(t *testing.T, root, body string) []Finding {
	r, _ := repo.Find(root)
	c, err := New(r)
	if err != nil {
		t.Fatal(err)
	}
	c.RunAll()
	return c.RunPR("HEAD~1", body)
}

func TestSpecNewOverlapAndExternalSearch(t *testing.T) {
	root := copyRepo(t)
	gitInit(t, root)
	s := readJSON(t, root, "shared/specs/spray-diary/schema.json")
	s["$id"] = repo.BaseURL + "industries/grains/specs/spray-records/schema.json"
	s["x-cp-id"] = "grains/spray-records"
	s["properties"].(map[string]any)["spec"] = map[string]any{"const": "grains/spray-records"}
	writeJSON(t, root, "industries/grains/specs/spray-records/schema.json", s)
	write(t, root, "industries/grains/specs/spray-records/KNOWN_ISSUES.md", "x")
	gitCommit(t, root)
	f := runPR(t, root, manifest)
	if !has(f, "overlap", "shared/spray-diary") {
		t.Fatalf("expected an overlap finding naming shared/spray-diary: %v", f)
	}
	if !has(f, "external-search", "external-search.yaml") {
		t.Fatalf("expected an external-search finding: %v", f)
	}
}

func TestDistinctSpecNewPassesOverlap(t *testing.T) {
	root := copyRepo(t)
	gitInit(t, root)
	write(t, root, "industries/grains/specs/soil-test/schema.json", `{
 "$schema": "https://json-schema.org/draft/2020-12/schema",
 "$id": "`+repo.BaseURL+`industries/grains/specs/soil-test/schema.json",
 "x-cp-id": "grains/soil-test", "x-cp-version": "0.1.0", "x-cp-status": "draft",
 "x-cp-external-basis": [], "x-cp-provenance": ["bootstrap-0001"],
 "type": "object", "properties": {"samples": {"type": "array", "items": {"type": "object", "properties": {
   "sample_depth_cm": {"type": "number"}, "phosphorus_colwell": {"type": "number"}, "lab_method": {"type": "string"}}}}}}`)
	write(t, root, "industries/grains/specs/soil-test/KNOWN_ISSUES.md", "x")
	write(t, root, "industries/grains/specs/soil-test/external-search.yaml", "searches: []\n")
	gitCommit(t, root)
	f := runPR(t, root, manifest)
	if len(f) != 0 {
		t.Fatalf("a distinct draft with a search record should pass: %v", f)
	}
}

func TestManifestRequired(t *testing.T) {
	root := copyRepo(t)
	gitInit(t, root)
	write(t, root, "industries/viticulture/sightings/s.yaml", "")
	os.Remove(filepath.Join(root, "industries/viticulture/sightings/s.yaml"))
	write(t, root, "README.md", "changed")
	gitCommit(t, root)
	if f := runPR(t, root, "no manifest here"); !has(f, "manifest", "cp-contribution") {
		t.Fatalf("expected a manifest finding: %v", f)
	}
	bad := strings.Replace(manifest, "anthropic/claude-opus-5-5", "not a model id", 1)
	if f := runPR(t, root, bad); !has(f, "manifest", "invalid") {
		t.Fatalf("expected an invalid manifest finding: %v", f)
	}
}

func TestFamilyOf(t *testing.T) {
	r, _ := repo.Find(".")
	cases := map[string]string{
		"anthropic/claude-opus-5-5": "anthropic",
		"openai/gpt-5":              "openai",
		"google/gemini-3-pro":       "google",
		"meta/llama-4":              "meta",
		"someone/unknown":           "other",
	}
	for m, want := range cases {
		if got := FamilyOf(r, m); got != want {
			t.Errorf("FamilyOf(%q) = %q, want %q", m, got, want)
		}
	}
}
