// Package checks implements the deterministic repository checks run by the checks
// workflow. Every check returns findings; an empty result means
// the repository passes.
package checks

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/fingerprint"
	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
	"github.com/morethanmachines/ag-data-skills/ci/internal/schema"
)

// Finding is one failed check.
type Finding struct {
	Path  string `json:"path"`
	Check string `json:"check"`
	Msg   string `json:"message"`
}

func (f Finding) String() string { return fmt.Sprintf("%s: [%s] %s", f.Path, f.Check, f.Msg) }

// Checker runs checks over one repository.
type Checker struct {
	R          *repo.Repo
	C          *schema.Compiler
	Cfg        *repo.Config
	industries map[string]bool
	registry   map[string]bool
	decisions  map[string]bool
	specs      map[string]string // spec id -> schema path
	profiles   map[string]string // profile id -> profile path
	findings   []Finding
}

// New prepares a checker.
func New(r *repo.Repo) (*Checker, error) {
	cfg, err := r.Config()
	if err != nil {
		return nil, fmt.Errorf("governance/config.yaml: %w", err)
	}
	inds, err := r.IndustryIDs()
	if err != nil {
		return nil, fmt.Errorf("governance/industries.yaml: %w", err)
	}
	return &Checker{R: r, C: schema.New(r), Cfg: cfg, industries: inds}, nil
}

func (c *Checker) fail(p, check, format string, a ...any) {
	c.findings = append(c.findings, Finding{Path: p, Check: check, Msg: fmt.Sprintf(format, a...)})
}

// Findings returns the findings so far, sorted.
func (c *Checker) Findings() []Finding {
	sort.SliceStable(c.findings, func(i, j int) bool {
		if c.findings[i].Path != c.findings[j].Path {
			return c.findings[i].Path < c.findings[j].Path
		}
		return c.findings[i].Check < c.findings[j].Check
	})
	return c.findings
}

// validateAgainst validates the file at p (JSON or YAML) against the schema at s.
func (c *Checker) validateAgainst(p, s, check string) any {
	v, err := c.R.ReadAny(p)
	if err != nil {
		c.fail(p, check, "cannot parse: %v", err)
		return nil
	}
	errs, err := c.C.Validate(s, v)
	if err != nil {
		c.fail(s, check, "schema does not compile: %v", err)
		return v
	}
	for _, e := range errs {
		c.fail(p, check, "does not validate against %s %s", s, e)
	}
	return v
}

var semver = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

var allowedSemantics = map[string]bool{
	"unit": true, "rate_basis": true, "moisture_basis": true, "method": true, "crs": true,
	"timezone": true, "measurement_status": true, "identifier_scheme": true,
}

var tierRank = map[string]int{"E0": 0, "E1": 1, "E2": 2, "E3": 3, "E4": 4, "E5": 5}

// RunAll runs every repository-wide check.
func (c *Checker) RunAll() []Finding {
	c.checkGovernance()
	c.checkRegistry()
	c.checkDecisions()
	c.checkForbidden()
	c.checkIndustryFolders()
	c.checkCodelists()
	c.checkSpecs()
	c.checkProfiles()
	c.checkExamples()
	c.checkCrosswalks()
	c.checkSightingsAndUsage()
	c.checkTasks()
	c.checkFixtures()
	c.checkNumericConfidence()
	return c.Findings()
}

