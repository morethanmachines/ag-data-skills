// Package scrub implements the privacy check: it rejects contributions that look as if they
// contain real farm data. It is a backstop; the skill is responsible for never contributing
// such data in the first place (implementation plan, section 6).
package scrub

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

// Finding is one suspected leak. Detail never contains the full suspected value.
type Finding struct {
	Path   string `json:"path"`
	Line   int    `json:"line"`
	Kind   string `json:"kind"`
	Detail string `json:"detail"`
}

func (f Finding) String() string {
	if f.Line > 0 {
		return fmt.Sprintf("%s:%d: [%s] %s", f.Path, f.Line, f.Kind, f.Detail)
	}
	return fmt.Sprintf("%s: [%s] %s", f.Path, f.Kind, f.Detail)
}

// Scanner holds the rules loaded from governance.
type Scanner struct {
	Areas  []repo.TestArea
	Deny   []string
	Exempt []string
}

// NewScanner loads the scanner rules from the repository.
func NewScanner(r *repo.Repo) (*Scanner, error) {
	syn, err := r.Synthetic()
	if err != nil {
		return nil, err
	}
	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}
	var deny struct {
		Markers []string `yaml:"markers"`
	}
	if r.Exists("governance/scrub-denylist.yaml") {
		if err := r.ReadConfigInto("governance/scrub-denylist.yaml", &deny); err != nil {
			return nil, err
		}
	}
	return &Scanner{Areas: syn.TestAreas, Deny: deny.Markers, Exempt: cfg.Privacy.ScrubExemptPaths}, nil
}

// IsExempt reports whether a path matches a governed exemption pattern.
func (s *Scanner) IsExempt(p string) bool {
	for _, pat := range s.Exempt {
		if ok, _ := path.Match(pat, p); ok {
			return true
		}
	}
	return false
}

// ---- identifier checksums ----

func digitsOf(s string) []int {
	var d []int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			d = append(d, int(c-'0'))
		}
	}
	return d
}

