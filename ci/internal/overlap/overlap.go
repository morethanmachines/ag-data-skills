// Package overlap measures how much a draft spec duplicates an existing spec.
// The score is the share of the draft's meaningful property
// names that an existing spec already has, with half credit when the kinds differ.
package overlap

import (
	"sort"

	"github.com/morethanmachines/ag-data-skills/ci/internal/fingerprint"
)

// structural names that say nothing about the concept a spec covers.
var stop = map[string]bool{
	"value": true, "unit": true, "id": true, "name": true, "type": true, "text": true, "code": true,
	"scheme": true, "label": true, "description": true, "notes": true, "status": true, "version": true,
	"spec": true, "spec_version": true, "id_scheme": true, "comment": true, "comments": true, "basis": true,
}

// Props collects normalised property names (at any depth) and their x-cp-kind.
func Props(schema any) map[string]string {
	out := map[string]string{}
	var walk func(n any)
	walk = func(n any) {
		switch t := n.(type) {
		case map[string]any:
			if props, ok := t["properties"].(map[string]any); ok {
				for k, v := range props {
					nk := fingerprint.Normalise(k)
					if stop[nk] {
						continue
					}
					kind := ""
					if m, ok := v.(map[string]any); ok {
						kind, _ = m["x-cp-kind"].(string)
					}
					if _, seen := out[nk]; !seen || out[nk] == "" {
						out[nk] = kind
					}
				}
			}
			for _, v := range t {
				walk(v)
			}
		case []any:
			for _, v := range t {
				walk(v)
			}
		}
	}
	walk(schema)
	return out
}

// Score returns the overlap of draft with existing, from 0 to 1.
func Score(draft, existing map[string]string) float64 {
	if len(draft) == 0 {
		return 0
	}
	var s float64
	for name, kind := range draft {
		if ek, ok := existing[name]; ok {
			if kind == "" || ek == "" || kind == ek {
				s += 1
			} else {
				s += 0.5
			}
		}
	}
	return s / float64(len(draft))
}

// Result is the overlap with one existing spec.
type Result struct {
	Spec  string
	Score float64
}

// Rank scores a draft against every existing spec, highest first.
func Rank(draft map[string]string, existing map[string]map[string]string) []Result {
	var out []Result
	for id, props := range existing {
		out = append(out, Result{Spec: id, Score: Score(draft, props)})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Spec < out[j].Spec
	})
	return out
}
