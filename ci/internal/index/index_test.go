package index

import (
	"bytes"
	"strings"
	"testing"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

func TestBuildIsDeterministicAndComplete(t *testing.T) {
	r, err := repo.Find(".")
	if err != nil {
		t.Fatal(err)
	}
	a, err := Build(r)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := Build(r)
	for p := range a {
		if !bytes.Equal(a[p], b[p]) {
			t.Fatalf("%s differs between builds", p)
		}
	}
	for _, p := range []string{"INDEX.md", "shared/INDEX.md", "index/shared.json", "industries/viticulture/INDEX.md", "index/viticulture.json", "GOVERNANCE.md", "shared/specs/spray-diary/CHANGELOG.md"} {
		if _, ok := a[p]; !ok {
			t.Errorf("missing generated file %s", p)
		}
	}
	vit := string(a["industries/viticulture/INDEX.md"])
	for _, want := range []string{"viticulture/vineyard-block", "viticulture/spray-diary", "profile of shared/spray-diary"} {
		if !strings.Contains(vit, want) {
			t.Errorf("viticulture index lacks %q", want)
		}
	}
	if !strings.Contains(string(a["GOVERNANCE.md"]), "| `reviews_required` | 3 |") {
		t.Error("GOVERNANCE.md lacks the review thresholds")
	}
}

func TestSnapshotExcludesFixturesTasksAndRecords(t *testing.T) {
	r, _ := repo.Find(".")
	files, err := SnapshotFiles(r)
	if err != nil {
		t.Fatal(err)
	}
	var sawSpec bool
	for _, f := range files {
		if strings.HasPrefix(f, "fixtures/") || strings.Contains(f, "/tasks/") || strings.HasPrefix(f, "records/") || strings.Contains(f, "/examples/") {
			t.Errorf("snapshot must not include %s", f)
		}
		sawSpec = sawSpec || f == "shared/specs/spray-diary/schema.json"
	}
	if !sawSpec {
		t.Error("snapshot lacks the shared spray diary schema")
	}
}

func TestExcluded(t *testing.T) {
	pats := []string{"*/tasks/", "*/examples/"}
	if !excluded("industries/grains/tasks/x/task.yaml", pats) || excluded("industries/grains/profiles/x.json", pats) {
		t.Fatal("exclusion patterns misapplied")
	}
}
