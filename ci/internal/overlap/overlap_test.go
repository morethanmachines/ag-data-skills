package overlap

import (
	"testing"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

func props(t *testing.T, js string) map[string]string {
	v, err := repo.DecodeJSON([]byte(js))
	if err != nil {
		t.Fatal(err)
	}
	return Props(v)
}

func TestOverlap(t *testing.T) {
	spray := props(t, `{"properties":{"records":{"items":{"properties":{"session_start":{},"products_applied":{"items":{"properties":{"product_name":{},"rate":{"x-cp-kind":"quantity"},"registration":{}}}},"management_units":{},"name":{}}}}}}`)
	dup := props(t, `{"properties":{"applications":{"items":{"properties":{"session_start":{},"product_name":{},"rate":{"x-cp-kind":"quantity"},"registration":{},"management_units":{}}}}}}`)
	soil := props(t, `{"properties":{"samples":{"items":{"properties":{"sample_depth":{},"phosphorus_colwell":{},"lab_method":{},"name":{}}}}}}`)
	if _, ok := spray["name"]; ok {
		t.Fatal("structural names should be ignored")
	}
	if s := Score(dup, spray); s < 0.6 {
		t.Fatalf("duplicate draft scored %v, want >= 0.6", s)
	}
	if s := Score(soil, spray); s != 0 {
		t.Fatalf("unrelated draft scored %v, want 0", s)
	}
	r := Rank(dup, map[string]map[string]string{"shared/spray-diary": spray, "x/soil": soil})
	if r[0].Spec != "shared/spray-diary" {
		t.Fatalf("rank = %v", r)
	}
	kindDiff := props(t, `{"properties":{"rate":{"x-cp-kind":"text"}}}`)
	if s := Score(kindDiff, spray); s != 0.5 {
		t.Fatalf("kind mismatch = %v, want 0.5", s)
	}
}
