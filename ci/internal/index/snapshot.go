package index

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
)

// SnapshotManifest is written to SNAPSHOT.json at the root of the snapshot.
type SnapshotManifest struct {
	Repository string   `json:"repository"`
	Commit     string   `json:"commit"`
	CommitDate string   `json:"commit_date"`
	Files      []string `json:"files"`
	Bytes      int64    `json:"bytes"`
	Note       string   `json:"note"`
}

func excluded(p string, patterns []string) bool {
	for _, pat := range patterns {
		pat = strings.Trim(pat, "/")
		if strings.HasPrefix(pat, "*/") {
			if strings.Contains("/"+p+"/", "/"+strings.TrimPrefix(pat, "*/")+"/") {
				return true
			}
			continue
		}
		if ok, _ := path.Match(pat, p); ok {
			return true
		}
	}
	return false
}

// SnapshotFiles lists the repository files the snapshot includes (config.snapshot).
func SnapshotFiles(r *repo.Repo) ([]string, error) {
	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}
	all, err := r.Files()
	if err != nil {
		return nil, err
	}
	var out []string
	for _, f := range all {
		if path.Base(f) == ".gitkeep" || excluded(f, cfg.Snapshot.Exclude) {
			continue
		}
		for _, inc := range cfg.Snapshot.Include {
			if f == inc || (strings.HasSuffix(inc, "/") && strings.HasPrefix(f, inc)) {
				out = append(out, f)
				break
			}
		}
	}
	sort.Strings(out)
	return out, nil
}

// BuildSnapshot replaces outDir with a fresh copy of the snapshot files and SNAPSHOT.json.
func BuildSnapshot(r *repo.Repo, outDir string) (*SnapshotManifest, error) {
	cfg, err := r.Config()
	if err != nil {
		return nil, err
	}
	files, err := SnapshotFiles(r)
	if err != nil {
		return nil, err
	}
	abs := r.Abs(outDir)
	if !strings.HasPrefix(abs, r.Root+string(filepath.Separator)) || !strings.HasSuffix(outDir, "snapshot") {
		return nil, fmt.Errorf("refusing to replace %s: the snapshot directory must be inside the repository and named snapshot", outDir)
	}
	if err := os.RemoveAll(abs); err != nil {
		return nil, err
	}
	m := &SnapshotManifest{Repository: "morethanmachines/ag-data-skills", Commit: r.GitHead(), Files: files,
		Note: "Bundled snapshot of the commons: indexes, specs, profiles, code lists, crosswalks and the external registry. Not fixtures, tasks or records."}
	cmd := exec.Command("git", "log", "-1", "--format=%cs")
	cmd.Dir = r.Root
	if out, err := cmd.Output(); err == nil {
		m.CommitDate = strings.TrimSpace(string(out))
	}
	for _, f := range files {
		b, err := os.ReadFile(r.Abs(f))
		if err != nil {
			return nil, err
		}
		dst := filepath.Join(abs, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(dst, b, 0o644); err != nil {
			return nil, err
		}
		m.Bytes += int64(len(b))
	}
	if cfg.Snapshot.MaxBytes > 0 && m.Bytes > cfg.Snapshot.MaxBytes {
		return m, fmt.Errorf("snapshot is %d bytes, over snapshot.max_bytes %d: split it by industry (plan 4.5)", m.Bytes, cfg.Snapshot.MaxBytes)
	}
	raw, _ := json.MarshalIndent(m, "", "  ")
	return m, os.WriteFile(filepath.Join(abs, "SNAPSHOT.json"), append(raw, '\n'), 0o644)
}