func (c *Checker) checkGovernance() {
	var tiers struct {
		Tiers []struct {
			ID     string  `yaml:"id"`
			Weight float64 `yaml:"weight"`
		} `yaml:"tiers"`
	}
	if err := c.R.ReadConfigInto("governance/evidence-tiers.yaml", &tiers); err != nil {
		c.fail("governance/evidence-tiers.yaml", "governance", "cannot parse: %v", err)
	} else {
		want := []string{"E5", "E4", "E3", "E2", "E1", "E0"}
		if len(tiers.Tiers) != len(want) {
			c.fail("governance/evidence-tiers.yaml", "governance", "must define exactly the tiers E5 to E0")
		} else {
			for i, t := range tiers.Tiers {
				if t.ID != want[i] {
					c.fail("governance/evidence-tiers.yaml", "governance", "tier %d is %s, want %s", i, t.ID, want[i])
				}
				if t.Weight < 0 || t.Weight > 1 {
					c.fail("governance/evidence-tiers.yaml", "governance", "%s weight %v outside 0 to 1", t.ID, t.Weight)
				}
				if i > 0 && t.Weight >= tiers.Tiers[i-1].Weight {
					c.fail("governance/evidence-tiers.yaml", "governance", "weights must strictly decrease from E5 to E0 (%s)", t.ID)
				}
			}
		}
	}
	inds, _ := c.R.Industries()
	seen := map[string]bool{}
	idRe := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	for _, i := range inds {
		if !idRe.MatchString(i.ID) {
			c.fail("governance/industries.yaml", "governance", "industry id %q is not lower-case kebab", i.ID)
		}
		if seen[i.ID] {
			c.fail("governance/industries.yaml", "governance", "duplicate industry id %q", i.ID)
		}
		seen[i.ID] = true
	}
	if c.Cfg.Drafts.OverlapThreshold <= 0 || c.Cfg.Drafts.OverlapThreshold > 1 {
		c.fail("governance/config.yaml", "governance", "drafts.overlap_threshold must be in (0, 1]")
	}
	if _, ok := tierRank[c.Cfg.Evidence.MinTierForCrosswalkRequiredFields]; !ok {
		c.fail("governance/config.yaml", "governance", "evidence.min_tier_for_crosswalk_required_fields is not a tier")
	}
	if _, ok := tierRank[c.Cfg.Evidence.MinTierForSpecChange]; !ok {
		c.fail("governance/config.yaml", "governance", "evidence.min_tier_for_spec_change is not a tier")
	}
	for _, s := range c.R.Glob("governance/schemas/*.json") {
		if _, err := c.C.Compile(s); err != nil {
			c.fail(s, "governance", "schema does not compile: %v", err)
		}
	}
}

func (c *Checker) checkRegistry() {
	c.registry = map[string]bool{}
	p := "external/registry.yaml"
	v := c.validateAgainst(p, "governance/schemas/external-registry.schema.json", "registry")
	m, _ := v.(map[string]any)
	entries, _ := m["entries"].([]any)
	for _, e := range entries {
		em, _ := e.(map[string]any)
		id, _ := em["id"].(string)
		if c.registry[id] {
			c.fail(p, "registry", "duplicate entry id %q", id)
		}
		c.registry[id] = true
	}
}

func (c *Checker) checkDecisions() {
	c.decisions = map[string]bool{}
	for _, p := range c.R.Glob("records/decisions/*.yaml") {
		id := strings.TrimSuffix(path.Base(p), ".yaml")
		c.decisions[id] = true
		v := c.validateAgainst(p, "governance/schemas/decision.schema.json", "decision")
		if m, ok := v.(map[string]any); ok && m["id"] != id {
			c.fail(p, "decision", "id %v does not match file name", m["id"])
		}
	}
}

var forbidden = regexp.MustCompile(`(^|/)(ground_truth|private)/|expected_answers|applications_truth|(^|/)generate\.py$|(^|/)check_exports\.py$`)

func (c *Checker) checkForbidden() {
	files, _ := c.R.Files()
	for _, f := range files {
		if forbidden.MatchString(f) {
			c.fail(f, "hidden-fixtures", "ground truth, generators and expected answers belong in the private hidden-fixtures store, never in this repository")
		}
	}
}

func (c *Checker) checkIndustryFolders() {
	entries, _ := os.ReadDir(c.R.Abs("industries"))
	for _, e := range entries {
		if e.IsDir() && !c.industries[e.Name()] {
			c.fail("industries/"+e.Name(), "industries", "industry %q is not in governance/industries.yaml", e.Name())
		}
	}
}

func (c *Checker) checkIndustryTags(p string, tags []any, allowShared bool) {
	for _, t := range tags {
		s, _ := t.(string)
		if s == "shared" && allowShared {
			continue
		}
		if !c.industries[s] {
			c.fail(p, "industries", "industry tag %q is not in governance/industries.yaml", s)
		}
	}
}

