// Package index generates the compact indexes agents read instead of browsing the
// repository, GOVERNANCE.md, spec changelogs, and the bundled
// snapshot (4.5). Output is deterministic: the same repository gives the same bytes.
package index

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/fingerprint"
	"github.com/morethanmachines/ag-data-skills/ci/internal/overlap"
	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
	"gopkg.in/yaml.v3"
)

// Entry is one line of an index.
type Entry struct {
	Kind        string   `json:"kind"` // spec, profile, codelist
	ID          string   `json:"id"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords,omitempty"`
	Status      string   `json:"status"`
	Version     string   `json:"version"`
	Path        string   `json:"path"`
	ProfileOf   string   `json:"profile_of,omitempty"`
}

// FingerprintEntry describes a known source shape.
type FingerprintEntry struct {
	ID         string   `json:"id"`
	Required   []string `json:"required_columns"`
	Optional   []string `json:"optional_columns,omitempty"`
	MediaType  string   `json:"media_type"`
	Crosswalk  string   `json:"crosswalk,omitempty"`
	Target     string   `json:"target,omitempty"`
	SystemHint string   `json:"system_hint,omitempty"`
	Sightings  int      `json:"sightings"`
	Orgs       []string `json:"sighting_organisations,omitempty"`
}

// Index is the machine-readable index of one industry or of shared.
type Index struct {
	Scope        string             `json:"scope"`
	Name         string             `json:"name"`
	Entries      []Entry            `json:"entries"`
	Fingerprints []FingerprintEntry `json:"fingerprints"`
}

var firstSentence = regexp.MustCompile(`^(.*?[.!?])(\s|$)`)