// ValidABN reports whether s (11 digits, spaces allowed) passes the ABN checksum.
func ValidABN(s string) bool {
	d := digitsOf(s)
	if len(d) != 11 {
		return false
	}
	w := []int{10, 1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	d[0]--
	if d[0] < 0 {
		return false
	}
	sum := 0
	for i := range d {
		sum += d[i] * w[i]
	}
	return sum%89 == 0
}

// ValidACN reports whether s (9 digits, spaces allowed) passes the ACN checksum.
func ValidACN(s string) bool {
	d := digitsOf(s)
	if len(d) != 9 {
		return false
	}
	sum := 0
	for i := 0; i < 8; i++ {
		sum += d[i] * (8 - i)
	}
	check := (10 - sum%10) % 10
	return check == d[8]
}

// mask keeps only the last two characters of a value.
func mask(v string) string {
	r := []rune(strings.TrimSpace(v))
	if len(r) <= 2 {
		return "**"
	}
	var b strings.Builder
	for i, c := range r {
		if i >= len(r)-2 || c == ' ' || c == '-' || c == '.' || c == '@' {
			b.WriteRune(c)
		} else {
			b.WriteRune('*')
		}
	}
	return b.String()
}

// ---- text detectors ----

var (
	reABN       = regexp.MustCompile(`\d{2} ?\d{3} ?\d{3} ?\d{3}`)
	reACN       = regexp.MustCompile(`\d{3} \d{3} \d{3}`)
	reACNLabel  = regexp.MustCompile(`(?i)\bACN\b[^0-9\n]{0,6}(\d{9}|\d{3} \d{3} \d{3})`)
	reEmail     = regexp.MustCompile(`[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(?:\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}`)
	rePhone     = regexp.MustCompile(`(?:\+61[ -]?\(?0?|\(?0)[2-478]\)?(?:[ -]?\d){8}`)
	rePhone13   = regexp.MustCompile(`1[38]00[ -]?\d{3}[ -]?\d{3}`)
	reStreet    = regexp.MustCompile(`\b\d{1,5}[A-Za-z]?(?:-\d{1,5})?\s+(?:[A-Z][A-Za-z']+\s+){1,3}(?:Street|St|Road|Rd|Avenue|Ave|Lane|Ln|Drive|Dr|Court|Ct|Place|Pl|Highway|Hwy|Crescent|Cres|Parade|Pde|Terrace|Tce|Boulevard|Blvd|Close|Way|Track|Trk)\b`)
	rePOBox     = regexp.MustCompile(`(?i)\b(?:P\.?\s?O\.?\s+Box|RMB|Locked\s+Bag)\s+\d+`)
	rePICLabel  = regexp.MustCompile(`(?i)\b(?:PIC|property identification code)\b[^A-Za-z0-9\n]{0,6}([A-Z0-9]{8})\b`)
	rePIC       = regexp.MustCompile(`\b(?:3[A-Z]{4}\d{3}|Q[A-Z]{3}\d{4}|[NSTWM][A-Z]{1,3}\d{4,6})\b`)
	reNLISRFID  = regexp.MustCompile(`\b(?:036|9\d{2}) ?\d{12}\b`)
	reNLISID    = regexp.MustCompile(`\b[2-7NQSTWM][A-Z0-9]{7}[A-Z][A-Z0-9]{2}\d{5}\b`)
	reCoordPair = regexp.MustCompile(`(-?\d{1,3}\.\d{2,})\s*,\s*(-?\d{1,3}\.\d{2,})`)
)

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// isolated reports whether the match at [s,e) is not part of a longer number, a decimal
// or a longer space-separated digit group.
func isolated(text string, s, e int) bool {
	if s > 0 && (isDigit(text[s-1]) || text[s-1] == '.') {
		return false
	}
	if s > 1 && text[s-1] == ' ' && isDigit(text[s-2]) {
		return false
	}
	if e < len(text) && isDigit(text[e]) {
		return false
	}
	if e+1 < len(text) && (text[e] == '.' || text[e] == ' ') && isDigit(text[e+1]) {
		return false
	}
	return true
}

var allowedEmailDomains = []string{"example.com", "example.org", "example.net", "users.noreply.github.com", "anthropic.com"}
var allowedEmailTLDs = []string{".example", ".test", ".invalid", ".localhost"}

func emailAllowed(e string) bool {
	at := strings.LastIndexByte(e, '@')
	dom := strings.ToLower(e[at+1:])
	for _, d := range allowedEmailDomains {
		if dom == d || strings.HasSuffix(dom, "."+d) {
			return true
		}
	}
	for _, t := range allowedEmailTLDs {
		if strings.HasSuffix(dom, t) {
			return true
		}
	}
	return false
}

func lineOf(text string, off int) int { return strings.Count(text[:off], "\n") + 1 }

// IsDataFile reports whether a path holds data (as opposed to prose or code). Some
// detectors (unlabelled PIC patterns, deny-list markers) only apply to data files.
func IsDataFile(p string) bool {
	switch strings.ToLower(path.Ext(p)) {
	case ".csv", ".tsv", ".json", ".geojson", ".yaml", ".yml", ".xlsx", ".txt", ".xml":
		return !strings.HasPrefix(p, "governance/") && !strings.HasPrefix(p, ".github/")
	}
	return false
}

// ScanText runs the text detectors over content.
func (s *Scanner) ScanText(p, text string, data bool) []Finding {
	var out []Finding
	add := func(off int, kind, detail string) {
		out = append(out, Finding{Path: p, Line: lineOf(text, off), Kind: kind, Detail: detail})
	}
	for _, m := range reABN.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if isolated(text, m[0], m[1]) && ValidABN(v) {
			add(m[0], "abn", "value passes the ABN checksum: "+mask(v)+" (synthetic ABNs must fail it)")
		}
	}
	for _, m := range reACN.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if isolated(text, m[0], m[1]) && ValidACN(v) {
			add(m[0], "acn", "value passes the ACN checksum: "+mask(v))
		}
	}
	for _, m := range reACNLabel.FindAllStringSubmatchIndex(text, -1) {
		v := text[m[2]:m[3]]
		if ValidACN(v) && !strings.Contains(v, " ") {
			add(m[2], "acn", "labelled ACN passes the checksum: "+mask(v))
		}
	}
	for _, m := range reEmail.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if !emailAllowed(v) {
			add(m[0], "email", "email address: "+mask(v))
		}
	}
	for _, re := range []*regexp.Regexp{rePhone, rePhone13} {
		for _, m := range re.FindAllStringIndex(text, -1) {
			if isolated(text, m[0], m[1]) {
				add(m[0], "phone", "Australian phone number: "+mask(text[m[0]:m[1]]))
			}
		}
	}
	for _, m := range reStreet.FindAllStringIndex(text, -1) {
		add(m[0], "address", "street address: "+mask(text[m[0]:m[1]]))
	}
	for _, m := range rePOBox.FindAllStringIndex(text, -1) {
		add(m[0], "address", "postal address: "+mask(text[m[0]:m[1]]))
	}
	for _, m := range rePICLabel.FindAllStringSubmatchIndex(text, -1) {
		v := text[m[2]:m[3]]
		if hasDigitsAndLetters(v) {
			add(m[2], "pic", "property identification code: "+mask(v))
		}
	}
	if data {
		for _, m := range rePIC.FindAllStringIndex(text, -1) {
			if m[1]-m[0] == 8 {
				add(m[0], "pic", "looks like a property identification code: "+mask(text[m[0]:m[1]]))
			}
		}
		for _, m := range reNLISRFID.FindAllStringIndex(text, -1) {
			add(m[0], "nlis", "looks like an NLIS RFID device number: "+mask(text[m[0]:m[1]]))
		}
		for _, m := range reNLISID.FindAllStringIndex(text, -1) {
			add(m[0], "nlis", "looks like an NLIS device id: "+mask(text[m[0]:m[1]]))
		}
		low := strings.ToLower(text)
		for _, d := range s.Deny {
			if d == "" {
				continue
			}
			if i := strings.Index(low, strings.ToLower(d)); i >= 0 {
				add(i, "denylist", fmt.Sprintf("contains real-data marker %q", d))
			}
		}
	}
	for _, m := range reCoordPair.FindAllStringSubmatchIndex(text, -1) {
		a, _ := strconv.ParseFloat(text[m[2]:m[3]], 64)
		b, _ := strconv.ParseFloat(text[m[4]:m[5]], 64)
		for _, ll := range [][2]float64{{a, b}, {b, a}} { // (lat, lon) then (lon, lat)
			lat, lon := ll[0], ll[1]
			if plausibleAU(lat, lon) && !s.inArea(lon, lat) {
				add(m[0], "coordinates", fmt.Sprintf("coordinate pair in Australia outside the synthetic test areas (near %.1f, %.1f)", lat, lon))
				break
			}
		}
	}
	return out
}