func (c *Checker) checkCodelists() {
	for _, p := range c.R.Glob("codelists/*/codelist.json") {
		v := c.validateAgainst(p, "governance/schemas/codelist.schema.json", "codelist")
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		dir := path.Base(path.Dir(p))
		if m["id"] != dir {
			c.fail(p, "codelist", "id %v does not match folder %q", m["id"], dir)
		}
		tags, _ := m["industries"].([]any)
		c.checkIndustryTags(p, tags, true)
		seen := map[string]bool{}
		vals, _ := m["values"].([]any)
		for _, x := range vals {
			xm, _ := x.(map[string]any)
			code, _ := xm["code"].(string)
			if seen[code] {
				c.fail(p, "codelist", "duplicate code %q", code)
			}
			seen[code] = true
		}
		c.checkAnnotations(p, m)
	}
}

// specID derives the expected x-cp-id from a schema path.
func specID(p string) string {
	parts := strings.Split(p, "/")
	if parts[0] == "shared" && len(parts) == 4 {
		return "shared/" + parts[2]
	}
	if parts[0] == "industries" && len(parts) == 5 {
		return parts[1] + "/" + parts[3]
	}
	return ""
}

func (c *Checker) specPaths() []string {
	return append(c.R.Glob("shared/specs/*/schema.json"), c.R.Glob("industries/*/specs/*/schema.json")...)
}

func (c *Checker) checkSpecs() {
	c.specs = map[string]string{}
	for _, p := range c.specPaths() {
		doc, err := c.R.ReadJSON(p)
		if err != nil {
			c.fail(p, "spec", "cannot parse: %v", err)
			continue
		}
		m, _ := doc.(map[string]any)
		if s, _ := m["$schema"].(string); s != "https://json-schema.org/draft/2020-12/schema" {
			c.fail(p, "spec", "$schema must be JSON Schema 2020-12")
		}
		if _, err := c.C.Compile(p); err != nil {
			c.fail(p, "spec", "does not compile: %v", err)
		}
		id, _ := m["x-cp-id"].(string)
		if want := specID(p); id != want {
			c.fail(p, "spec", "x-cp-id %q does not match its path (want %q)", id, want)
		}
		c.specs[id] = p
		ver, _ := m["x-cp-version"].(string)
		mm := semver.FindStringSubmatch(ver)
		status, _ := m["x-cp-status"].(string)
		switch {
		case mm == nil:
			c.fail(p, "spec", "x-cp-version %q is not semver", ver)
		case mm[1] == "0" && status != "draft":
			c.fail(p, "spec", "0.x versions must have x-cp-status draft")
		case mm[1] != "0" && status != "active" && status != "deprecated":
			c.fail(p, "spec", "versions 1.0.0 and above must be active or deprecated")
		}
		if _, ok := m["x-cp-external-basis"]; !ok {
			c.fail(p, "spec", "missing x-cp-external-basis (use [] if none)")
		}
		if prov, ok := m["x-cp-provenance"].([]any); !ok || len(prov) == 0 {
			c.fail(p, "spec", "missing x-cp-provenance decision record ids")
		}
		if _, ok := m["x-cp-owner"]; ok {
			c.fail(p, "spec", "specs have no owner field")
		}
		if !c.R.Exists(path.Join(path.Dir(p), "KNOWN_ISSUES.md")) {
			c.fail(path.Dir(p), "spec", "missing KNOWN_ISSUES.md")
		}
		c.checkAnnotations(p, m)
	}
}