func sentence(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if m := firstSentence.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

func str(m map[string]any, k string) string { s, _ := m[k].(string); return s }

func keywords(doc any) []string {
	props := overlap.Props(doc)
	var k []string
	for name := range props {
		k = append(k, name)
	}
	sort.Strings(k)
	if len(k) > 24 {
		k = k[:24]
	}
	return k
}

func specEntry(r *repo.Repo, p, kind string) (Entry, error) {
	doc, err := r.ReadJSON(p)
	if err != nil {
		return Entry{}, err
	}
	m, _ := doc.(map[string]any)
	return Entry{Kind: kind, ID: str(m, "x-cp-id"), Description: sentence(str(m, "description")), Keywords: keywords(doc),
		Status: str(m, "x-cp-status"), Version: str(m, "x-cp-version"), Path: p, ProfileOf: str(m, "x-cp-profile-of")}, nil
}

func codelistEntry(r *repo.Repo, p string) (Entry, []string, error) {
	doc, err := r.ReadJSON(p)
	if err != nil {
		return Entry{}, nil, err
	}
	m, _ := doc.(map[string]any)
	var tags []string
	if arr, ok := m["industries"].([]any); ok {
		for _, t := range arr {
			if s, ok := t.(string); ok {
				tags = append(tags, s)
			}
		}
	}
	return Entry{Kind: "codelist", ID: "codelists/" + str(m, "id"), Description: sentence(str(m, "description")),
		Status: str(m, "status"), Version: str(m, "version"), Path: p}, tags, nil
}

func (b *builder) fingerprints(scopeDir string) []FingerprintEntry {
	byID := map[string]*FingerprintEntry{}
	for _, cw := range b.r.Glob(scopeDir + "/crosswalks/*/fingerprint.json") {
		raw, err := os.ReadFile(b.r.Abs(cw))
		if err != nil {
			continue
		}
		var fp struct {
			MediaType string `json:"media_type"`
			Columns   []fingerprint.Column
			Hint      *struct {
				Name string `json:"name"`
			} `json:"system_hint"`
		}
		if json.Unmarshal(raw, &fp) != nil {
			continue
		}
		e := &FingerprintEntry{ID: fingerprint.ID(fingerprint.Fingerprint{Columns: fp.Columns}), MediaType: fp.MediaType, Crosswalk: path.Dir(cw) + "/crosswalk.yaml"}
		for _, c := range fp.Columns {
			if c.Optional {
				e.Optional = append(e.Optional, fingerprint.Normalise(c.Name))
			} else {
				e.Required = append(e.Required, fingerprint.Normalise(c.Name))
			}
		}
		if fp.Hint != nil {
			e.SystemHint = fp.Hint.Name
		}
		if cwDoc, err := b.r.ReadYAML(path.Dir(cw) + "/crosswalk.yaml"); err == nil {
			if t, ok := cwDoc.(map[string]any)["target"].(map[string]any); ok {
				e.Target = str(t, "spec")
				if pr := str(t, "profile"); pr != "" {
					e.Target = pr
				}
			}
		}
		byID[e.ID] = e
	}
	for _, s := range b.r.Glob(scopeDir + "/sightings/*.yaml") {
		doc, err := b.r.ReadYAML(s)
		if err != nil {
			continue
		}
		m, _ := doc.(map[string]any)
		fpm, _ := m["fingerprint"].(map[string]any)
		raw, _ := json.Marshal(fpm)
		var fp struct {
			MediaType string `json:"media_type"`
			Columns   []fingerprint.Column
		}
		if json.Unmarshal(raw, &fp) != nil {
			continue
		}
		id := fingerprint.ID(fingerprint.Fingerprint{Columns: fp.Columns})
		e, ok := byID[id]
		if !ok {
			e = &FingerprintEntry{ID: id, MediaType: fp.MediaType, Target: str(m, "target_spec")}
			for _, c := range fp.Columns {
				if c.Optional {
					e.Optional = append(e.Optional, fingerprint.Normalise(c.Name))
				} else {
					e.Required = append(e.Required, fingerprint.Normalise(c.Name))
				}
			}
			byID[id] = e
		}
		e.Sightings++
		if org := str(m, "organisation"); org != "" {
			seen := false
			for _, o := range e.Orgs {
				seen = seen || o == org
			}
			if !seen {
				e.Orgs = append(e.Orgs, org)
				sort.Strings(e.Orgs)
			}
		}
	}
	var out []FingerprintEntry
	for _, e := range byID {
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

type builder struct {
	r   *repo.Repo
	out map[string][]byte
}

func (b *builder) put(p string, data []byte) { b.out[p] = data }

func jsonBytes(v any) []byte {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(v)
	return buf.Bytes()
}

func renderIndexMD(ix Index, title string) []byte {
	var w strings.Builder
	fmt.Fprintf(&w, "# %s\n\n", title)
	w.WriteString("Generated by the build-index workflow (`go -C ci run . build-index`). Do not edit by hand.\n\n")
	w.WriteString("## Specs, profiles and code lists\n\n")
	if len(ix.Entries) == 0 {
		w.WriteString("None yet.\n")
	}
	for _, e := range ix.Entries {
		extra := ""
		if e.ProfileOf != "" {
			extra = " (profile of " + e.ProfileOf + ")"
		}
		fmt.Fprintf(&w, "- %s `%s` %s %s%s: %s Path: `%s`", e.Kind, e.ID, e.Version, e.Status, extra, e.Description, e.Path)
		if len(e.Keywords) > 0 {
			fmt.Fprintf(&w, " Keywords: %s.", strings.Join(e.Keywords, ", "))
		}
		w.WriteString("\n")
	}
	w.WriteString("\n## Known source shapes (fingerprints)\n\n")
	if len(ix.Fingerprints) == 0 {
		w.WriteString("None yet. Map directly to the spec, and contribute a sighting.\n")
	}
	for _, f := range ix.Fingerprints {
		fmt.Fprintf(&w, "- `%s` %s columns: %s", f.ID, f.MediaType, strings.Join(f.Required, ", "))
		if len(f.Optional) > 0 {
			fmt.Fprintf(&w, "; optional: %s", strings.Join(f.Optional, ", "))
		}
		if f.Crosswalk != "" {
			fmt.Fprintf(&w, ". Crosswalk: `%s`", f.Crosswalk)
		}
		if f.Target != "" {
			fmt.Fprintf(&w, ". Target: %s", f.Target)
		}
		if f.SystemHint != "" {
			fmt.Fprintf(&w, ". System hint: %s", f.SystemHint)
		}
		fmt.Fprintf(&w, ". Sightings: %d from %d organisation(s).\n", f.Sightings, len(f.Orgs))
	}
	return []byte(w.String())
}

// Build computes every generated file. Keys are repository-relative paths.
func Build(r *repo.Repo) (map[string][]byte, error) {
	b := &builder{r: r, out: map[string][]byte{}}
	inds, err := r.Industries()
	if err != nil {
		return nil, err
	}
	// Code lists by tag.
	type cl struct {
		e    Entry
		tags []string
	}
	var cls []cl
	for _, p := range r.Glob("codelists/*/codelist.json") {
		e, tags, err := codelistEntry(r, p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		cls = append(cls, cl{e, tags})
	}
	hasTag := func(tags []string, t string) bool {
		for _, x := range tags {
			if x == t {
				return true
			}
		}
		return false
	}
	// Shared index.
	shared := Index{Scope: "shared", Name: "Shared"}
	for _, p := range r.Glob("shared/specs/*/schema.json") {
		e, err := specEntry(r, p, "spec")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		shared.Entries = append(shared.Entries, e)
	}
	for _, c := range cls {
		if hasTag(c.tags, "shared") {
			shared.Entries = append(shared.Entries, c.e)
		}
	}
	shared.Fingerprints = b.fingerprints("shared")
	b.put("index/shared.json", jsonBytes(shared))
	b.put("shared/INDEX.md", renderIndexMD(shared, "Shared index"))

	// Industry indexes.
	var root strings.Builder
	root.WriteString("# Ag Data Skills commons index\n\n")
	root.WriteString("Generated by the build-index workflow. Do not edit by hand.\n\n")
	root.WriteString("Read this, then only the index for the industry you are working in and the shared index. Fetch single files by path; never clone or browse the whole repository.\n\n")
	root.WriteString("- shared: concepts used across industries. Index: `shared/INDEX.md` (`index/shared.json`)\n")
	for _, ind := range inds {
		dir := "industries/" + ind.ID
		ix := Index{Scope: "industry", Name: ind.ID}
		for _, p := range r.Glob(dir + "/specs/*/schema.json") {
			e, err := specEntry(r, p, "spec")
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p, err)
			}
			ix.Entries = append(ix.Entries, e)
		}
		for _, p := range r.Glob(dir + "/profiles/*.json") {
			e, err := specEntry(r, p, "profile")
			if err != nil {
				return nil, fmt.Errorf("%s: %w", p, err)
			}
			ix.Entries = append(ix.Entries, e)
		}
		for _, c := range cls {
			if hasTag(c.tags, ind.ID) {
				ix.Entries = append(ix.Entries, c.e)
			}
		}
		ix.Fingerprints = b.fingerprints(dir)
		status := "No specs or profiles yet: use the shared specs and contribute a profile"
		if len(ix.Entries) > 0 || len(ix.Fingerprints) > 0 {
			status = fmt.Sprintf("%d entries, %d known source shapes", len(ix.Entries), len(ix.Fingerprints))
			b.put("index/"+ind.ID+".json", jsonBytes(ix))
			b.put(dir+"/INDEX.md", renderIndexMD(ix, ind.Name+" ("+ind.ID+") index"))
			fmt.Fprintf(&root, "- %s: %s. %s. Index: `%s/INDEX.md` (`index/%s.json`)\n", ind.ID, ind.Name, status, dir, ind.ID)
		} else {
			fmt.Fprintf(&root, "- %s: %s. %s.\n", ind.ID, ind.Name, status)
		}
	}
	b.put("INDEX.md", []byte(root.String()))

	gov, err := governanceMD(r)
	if err != nil {
		return nil, err
	}
	b.put("GOVERNANCE.md", gov)
	if err := b.changelogs(); err != nil {
		return nil, err
	}
	return b.out, nil
}

func (b *builder) changelogs() error {
	type rec struct {
		ID       string            `yaml:"id"`
		Date     string            `yaml:"date"`
		Summary  string            `yaml:"summary"`
		Versions map[string]string `yaml:"resulting_versions"`
	}
	bySpec := map[string][]rec{}
	for _, p := range b.r.Glob("records/decisions/*.yaml") {
		var d rec
		if err := b.r.ReadConfigInto(p, &d); err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		for spec := range d.Versions {
			bySpec[spec] = append(bySpec[spec], d)
		}
	}
	paths := append(b.r.Glob("shared/specs/*/schema.json"), b.r.Glob("industries/*/specs/*/schema.json")...)
	for _, p := range paths {
		doc, err := b.r.ReadJSON(p)
		if err != nil {
			return err
		}
		id := str(doc.(map[string]any), "x-cp-id")
		recs := bySpec[id]
		sort.Slice(recs, func(i, j int) bool {
			if recs[i].Date != recs[j].Date {
				return recs[i].Date > recs[j].Date
			}
			return recs[i].ID > recs[j].ID
		})
		var w strings.Builder
		fmt.Fprintf(&w, "# Changelog: %s\n\nGenerated from decision records by the build-index workflow. Do not edit by hand.\n", id)
		last := ""
		for _, rc := range recs {
			v := rc.Versions[id]
			if v != last {
				fmt.Fprintf(&w, "\n## %s\n\n", v)
				last = v
			}
			fmt.Fprintf(&w, "- %s (%s): %s\n", rc.ID, rc.Date, sentence(rc.Summary))
		}
		b.put(path.Dir(p)+"/CHANGELOG.md", []byte(w.String()))
	}
	return nil
}

func governanceMD(r *repo.Repo) ([]byte, error) {
	raw, err := os.ReadFile(r.Abs("governance/config.yaml"))
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	var w strings.Builder
	w.WriteString("# Governance\n\nGenerated from `governance/config.yaml` by the build-index workflow. Do not edit by hand. ")
	w.WriteString("Every rule here is executed by the workflows; changing one is a `governance_change` contribution and goes through the same checks and agent reviews as any other. No step waits for a person.\n\n")
	if len(node.Content) > 0 {
		top := node.Content[0]
		for i := 0; i+1 < len(top.Content); i += 2 {
			k, v := top.Content[i], top.Content[i+1]
			if v.Kind != yaml.MappingNode {
				fmt.Fprintf(&w, "- `%s`: %s\n", k.Value, v.Value)
				continue
			}
			fmt.Fprintf(&w, "\n## %s\n\n", k.Value)
			if k.HeadComment != "" {
				w.WriteString(strings.TrimSpace(strings.ReplaceAll(k.HeadComment, "# ", "")) + "\n\n")
			}
			w.WriteString("| Rule | Value |\n|---|---|\n")
			for j := 0; j+1 < len(v.Content); j += 2 {
				val := v.Content[j+1]
				s := val.Value
				if val.Kind == yaml.SequenceNode {
					var items []string
					for _, it := range val.Content {
						items = append(items, it.Value)
					}
					s = strings.Join(items, ", ")
					if s == "" {
						s = "(none)"
					}
				}
				fmt.Fprintf(&w, "| `%s` | %s |\n", v.Content[j].Value, s)
			}
		}
	}
	var tiers struct {
		Tiers []struct {
			ID, Name string
			UseWhen  string `yaml:"use_when"`
			Weight   float64
		}
	}
	if err := r.ReadConfigInto("governance/evidence-tiers.yaml", &tiers); err == nil {
		w.WriteString("\n## Evidence tiers\n\nAgents record a tier and a reference, never a numeric confidence. Scores are the tier weights.\n\n| Tier | Name | Use when | Weight |\n|---|---|---|---|\n")
		for _, t := range tiers.Tiers {
			fmt.Fprintf(&w, "| %s | %s | %s | %.2f |\n", t.ID, t.Name, t.UseWhen, t.Weight)
		}
	}
	inds, _ := r.Industries()
	w.WriteString("\n## Industries\n\n| Id | Industry | RDC |\n|---|---|---|\n")
	for _, i := range inds {
		fmt.Fprintf(&w, "| `%s` | %s | %s |\n", i.ID, i.Name, i.RDC)
	}
	return []byte(w.String()), nil
}

// Write writes generated files to disk and returns the paths that changed.
func Write(r *repo.Repo, files map[string][]byte) ([]string, error) {
	var changed []string
	for _, p := range sortedKeys(files) {
		old, _ := os.ReadFile(r.Abs(p))
		if bytes.Equal(old, files[p]) {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(r.Abs(p)), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(r.Abs(p), files[p], 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, p)
	}
	return changed, nil
}

// Stale returns the generated paths whose content on disk differs.
func Stale(r *repo.Repo, files map[string][]byte) []string {
	var out []string
	for _, p := range sortedKeys(files) {
		old, _ := os.ReadFile(r.Abs(p))
		if !bytes.Equal(old, files[p]) {
			out = append(out, p)
		}
	}
	return out
}

func sortedKeys(m map[string][]byte) []string {
	var k []string
	for p := range m {
		k = append(k, p)
	}
	sort.Strings(k)
	return k
}

// FillDecisionCommits sets commit: null in decision records to the commit that added them.
func FillDecisionCommits(r *repo.Repo) ([]string, error) {
	var changed []string
	re := regexp.MustCompile(`(?m)^commit: null$`)
	for _, p := range r.Glob("records/decisions/*.yaml") {
		raw, err := os.ReadFile(r.Abs(p))
		if err != nil {
			return nil, err
		}
		if !re.Match(raw) {
			continue
		}
		h := r.FirstCommit(p)
		if h == "" {
			continue
		}
		raw = re.ReplaceAll(raw, []byte("commit: "+h))
		if err := os.WriteFile(r.Abs(p), raw, 0o644); err != nil {
			return nil, err
		}
		changed = append(changed, p)
	}
	return changed, nil
}
