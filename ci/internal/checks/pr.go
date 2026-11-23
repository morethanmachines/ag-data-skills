package checks

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/overlap"
	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

var fenced = regexp.MustCompile("(?s)```ya?ml\\s*\\n(.*?)```")

// ExtractBlock finds the first fenced YAML block in text whose top-level key is key.
func ExtractBlock(text, key string) (map[string]any, error) {
	for _, m := range fenced.FindAllStringSubmatch(text, -1) {
		v, err := repo.DecodeYAML([]byte(m[1]))
		if err != nil {
			continue
		}
		if mm, ok := v.(map[string]any); ok {
			if _, ok := mm[key]; ok {
				return mm, nil
			}
		}
	}
	return nil, fmt.Errorf("no fenced yaml block with %q found", key)
}

// FamilyOf maps a declared model id to its family using governance/model-families.yaml.
func FamilyOf(r *repo.Repo, model string) string {
	var mf struct {
		Families []struct {
			ID       string   `yaml:"id"`
			Patterns []string `yaml:"patterns"`
		} `yaml:"families"`
	}
	if err := r.ReadConfigInto("governance/model-families.yaml", &mf); err != nil {
		return "other"
	}
	m := strings.ToLower(model)
	for _, f := range mf.Families {
		for _, p := range f.Patterns {
			if ok, _ := path.Match(p, m); ok {
				return f.ID
			}
			// path.Match treats '/' specially; also try against the part after the provider.
			if i := strings.IndexByte(m, '/'); i >= 0 {
				if ok, _ := path.Match(p, m[i+1:]); ok {
					return f.ID
				}
			}
		}
	}
	return "other"
}

// RunPR runs the checks that depend on a pull request: its manifest and the files it adds.
// RunAll must have been called first.
func (c *Checker) RunPR(base, body string) []Finding {
	var ctype string
	man, err := ExtractBlock(body, "cp-contribution")
	if err != nil {
		c.fail("PR body", "manifest", "%v (use .github/PULL_REQUEST_TEMPLATE.md)", err)
	} else {
		errs, err := c.C.Validate("governance/schemas/manifest.schema.json", man)
		if err != nil {
			c.fail("PR body", "manifest", "manifest schema: %v", err)
		}
		for _, e := range errs {
			c.fail("PR body", "manifest", "manifest is invalid %s", e)
		}
		cm, _ := man["cp-contribution"].(map[string]any)
		ctype, _ = cm["type"].(string)
		for k, x := range cm {
			if strings.Contains(strings.ToLower(k), "confidence") {
				_ = x
				c.fail("PR body", "numeric-confidence", "manifest field %q: numeric confidence is not allowed", k)
			}
		}
	}
	added, err := c.R.AddedFiles(base)
	if err != nil {
		c.fail("PR", "git", "%v", err)
		return c.Findings()
	}
	var newSpecs []string
	for _, f := range added {
		if id := specID(f); id != "" && strings.HasSuffix(f, "/schema.json") {
			newSpecs = append(newSpecs, f)
		}
	}
	if len(newSpecs) > 0 && ctype != "" && ctype != "spec_new" && ctype != "governance_change" {
		c.fail("PR body", "manifest", "the PR adds a new spec, so its type must be spec_new (got %s)", ctype)
	}
	for _, sp := range newSpecs {
		dir := path.Dir(sp)
		if !c.R.Exists(dir + "/external-search.yaml") {
			c.fail(dir, "external-search", "a new spec must include external-search.yaml recording the searches run and why no external standard fitted")
		}
		c.overlapCheck(sp)
	}
	return c.Findings()
}

// overlapCheck fails a new draft that overlaps an existing spec above the threshold.
func (c *Checker) overlapCheck(sp string) {
	doc, err := c.R.ReadJSON(sp)
	if err != nil {
		return
	}
	draft := overlap.Props(doc)
	existing := map[string]map[string]string{}
	for _, other := range c.specPaths() {
		if other == sp {
			continue
		}
		od, err := c.R.ReadJSON(other)
		if err != nil {
			continue
		}
		existing[specID(other)] = overlap.Props(od)
	}
	ranked := overlap.Rank(draft, existing)
	if len(ranked) > 0 && ranked[0].Score > c.Cfg.Drafts.OverlapThreshold {
		c.fail(sp, "overlap", "overlaps %s (score %.2f, threshold %.2f): contribute a crosswalk, profile or extension to %s instead",
			ranked[0].Spec, ranked[0].Score, c.Cfg.Drafts.OverlapThreshold, ranked[0].Spec)
	}
}
