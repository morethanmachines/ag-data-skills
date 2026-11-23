// Package repo locates the commons repository and reads its files.
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

// BaseURL is the prefix of every $id in the commons. It maps to the repository root.
const BaseURL = "https://raw.githubusercontent.com/morethanmachines/ag-data-skills/main/"

// Repo is a checked-out commons repository.
type Repo struct {
	Root string
}

// Find walks up from dir until it finds governance/config.yaml.
func Find(dir string) (*Repo, error) {
	d, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "governance", "config.yaml")); err == nil {
			return &Repo{Root: d}, nil
		}
		parent := filepath.Dir(d)
		if parent == d {
			return nil, fmt.Errorf("no commons repository (governance/config.yaml) above %s", dir)
		}
		d = parent
	}
}

// Abs returns the absolute path of a repository-relative path.
func (r *Repo) Abs(rel string) string { return filepath.Join(r.Root, filepath.FromSlash(rel)) }

// Exists reports whether a repository-relative path exists.
func (r *Repo) Exists(rel string) bool {
	_, err := os.Stat(r.Abs(rel))
	return err == nil
}

// skipDirs are never part of the commons content.
var skipDirs = map[string]bool{".git": true, "private": true, "dist": true, "node_modules": true}

// IsSkipped reports whether a repository-relative path is outside the commons content
// (private material, build output, the generated snapshot).
func IsSkipped(rel string) bool {
	first := strings.SplitN(rel, "/", 2)[0]
	if skipDirs[first] {
		return true
	}
	return strings.HasPrefix(rel, "skills/ag-data-commons/snapshot/") || strings.HasPrefix(rel, "ci/bin/")
}

// Files lists every repository-relative file path (slash-separated), sorted.
func (r *Repo) Files() ([]string, error) {
	var out []string
	err := filepath.WalkDir(r.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(r.Root, p)
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if IsSkipped(rel) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			out = append(out, rel)
		}
		return nil
	})
	sort.Strings(out)
	return out, err
}

// Glob returns sorted repository-relative paths matching a slash pattern.
func (r *Repo) Glob(pattern string) []string {
	m, _ := filepath.Glob(r.Abs(pattern))
	var out []string
	for _, p := range m {
		rel, _ := filepath.Rel(r.Root, p)
		out = append(out, filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out
}

// ChangedFiles lists files added or modified relative to base (merge base with HEAD).
// Deleted files are not returned.
func (r *Repo) ChangedFiles(base string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=ACMR", base+"...HEAD")
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff against %s: %w", base, err)
	}
	var files []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" && !IsSkipped(l) {
			files = append(files, l)
		}
	}
	return files, nil
}

// AddedFiles lists files added relative to base.
func (r *Repo) AddedFiles(base string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "--diff-filter=A", base+"...HEAD")
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff against %s: %w", base, err)
	}
	var files []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			files = append(files, l)
		}
	}
	return files, nil
}

// ReadJSON reads a JSON file into a value suitable for jsonschema validation.
func (r *Repo) ReadJSON(rel string) (any, error) {
	b, err := os.ReadFile(r.Abs(rel))
	if err != nil {
		return nil, err
	}
	return DecodeJSON(b)
}

// DecodeJSON decodes JSON bytes, keeping numbers as json.Number.
func DecodeJSON(b []byte) (any, error) {
	return jsonschema.UnmarshalJSON(bytes.NewReader(b))
}

// ReadYAML reads a YAML file and converts it to the same value model as ReadJSON.
func (r *Repo) ReadYAML(rel string) (any, error) {
	b, err := os.ReadFile(r.Abs(rel))
	if err != nil {
		return nil, err
	}
	return DecodeYAML(b)
}

// DecodeYAML decodes YAML into JSON-compatible values (string keys, json.Number numbers).
func DecodeYAML(b []byte) (any, error) {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return nil, err
	}
	j, err := json.Marshal(normalise(v))
	if err != nil {
		return nil, err
	}
	return DecodeJSON(j)
}

func normalise(v any) any {
	switch t := v.(type) {
	case map[string]any:
		m := make(map[string]any, len(t))
		for k, x := range t {
			m[k] = normalise(x)
		}
		return m
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, x := range t {
			m[fmt.Sprint(k)] = normalise(x)
		}
		return m
	case []any:
		for i := range t {
			t[i] = normalise(t[i])
		}
		return t
	default:
		return v
	}
}

// ReadAny reads a JSON or YAML file by extension.
func (r *Repo) ReadAny(rel string) (any, error) {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".yaml", ".yml":
		return r.ReadYAML(rel)
	default:
		return r.ReadJSON(rel)
	}
}

// ReadConfigInto unmarshals a YAML file into a Go struct.
func (r *Repo) ReadConfigInto(rel string, out any) error {
	b, err := os.ReadFile(r.Abs(rel))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, out)
}

// GitHead returns the commit hash of HEAD, or "" if unavailable.
func (r *Repo) GitHead() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// FirstCommit returns the first commit that added a path, or "".
func (r *Repo) FirstCommit(rel string) string {
	cmd := exec.Command("git", "log", "--diff-filter=A", "--format=%H", "--", rel)
	cmd.Dir = r.Root
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines[len(lines)-1]
}
