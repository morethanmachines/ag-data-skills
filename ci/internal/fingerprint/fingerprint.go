// Package fingerprint normalises source-shape column names, computes fingerprint ids and
// scores how well a file's columns match a known fingerprint (implementation plan, 4.1).
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Normalise lower-cases a column name, replaces '%' with 'pct' and '°' with nothing,
// collapses every run of non-alphanumerics to a single underscore and trims underscores.
func Normalise(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = strings.ReplaceAll(s, "%", " pct ")
	s = strings.ReplaceAll(s, "°", "")
	s = nonAlnum.ReplaceAllString(s, "_")
	return strings.Trim(s, "_")
}

// Column is a fingerprint column.
type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Pattern  string `json:"pattern,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// Fingerprint is the subset of fingerprint.json that identity and matching depend on.
type Fingerprint struct {
	Columns []Column `json:"columns"`
}

// ID is "fp-" plus the first 12 hex characters of the SHA-256 of the sorted,
// newline-joined normalised names of the required columns.
func ID(fp Fingerprint) string {
	var req []string
	for _, c := range fp.Columns {
		if !c.Optional {
			req = append(req, Normalise(c.Name))
		}
	}
	sort.Strings(req)
	sum := sha256.Sum256([]byte(strings.Join(req, "\n")))
	return "fp-" + hex.EncodeToString(sum[:])[:12]
}

// Match scores a file's columns against a fingerprint, from 0 to 1. The column score is
// matched / (required columns plus the file's columns that the fingerprint does not know),
// so missing optional columns do not count against a file but unknown and missing required
// columns do. It is multiplied by the share of matched columns whose value patterns agree,
// counting only columns where both sides state a pattern.
func Match(fileCols []Column, fp Fingerprint) float64 {
	known := map[string]Column{}
	required := map[string]bool{}
	for _, c := range fp.Columns {
		n := Normalise(c.Name)
		known[n] = c
		if !c.Optional {
			required[n] = true
		}
	}
	file := map[string]Column{}
	for _, c := range fileCols {
		file[Normalise(c.Name)] = c
	}
	matched, unknown := 0, 0
	compared, agree := 0, 0
	for n, fc := range file {
		if kc, ok := known[n]; ok {
			matched++
			if fc.Pattern != "" && kc.Pattern != "" {
				compared++
				if fc.Pattern == kc.Pattern {
					agree++
				}
			}
		} else {
			unknown++
		}
	}
	denom := len(required) + unknown
	for n := range file {
		if k, ok := known[n]; ok && k.Optional {
			denom++
		}
	}
	if denom == 0 {
		return 0
	}
	score := float64(matched) / float64(denom)
	if compared > 0 {
		score *= float64(agree) / float64(compared)
	}
	return score
}