// checkAnnotations checks x-cp- references anywhere in a document.
func (c *Checker) checkAnnotations(p string, doc any) {
	var walk func(n any, at string)
	walk = func(n any, at string) {
		switch t := n.(type) {
		case map[string]any:
			for k, v := range t {
				switch k {
				case "x-cp-external-basis", "external_basis":
					arr, _ := v.([]any)
					for _, x := range arr {
						if s, _ := x.(string); !c.registry[s] {
							c.fail(p, "external-basis", "%s: %q is not in external/registry.yaml", at, s)
						}
					}
				case "x-cp-provenance":
					arr, _ := v.([]any)
					for _, x := range arr {
						if s, _ := x.(string); !c.decisions[s] {
							c.fail(p, "provenance", "%s: decision record %q does not exist in records/decisions/", at, s)
						}
					}
				case "provenance":
					arr, _ := v.([]any)
					for _, x := range arr {
						if s, _ := x.(string); strings.HasPrefix(s, "bootstrap-") || strings.HasPrefix(s, "pr-") {
							if !c.decisions[s] {
								c.fail(p, "provenance", "%s: decision record %q does not exist", at, s)
							}
						}
					}
				case "x-cp-codelist":
					s, _ := v.(string)
					switch {
					case strings.HasPrefix(s, "codelists/"):
						if !c.R.Exists(s + "/codelist.json") {
							c.fail(p, "codelist-ref", "%s: code list %q does not exist", at, s)
						}
					case strings.HasPrefix(s, "external:"):
						if !c.registry[strings.TrimPrefix(s, "external:")] {
							c.fail(p, "codelist-ref", "%s: %q is not in external/registry.yaml", at, s)
						}
					default:
						c.fail(p, "codelist-ref", "%s: x-cp-codelist must be codelists/<id> or external:<id>", at)
					}
				case "x-cp-required-semantics":
					arr, _ := v.([]any)
					for _, x := range arr {
						if s, _ := x.(string); !allowedSemantics[s] {
							c.fail(p, "semantics", "%s: unknown semantic attribute %q", at, s)
						}
					}
				}
				walk(v, at+"/"+k)
			}
		case []any:
			for i, v := range t {
				walk(v, at+"/"+strconv.Itoa(i))
			}
		}
	}
	walk(doc, "")
}

// ---- profiles ----

var profileTopKeys = map[string]bool{"$schema": true, "$id": true, "title": true, "description": true, "allOf": true, "properties": true, "required": true, "$comment": true}
var profileConstraintKeys = map[string]bool{
	"const": true, "enum": true, "minimum": true, "maximum": true, "exclusiveMinimum": true, "exclusiveMaximum": true,
	"minLength": true, "maxLength": true, "pattern": true, "minItems": true, "maxItems": true, "format": true,
	"description": true, "title": true, "x-cp-codelist": true, "$comment": true, "examples": true,
}

func typeSet(v any) map[string]bool {
	s := map[string]bool{}
	switch t := v.(type) {
	case string:
		s[t] = true
	case []any:
		for _, x := range t {
			if str, ok := x.(string); ok {
				s[str] = true
			}
		}
	}
	return s
}

// resolve follows local $refs ("#/$defs/...") in the shared document.
func resolve(root map[string]any, n any) map[string]any {
	m, _ := n.(map[string]any)
	for i := 0; i < 20 && m != nil; i++ {
		ref, ok := m["$ref"].(string)
		if !ok || !strings.HasPrefix(ref, "#/") {
			return m
		}
		var cur any = root
		for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
			cm, _ := cur.(map[string]any)
			cur = cm[part]
		}
		m, _ = cur.(map[string]any)
	}
	return m
}

