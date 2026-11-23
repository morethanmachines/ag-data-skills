package fingerprint

import "testing"

func cols(names ...string) []Column {
	var c []Column
	for _, n := range names {
		c = append(c, Column{Name: n, Type: "string"})
	}
	return c
}

func TestNormalise(t *testing.T) {
	cases := map[string]string{
		"Temp (°C)":         "temp_c",
		"RH (%)":            "rh_pct",
		"APVMA No.":         "apvma_no",
		"  Rate /100L ":     "rate_100l",
		"Water Rate (L/ha)": "water_rate_l_ha",
	}
	for in, want := range cases {
		if got := Normalise(in); got != want {
			t.Errorf("Normalise(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIDIgnoresOrderCaseAndOptional(t *testing.T) {
	a := Fingerprint{Columns: append(cols("Date", "Block", "Product"), Column{Name: "Notes", Optional: true})}
	b := Fingerprint{Columns: cols("product", "DATE", "block")}
	if ID(a) != ID(b) {
		t.Fatalf("ids differ: %s %s", ID(a), ID(b))
	}
	if len(ID(a)) != 15 {
		t.Fatalf("unexpected id %s", ID(a))
	}
}

func TestMatch(t *testing.T) {
	fp := Fingerprint{Columns: append(cols("Date", "Block", "Product", "Rate"), Column{Name: "Notes", Optional: true})}
	if s := Match(cols("Date", "Block", "Product", "Rate"), fp); s != 1 {
		t.Fatalf("exact match without optional column = %v, want 1", s)
	}
	if s := Match(cols("Date", "Block", "Product", "Rate", "Notes"), fp); s != 1 {
		t.Fatalf("exact match with optional column = %v, want 1", s)
	}
	if s := Match(cols("Date", "Block", "Product", "Rate", "Extra"), fp); s != 0.8 {
		t.Fatalf("one unknown column = %v, want 0.8", s)
	}
	if s := Match(cols("Date", "Block", "Product"), fp); s != 0.75 {
		t.Fatalf("one missing required column = %v, want 0.75", s)
	}
	pf := Fingerprint{Columns: []Column{{Name: "Date", Pattern: "DD/MM/YYYY"}, {Name: "Block"}}}
	if s := Match([]Column{{Name: "Date", Pattern: "YYYY-MM-DD"}, {Name: "Block"}}, pf); s != 0 {
		t.Fatalf("disagreeing pattern = %v, want 0", s)
	}
}
