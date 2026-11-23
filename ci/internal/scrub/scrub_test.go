package scrub

import (
	"archive/zip"
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

func testScanner() *Scanner {
	return &Scanner{
		Areas: []repo.TestArea{
			{ID: "gulf-st-vincent", MinLon: 137.95, MaxLon: 138.25, MinLat: -35.20, MaxLat: -34.85},
			{ID: "indian-ocean-geraldton", MinLon: 113.50, MaxLon: 114.30, MinLat: -29.20, MaxLat: -28.40},
		},
		Deny: []string{"Pty Ltd", "Tax File Number"},
	}
}

func kinds(fs []Finding) map[string]int {
	m := map[string]int{}
	for _, f := range fs {
		m[f.Kind]++
	}
	return m
}

// randomABN returns an 11-digit string that passes (want=true) or fails the ABN checksum.
func randomABN(rng *rand.Rand, want bool) string {
	for {
		b := make([]byte, 11)
		b[0] = byte('1' + rng.Intn(9))
		for i := 1; i < 11; i++ {
			b[i] = byte('0' + rng.Intn(10))
		}
		if ValidABN(string(b)) == want {
			return string(b)
		}
	}
}

func formatABN(s string, style int) string {
	switch style {
	case 0:
		return s
	default:
		return s[0:2] + " " + s[2:5] + " " + s[5:8] + " " + s[8:11]
	}
}

func randomACN(rng *rand.Rand, want bool) string {
	for {
		b := make([]byte, 9)
		for i := range b {
			b[i] = byte('0' + rng.Intn(10))
		}
		if ValidACN(string(b)) == want {
			return string(b[0:3]) + " " + string(b[3:6]) + " " + string(b[6:9])
		}
	}
}

func TestABNChecksumKnownValues(t *testing.T) {
	// The worked example published by the Australian Business Register.
	if !ValidABN("51 824 753 556") {
		t.Fatal("ABR worked example should be valid")
	}
	if ValidABN("51 000 000 000") {
		t.Fatal("fixture ABN must fail the checksum")
	}
	if ValidABN("1234") {
		t.Fatal("short value is not an ABN")
	}
}

func TestACNChecksumKnownValues(t *testing.T) {
	// Worked examples published by ASIC.
	for _, v := range []string{"000 000 019", "000 250 000", "010 499 966"} {
		if !ValidACN(v) {
			t.Fatalf("%s should be a valid ACN", v)
		}
	}
	if ValidACN("000 000 018") {
		t.Fatal("000 000 018 should be invalid")
	}
}

// Property: real-looking (checksum-valid) ABNs are always rejected, in any common format
// and any surrounding context.
func TestPropertyValidABNAlwaysRejected(t *testing.T) {
	s := testScanner()
	rng := rand.New(rand.NewSource(1))
	contexts := []string{"%s", "ABN: %s", "business_abn,%s\n", `{"abn": "%s"}`, "Grower %s Pty", "id %s end"}
	for i := 0; i < 2000; i++ {
		v := formatABN(randomABN(rng, true), i%2)
		text := fmt.Sprintf(contexts[i%len(contexts)], v)
		if kinds(s.ScanText("x.csv", text, true))["abn"] == 0 {
			t.Fatalf("valid ABN not rejected: %q", text)
		}
	}
}

// Property: synthetic (checksum-failing) ABNs are always accepted, and never mistaken for
// an ACN, phone number or anything else.
func TestPropertySyntheticABNAlwaysAccepted(t *testing.T) {
	s := testScanner()
	rng := rand.New(rand.NewSource(2))
	for i := 0; i < 2000; i++ {
		v := formatABN(randomABN(rng, false), i%2)
		text := fmt.Sprintf("business_abn,%s\n", v)
		if f := s.ScanText("x.csv", text, true); len(f) != 0 {
			t.Fatalf("synthetic ABN %q rejected: %v", v, f)
		}
	}
}

func TestPropertyACN(t *testing.T) {
	s := testScanner()
	rng := rand.New(rand.NewSource(3))
	for i := 0; i < 1000; i++ {
		v := randomACN(rng, true)
		if kinds(s.ScanText("x.csv", "acn,"+v+"\n", true))["acn"] == 0 {
			t.Fatalf("valid ACN %q not rejected", v)
		}
		w := randomACN(rng, false)
		if f := s.ScanText("x.csv", "acn,"+w+"\n", true); len(f) != 0 {
			t.Fatalf("synthetic ACN %q rejected: %v", w, f)
		}
	}
}

func TestPropertyCoordinates(t *testing.T) {
	s := testScanner()
	rng := rand.New(rand.NewSource(4))
	for i := 0; i < 1000; i++ {
		// Inside the Gulf St Vincent box: accepted.
		lon := 137.96 + rng.Float64()*0.28
		lat := -35.19 + rng.Float64()*0.33
		doc := fmt.Sprintf(`{"type":"Feature","geometry":{"type":"Point","coordinates":[%.6f,%.6f]}}`, lon, lat)
		v, _ := repo.DecodeJSON([]byte(doc))
		if f := s.ScanValue("a.geojson", v); len(f) != 0 {
			t.Fatalf("synthetic coordinate rejected: %s", doc)
		}
		// Real-looking farmland (inland South Australia or New South Wales): rejected.
		lon = 139 + rng.Float64()*10
		lat = -36 + rng.Float64()*6
		doc = fmt.Sprintf(`{"geometry":{"coordinates":[[[%.6f,%.6f],[%.6f,%.6f]]]}}`, lon, lat, lon+0.01, lat)
		v, _ = repo.DecodeJSON([]byte(doc))
		if len(s.ScanValue("a.geojson", v)) == 0 {
			t.Fatalf("real-looking coordinate accepted: %s", doc)
		}
		text := fmt.Sprintf("site at %.5f, %.5f today", lat, lon)
		if kinds(s.ScanText("notes.md", text, false))["coordinates"] == 0 {
			t.Fatalf("real-looking coordinate pair in text accepted: %s", text)
		}
	}
}

func TestLatLonKeysAndCSV(t *testing.T) {
	s := testScanner()
	v, _ := repo.DecodeYAML([]byte("location:\n  latitude: -34.1\n  longitude: 142.2\n"))
	if len(s.ScanValue("x.yaml", v)) == 0 {
		t.Fatal("lat/lon keys outside the test area should be rejected")
	}
	v, _ = repo.DecodeYAML([]byte("location:\n  latitude: -35.01\n  longitude: 138.1\n"))
	if len(s.ScanValue("x.yaml", v)) != 0 {
		t.Fatal("lat/lon keys inside the test area should be accepted")
	}
	if len(s.ScanCSV("x.csv", []byte("id,lat,lon\n1,-31.95,115.86\n"))) == 0 {
		t.Fatal("CSV lat/lon outside the test area should be rejected")
	}
	if len(s.ScanCSV("x.csv", []byte("id,lat,lon\n1,-28.8,113.9\n"))) != 0 {
		t.Fatal("CSV lat/lon inside the test area should be accepted")
	}
}

func TestOtherDetectors(t *testing.T) {
	s := testScanner()
	cases := []struct {
		text, kind string
		data       bool
	}{
		{"contact grower@realfarm.com.au", "email", false},
		{"mobile 0412 345 678", "phone", false},
		{"ph (08) 9123 4567", "phone", false},
		{"+61 412 345 678", "phone", false},
		{"1800 123 456", "phone", false},
		{"Office at 12 Main Street, Clare", "address", false},
		{"post to PO Box 123", "address", false},
		{"PIC: 3ABCD123", "pic", false},
		{"pic,QABC1234\n", "pic", true},
		{"tag,982 123456789012\n", "nlis", true},
		{"owner,Example Farms Pty Ltd\n", "denylist", true},
	}
	for _, c := range cases {
		if kinds(s.ScanText("x", c.text, c.data))[c.kind] == 0 {
			t.Errorf("%q: expected a %s finding", c.text, c.kind)
		}
	}
}

func TestNoFalsePositivesOnTypicalFixtureContent(t *testing.T) {
	s := testScanner()
	ok := []string{
		"Record,Date,Start,Finish,Block,Variety,Area (ha),Product,APVMA No.,Rate /100L\nAPP-2026-001,16/09/2025,06:30,09:00,Top Shiraz,Shiraz,2.4,Synth Sulfur 800 WG,SYN-90001,400\n",
		`{"generated": "2025-12-01T09:00:00+08:00", "created": 1677735113000, "id": "37820265fa6f31d807c7e13056bd4214"}`,
		"contact someone@example.com or noreply@anthropic.com",
		"coordinates [138.20299999999997, -35.002]",
		"Synth Glypho 540,SYN-91001,1.5,L/ha,80,ground boom,,4,12,E,18,55,Synthetic Aerial & Ground",
		"vintage 2026, total 47.16 kg, 8.064, 0.936",
	}
	for _, text := range ok {
		if f := s.ScanText("fixture.csv", text, true); len(f) != 0 {
			t.Errorf("false positive on %q: %v", text, f)
		}
	}
}

func TestMaskNeverRevealsValue(t *testing.T) {
	v := "51 824 753 556"
	m := mask(v)
	if strings.Contains(m, "824") || !strings.HasSuffix(m, "56") {
		t.Fatalf("mask leaked or lost the suffix: %q", m)
	}
}

func TestXLSXIsScanned(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create("xl/sharedStrings.xml")
	w.Write([]byte(`<sst><si><t>Operator</t></si><si><t>grower@realfarm.com.au</t></si></sst>`))
	zw.Close()
	f := testScanner().ScanFile("records.xlsx", buf.Bytes())
	if kinds(f)["email"] == 0 {
		t.Fatalf("email inside xlsx not found: %v", f)
	}
}

func TestExemptPaths(t *testing.T) {
	s := testScanner()
	s.Exempt = []string{"ci/internal/scrub/*_test.go"}
	if f := s.ScanFile("ci/internal/scrub/scrub_test.go", []byte("grower@realfarm.com.au")); len(f) != 0 {
		t.Fatal("exempt path should be skipped")
	}
	if f := s.ScanFile("fixtures/x.csv", []byte("grower@realfarm.com.au")); len(f) == 0 {
		t.Fatal("non-exempt path should be scanned")
	}
}