// ProfileViolations walks a profile subtree against the shared spec subtree it constrains
// and reports anything that redefines rather than constrains.
func ProfileViolations(sharedRoot map[string]any, prof map[string]any, shared any, at string) []string {
	var out []string
	s := resolve(sharedRoot, shared)
	if s == nil {
		return []string{at + ": the shared spec has nothing here to constrain"}
	}
	for k, v := range prof {
		switch {
		case k == "properties":
			sp, _ := s["properties"].(map[string]any)
			pm, _ := v.(map[string]any)
			for name, sub := range pm {
				ss, ok := sp[name]
				if !ok {
					out = append(out, fmt.Sprintf("%s: introduces property %q that the shared spec does not define (propose it to the shared spec or use x_local_)", at, name))
					continue
				}
				subm, _ := sub.(map[string]any)
				out = append(out, ProfileViolations(sharedRoot, subm, ss, at+"/"+name)...)
			}
		case k == "items":
			si, ok := s["items"]
			if !ok {
				out = append(out, at+": constrains items of something that is not an array in the shared spec")
				continue
			}
			im, _ := v.(map[string]any)
			out = append(out, ProfileViolations(sharedRoot, im, si, at+"/items")...)
		case k == "required":
			sp, _ := s["properties"].(map[string]any)
			arr, _ := v.([]any)
			for _, x := range arr {
				if name, _ := x.(string); sp[name] == nil {
					out = append(out, fmt.Sprintf("%s: requires %q, which the shared spec does not define", at, name))
				}
			}
		case k == "type":
			pt, st := typeSet(v), typeSet(s["type"])
			if len(st) == 0 {
				continue // shared leaves type open; narrowing is fine
			}
			for t := range pt {
				if !st[t] && !(t == "integer" && st["number"]) {
					out = append(out, fmt.Sprintf("%s: type %q is not allowed by the shared spec (profiles may only narrow types)", at, t))
				}
			}
		case profileConstraintKeys[k]:
		default:
			out = append(out, fmt.Sprintf("%s: keyword %q is not allowed in a profile (profiles may only add constraints)", at, k))
		}
	}
	return out
}

func (c *Checker) checkProfiles() {
	c.profiles = map[string]string{}
	for _, p := range c.R.Glob("industries/*/profiles/*.json") {
		doc, err := c.R.ReadJSON(p)
		if err != nil {
			c.fail(p, "profile", "cannot parse: %v", err)
			continue
		}
		m, _ := doc.(map[string]any)
		ind := strings.Split(p, "/")[1]
		name := strings.TrimSuffix(path.Base(p), ".json")
		id, _ := m["x-cp-id"].(string)
		if id != ind+"/"+name {
			c.fail(p, "profile", "x-cp-id %q does not match its path (want %q)", id, ind+"/"+name)
		}
		c.profiles[id] = p
		of, _ := m["x-cp-profile-of"].(string)
		if of != "shared/"+name {
			c.fail(p, "profile", "x-cp-profile-of %q must be the shared spec named by the file (shared/%s)", of, name)
			continue
		}
		sharedPath := "shared/specs/" + name + "/schema.json"
		if !c.R.Exists(sharedPath) {
			c.fail(p, "profile", "shared spec %s does not exist", sharedPath)
			continue
		}
		if _, err := c.C.Compile(p); err != nil {
			c.fail(p, "profile", "does not compile: %v", err)
			continue
		}
		wantRef := "../../../" + sharedPath
		all, _ := m["allOf"].([]any)
		okRef := len(all) == 1
		if okRef {
			r0, _ := all[0].(map[string]any)
			okRef = len(r0) == 1 && r0["$ref"] == wantRef
		}
		if !okRef {
			c.fail(p, "profile", "allOf must be exactly [{\"$ref\": %q}]", wantRef)
		}
		top := map[string]any{}
		for k, v := range m {
			if strings.HasPrefix(k, "x-cp-") {
				continue
			}
			if !profileTopKeys[k] {
				c.fail(p, "profile", "top-level keyword %q is not allowed in a profile", k)
				continue
			}
			if k == "properties" || k == "required" {
				top[k] = v
			}
		}
		shared, _ := c.R.ReadJSON(sharedPath)
		sm, _ := shared.(map[string]any)
		for _, v := range ProfileViolations(sm, top, sm, "") {
			c.fail(p, "profile", "%s", v)
		}
		c.checkAnnotations(p, m)
	}
}

// ---- examples ----

func (c *Checker) checkExamples() {
	for id, sp := range c.specs {
		dir := path.Dir(sp)
		for _, ex := range c.R.Glob(dir + "/examples/*.json") {
			v := c.validateAgainst(ex, sp, "example")
			m, _ := v.(map[string]any)
			if pid, ok := m["profile"].(string); ok {
				pp, found := c.profiles[pid]
				if !found {
					c.fail(ex, "example", "profile %q does not exist", pid)
					continue
				}
				c.validateAgainst(ex, pp, "example")
			}
			_ = id
		}
	}
}

// ---- crosswalks ----