func hasDigitsAndLetters(v string) bool {
	var d, l bool
	for _, c := range v {
		if c >= '0' && c <= '9' {
			d = true
		} else if c >= 'A' && c <= 'Z' {
			l = true
		}
	}
	return d && l
}

func plausibleAU(lat, lon float64) bool {
	return lat >= -45 && lat <= -9 && lon >= 110 && lon <= 155
}

func (s *Scanner) inArea(lon, lat float64) bool {
	for _, a := range s.Areas {
		if lon >= a.MinLon && lon <= a.MaxLon && lat >= a.MinLat && lat <= a.MaxLat {
			return true
		}
	}
	return false
}

// ---- structured coordinate checks ----

func num(v any) (float64, bool) {
	switch t := v.(type) {
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case float64:
		return t, true
	case int:
		return float64(t), true
	}
	return 0, false
}

// ScanValue checks coordinates in a decoded JSON or YAML document: GeoJSON coordinates
// arrays and objects with latitude and longitude keys. Every coordinate must lie in a
// designated test area.
func (s *Scanner) ScanValue(p string, v any) []Finding {
	var out []Finding
	bad := 0
	var walkCoords func(c any)
	walkCoords = func(c any) {
		arr, ok := c.([]any)
		if !ok {
			return
		}
		if len(arr) >= 2 {
			lon, ok1 := num(arr[0])
			lat, ok2 := num(arr[1])
			if ok1 && ok2 {
				if !s.inArea(lon, lat) {
					bad++
				}
				return
			}
		}
		for _, x := range arr {
			walkCoords(x)
		}
	}
	var walk func(x any)
	walk = func(x any) {
		switch t := x.(type) {
		case map[string]any:
			var lat, lon *float64
			for k, val := range t {
				lk := strings.ToLower(k)
				switch lk {
				case "coordinates":
					walkCoords(val)
				case "lat", "latitude", "lattitude":
					if f, ok := num(val); ok {
						lat = &f
					}
				case "lon", "lng", "long", "longitude":
					if f, ok := num(val); ok {
						lon = &f
					}
				}
			}
			if lat != nil && lon != nil && !s.inArea(*lon, *lat) {
				bad++
			}
			for k, val := range t {
				if strings.ToLower(k) != "coordinates" {
					walk(val)
				}
			}
		case []any:
			for _, e := range t {
				walk(e)
			}
		}
	}
	walk(v)
	if bad > 0 {
		out = append(out, Finding{Path: p, Kind: "coordinates", Detail: fmt.Sprintf("%d coordinate(s) outside the designated synthetic test areas (governance/synthetic.yaml)", bad)})
	}
	return out
}