func (c *Checker) crosswalkDirs() []string {
	var out []string
	for _, pat := range []string{"shared/crosswalks/*/crosswalk.yaml", "industries/*/crosswalks/*/crosswalk.yaml"} {
		for _, p := range c.R.Glob(pat) {
			out = append(out, path.Dir(p))
		}
	}
	return out
}

func (c *Checker) checkCrosswalks() {
	minTier := tierRank[c.Cfg.Evidence.MinTierForCrosswalkRequiredFields]
	for _, dir := range c.crosswalkDirs() {
		cwPath, fpPath := dir+"/crosswalk.yaml", dir+"/fingerprint.json"
		cwv := c.validateAgainst(cwPath, "governance/schemas/crosswalk.schema.json", "crosswalk")
		if !c.R.Exists(fpPath) {
			c.fail(dir, "crosswalk", "missing fingerprint.json")
			continue
		}
		c.validateAgainst(fpPath, "governance/schemas/fingerprint.schema.json", "fingerprint")
		b, _ := os.ReadFile(c.R.Abs(fpPath))
		var fp fingerprint.Fingerprint
		if err := json.Unmarshal(b, &fp); err == nil {
			if id := fingerprint.ID(fp); path.Base(dir) != id {
				c.fail(dir, "crosswalk", "folder name must be the fingerprint id %s", id)
			}
		}
		cw, ok := cwv.(map[string]any)
		if !ok {
			continue
		}
		unresolved := map[string]bool{}
		us, _ := cw["unresolved_semantics"].([]any)
		for _, u := range us {
			um, _ := u.(map[string]any)
			t, _ := um["target"].(string)
			unresolved[t] = true
		}
		fields, _ := cw["fields"].([]any)
		for _, f := range fields {
			fm, _ := f.(map[string]any)
			target, _ := fm["target"].(string)
			ev, _ := fm["evidence"].(map[string]any)
			tier, _ := ev["tier"].(string)
			if tierRank[tier] < minTier && !unresolved[target] {
				c.fail(cwPath, "evidence", "field %s has evidence %s; below %s it must be listed in unresolved_semantics", target, tier, c.Cfg.Evidence.MinTierForCrosswalkRequiredFields)
			}
		}
		target, _ := cw["target"].(map[string]any)
		specRef, _ := target["spec"].(string)
		specPath := specRef + "/schema.json"
		if !c.R.Exists(specPath) {
			c.fail(cwPath, "crosswalk", "target spec %s does not exist", specRef)
			continue
		}
		fx, _ := cw["fixtures"].(map[string]any)
		resolveRel := func(s string) string { return path.Clean(path.Join(dir, s)) }
		var inputs []string
		switch in := fx["input"].(type) {
		case string:
			inputs = []string{in}
		case []any:
			for _, x := range in {
				if s, ok := x.(string); ok {
					inputs = append(inputs, s)
				}
			}
		}
		for _, in := range inputs {
			if !c.R.Exists(resolveRel(in)) {
				c.fail(cwPath, "crosswalk", "fixture input %s does not exist", in)
			}
		}
		exp, _ := fx["expected_output"].(string)
		ep := resolveRel(exp)
		if !c.R.Exists(ep) {
			c.fail(cwPath, "crosswalk", "expected output %s does not exist", exp)
			continue
		}
		// Regression: the expected output must validate against the current spec and profile.
		c.validateAgainst(ep, specPath, "regression")
		if prof, _ := target["profile"].(string); prof != "" {
			if !c.R.Exists(prof) {
				c.fail(cwPath, "crosswalk", "target profile %s does not exist", prof)
			} else {
				c.validateAgainst(ep, prof, "regression")
			}
		}
	}
}

func (c *Checker) checkSightingsAndUsage() {
	for _, p := range append(c.R.Glob("industries/*/sightings/*.yaml"), c.R.Glob("shared/sightings/*.yaml")...) {
		v := c.validateAgainst(p, "governance/schemas/sighting.schema.json", "sighting")
		if m, ok := v.(map[string]any); ok {
			tags, _ := m["industries"].([]any)
			c.checkIndustryTags(p, tags, false)
		}
	}
	pats := []string{"shared/specs/*/extensions/usage/*.yaml", "industries/*/specs/*/extensions/usage/*.yaml"}
	for _, pat := range pats {
		for _, p := range c.R.Glob(pat) {
			c.validateAgainst(p, "governance/schemas/extension-usage.schema.json", "extension-usage")
		}
	}
}