// ScanCSV checks latitude and longitude columns of a CSV file.
func (s *Scanner) ScanCSV(p string, content []byte) []Finding {
	rd := csv.NewReader(bytes.NewReader(content))
	rd.FieldsPerRecord = -1
	rows, err := rd.ReadAll()
	if err != nil || len(rows) < 2 {
		return nil
	}
	latI, lonI := -1, -1
	for i, h := range rows[0] {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "lat", "latitude":
			latI = i
		case "lon", "lng", "long", "longitude":
			lonI = i
		}
	}
	if latI < 0 || lonI < 0 {
		return nil
	}
	bad := 0
	for _, r := range rows[1:] {
		if latI >= len(r) || lonI >= len(r) {
			continue
		}
		lat, e1 := strconv.ParseFloat(strings.TrimSpace(r[latI]), 64)
		lon, e2 := strconv.ParseFloat(strings.TrimSpace(r[lonI]), 64)
		if e1 == nil && e2 == nil && !s.inArea(lon, lat) {
			bad++
		}
	}
	if bad > 0 {
		return []Finding{{Path: p, Kind: "coordinates", Detail: fmt.Sprintf("%d row(s) with coordinates outside the synthetic test areas", bad)}}
	}
	return nil
}

// XLSXText extracts the text of every shared string and cell value in a workbook.
func XLSXText(content []byte) (string, error) {
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", err
	}
	var names []string
	files := map[string]*zip.File{}
	for _, f := range zr.File {
		if f.Name == "xl/sharedStrings.xml" || strings.HasPrefix(f.Name, "xl/worksheets/") || f.Name == "docProps/core.xml" || f.Name == "docProps/app.xml" {
			names = append(names, f.Name)
			files[f.Name] = f
		}
	}
	sort.Strings(names)
	var b strings.Builder
	tag := regexp.MustCompile(`<[^>]+>`)
	for _, n := range names {
		rc, err := files[n].Open()
		if err != nil {
			return "", err
		}
		x, _ := io.ReadAll(rc)
		rc.Close()
		// Put each element's text on its own line so values do not run together.
		t := tag.ReplaceAllString(strings.ReplaceAll(string(x), "<", "\n<"), "")
		for _, l := range strings.Split(t, "\n") {
			if l = strings.TrimSpace(l); l != "" {
				b.WriteString(l)
				b.WriteByte('\n')
			}
		}
	}
	return b.String(), nil
}

// ScanFile scans one file's content.
func (s *Scanner) ScanFile(p string, content []byte) []Finding {
	if s.IsExempt(p) {
		return nil
	}
	data := IsDataFile(p)
	out := s.ScanText(p, p, false) // the file name itself
	for i := range out {
		out[i].Detail = "in file name: " + out[i].Detail
		out[i].Line = 0
	}
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".xlsx", ".xlsm":
		t, err := XLSXText(content)
		if err != nil {
			return append(out, Finding{Path: p, Kind: "unreadable", Detail: "could not read workbook: " + err.Error()})
		}
		f := s.ScanText(p, t, true)
		for i := range f {
			f[i].Line = 0
		}
		return append(out, f...)
	case ".png", ".jpg", ".jpeg", ".gif", ".pdf", ".zip", ".gz", ".tgz", ".xls", ".docx", ".sqlite", ".db", ".shp", ".kmz":
		return append(out, Finding{Path: p, Kind: "binary", Detail: "binary file type the privacy check cannot read; contribute text formats (CSV, JSON, YAML) or xlsx"})
	}
	if bytes.IndexByte(content, 0) >= 0 {
		return append(out, Finding{Path: p, Kind: "binary", Detail: "binary content the privacy check cannot read"})
	}
	text := string(content)
	out = append(out, s.ScanText(p, text, data)...)
	switch ext {
	case ".json", ".geojson":
		if v, err := repo.DecodeJSON(content); err == nil {
			out = append(out, s.ScanValue(p, v)...)
		}
	case ".yaml", ".yml":
		if v, err := repo.DecodeYAML(content); err == nil {
			out = append(out, s.ScanValue(p, v)...)
		}
	case ".csv":
		out = append(out, s.ScanCSV(p, content)...)
	}
	return out
}