func (c *Checker) checkTasks() {
	for _, p := range c.R.Glob("industries/*/tasks/*/task.yaml") {
		v := c.validateAgainst(p, "governance/schemas/task.schema.json", "task")
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		dir := path.Dir(p)
		if m["id"] != path.Base(dir) {
			c.fail(p, "task", "id %v does not match folder", m["id"])
		}
		if m["industry"] != strings.Split(p, "/")[1] {
			c.fail(p, "task", "industry %v does not match folder", m["industry"])
		}
		if fs, _ := m["fixture_set"].(string); !c.R.Exists("fixtures/" + fs) {
			c.fail(p, "task", "fixture set %q does not exist", fs)
		}
		ins, _ := m["inputs"].([]any)
		for _, x := range ins {
			if s, _ := x.(string); !c.R.Exists(s) {
				c.fail(p, "task", "input %s does not exist", s)
			}
		}
		if as, _ := m["answer_schema"].(string); as != "" {
			if _, err := c.C.Compile(dir + "/" + as); err != nil {
				c.fail(p, "task", "answer schema does not compile: %v", err)
			}
		}
		if sp, _ := m["spec"].(string); sp != "" && c.specs[sp] == "" && c.profiles[sp] == "" {
			c.fail(p, "task", "spec %q is neither a spec nor a profile", sp)
		}
		for _, f := range c.R.Glob(dir + "/*") {
			b := path.Base(f)
			if b != "task.yaml" && b != "answer.schema.json" && b != "README.md" {
				c.fail(f, "hidden-fixtures", "task folders hold only task.yaml, answer.schema.json and README.md; expected answers belong in the hidden store")
			}
		}
	}
}

func (c *Checker) checkFixtures() {
	entries, _ := os.ReadDir(c.R.Abs("fixtures"))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if !c.R.Exists("fixtures/" + e.Name() + "/README.md") {
			c.fail("fixtures/"+e.Name(), "fixture", "missing README.md (state that the data is synthetic)")
		}
	}
}

// checkNumericConfidence rejects any key containing "confidence" with a numeric value.
func (c *Checker) checkNumericConfidence() {
	files, _ := c.R.Files()
	pct := regexp.MustCompile(`^\s*[0-9]+(\.[0-9]+)?\s*%?\s*$`)
	for _, f := range files {
		ext := strings.ToLower(path.Ext(f))
		if ext != ".json" && ext != ".yaml" && ext != ".yml" && ext != ".geojson" {
			continue
		}
		if strings.HasPrefix(f, "ci/") || strings.HasPrefix(f, ".github/") {
			continue
		}
		v, err := c.R.ReadAny(f)
		if err != nil {
			if strings.HasPrefix(f, "governance/") || strings.HasPrefix(f, "shared/") || strings.HasPrefix(f, "industries/") || strings.HasPrefix(f, "codelists/") || strings.HasPrefix(f, "records/") || strings.HasPrefix(f, "external/") {
				c.fail(f, "parse", "cannot parse: %v", err)
			}
			continue
		}
		var walk func(n any, at string)
		walk = func(n any, at string) {
			switch t := n.(type) {
			case map[string]any:
				for k, x := range t {
					if strings.Contains(strings.ToLower(k), "confidence") {
						isNum := false
						switch y := x.(type) {
						case json.Number:
							isNum = true
						case string:
							isNum = pct.MatchString(y)
						}
						if isNum {
							c.fail(f, "numeric-confidence", "%s/%s: numeric confidence is not allowed; record an evidence tier (E0 to E5) and a reference", at, k)
						}
					}
					walk(x, at+"/"+k)
				}
			case []any:
				for i, x := range t {
					walk(x, at+"/"+strconv.Itoa(i))
				}
			}
		}
		walk(v, "")
	}
}
