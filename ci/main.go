// Command ci implements the deterministic checks of the Ag Data Skills commons. The
// GitHub Actions workflows call it; contributors' agents can run the same commands locally.
//
//	go -C ci run . checks [--base origin/main] [--body-file pr.md]
//	go -C ci run . scrub [--base origin/main | --all] [files...]
//	go -C ci run . build-index [--check] [--fill-commits]
//	go -C ci run . snapshot [--out skills/ag-data-commons/snapshot]
//	go -C ci run . fingerprint-id path/to/fingerprint.json
//	go -C ci run . match path/to/file.csv
//	go -C ci run . family provider/model
package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/checks"
	"github.com/morethanmachines/ag-data-skills/ci/internal/fingerprint"
	"github.com/morethanmachines/ag-data-skills/ci/internal/index"
	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
	"github.com/morethanmachines/ag-data-skills/ci/internal/scrub"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cmd, args := os.Args[1], os.Args[2:]
	var err error
	code := 0
	switch cmd {
	case "checks", "validate":
		code, err = runChecks(args)
	case "scrub":
		code, err = runScrub(args)
	case "build-index":
		code, err = runBuildIndex(args)
	case "snapshot":
		err = runSnapshot(args)
	case "fingerprint-id":
		err = runFingerprintID(args)
	case "match":
		err = runMatch(args)
	case "family":
		err = runFamily(args)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(2)
	}
	os.Exit(code)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: ci <checks|scrub|build-index|snapshot|fingerprint-id|match|family> [flags]")
	os.Exit(2)
}

func findRepo(root string) (*repo.Repo, error) {
	if root != "" {
		return repo.Find(root)
	}
	wd, _ := os.Getwd()
	return repo.Find(wd)
}

func annotate(path, msg string) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		file := path
		if strings.Contains(file, " ") {
			file = ""
		}
		fmt.Printf("::error file=%s::%s\n", file, strings.ReplaceAll(msg, "\n", " "))
	}
}

func writeJSON(p string, v any) error {
	if p == "" {
		return nil
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return os.WriteFile(p, append(b, '\n'), 0o644)
}

func runChecks(args []string) (int, error) {
	fs := flag.NewFlagSet("checks", flag.ExitOnError)
	root := fs.String("root", "", "repository root (default: search upwards)")
	base := fs.String("base", "", "base ref for pull request checks")
	body := fs.String("body-file", "", "file containing the pull request body")
	out := fs.String("json", "", "write findings as JSON to this file")
	fs.Parse(args)
	r, err := findRepo(*root)
	if err != nil {
		return 0, err
	}
	c, err := checks.New(r)
	if err != nil {
		return 0, err
	}
	findings := c.RunAll()
	if *base != "" {
		text := ""
		if *body != "" {
			b, err := os.ReadFile(*body)
			if err != nil {
				return 0, err
			}
			text = string(b)
		}
		findings = c.RunPR(*base, text)
	}
	if err := writeJSON(*out, findings); err != nil {
		return 0, err
	}
	for _, f := range findings {
		fmt.Println(f)
		annotate(f.Path, "["+f.Check+"] "+f.Msg)
	}
	if len(findings) > 0 {
		fmt.Printf("checks: %d problem(s)\n", len(findings))
		return 1, nil
	}
	fmt.Println("checks: all passed")
	return 0, nil
}

func runScrub(args []string) (int, error) {
	fs := flag.NewFlagSet("scrub", flag.ExitOnError)
	root := fs.String("root", "", "repository root holding the content to scan")
	rules := fs.String("rules", "", "repository root whose governance rules to apply (default: --root); in pull_request_target workflows this is the trusted base checkout")
	base := fs.String("base", "", "scan files changed since this ref")
	all := fs.Bool("all", false, "scan every file in the repository")
	out := fs.String("json", "", "write findings as JSON to this file")
	fs.Parse(args)
	r, err := findRepo(*root)
	if err != nil {
		return 0, err
	}
	rr := r
	if *rules != "" {
		if rr, err = repo.Find(*rules); err != nil {
			return 0, err
		}
	}
	s, err := scrub.NewScanner(rr)
	if err != nil {
		return 0, err
	}
	var files []string
	switch {
	case len(fs.Args()) == 1 && strings.HasPrefix(fs.Arg(0), "@"):
		raw, rerr := os.ReadFile(strings.TrimPrefix(fs.Arg(0), "@"))
		if rerr != nil {
			return 0, rerr
		}
		for _, l := range strings.Split(string(raw), "\n") {
			if l = strings.TrimSpace(l); l != "" && !repo.IsSkipped(l) {
				files = append(files, l)
			}
		}
	case len(fs.Args()) > 0:
		files = fs.Args()
	case *base != "":
		files, err = r.ChangedFiles(*base)
	default:
		_ = all
		files, err = r.Files()
	}
	if err != nil {
		return 0, err
	}
	var findings []scrub.Finding
	for _, f := range files {
		rel := f
		if filepath.IsAbs(f) {
			rel, _ = filepath.Rel(r.Root, f)
		}
		rel = filepath.ToSlash(rel)
		b, err := os.ReadFile(r.Abs(rel))
		if err != nil {
			continue
		}
		findings = append(findings, s.ScanFile(rel, b)...)
	}
	if err := writeJSON(*out, findings); err != nil {
		return 0, err
	}
	for _, f := range findings {
		fmt.Println(f)
		annotate(f.Path, "["+f.Kind+"] "+f.Detail)
	}
	if len(findings) > 0 {
		fmt.Printf("scrub: %d possible leak(s) in %d file(s) scanned\n", len(findings), len(files))
		return 1, nil
	}
	fmt.Printf("scrub: %d file(s) scanned, nothing found\n", len(files))
	return 0, nil
}

func runBuildIndex(args []string) (int, error) {
	fs := flag.NewFlagSet("build-index", flag.ExitOnError)
	root := fs.String("root", "", "repository root")
	check := fs.Bool("check", false, "report stale generated files instead of writing them")
	fill := fs.Bool("fill-commits", false, "fill commit: null in decision records")
	fs.Parse(args)
	r, err := findRepo(*root)
	if err != nil {
		return 0, err
	}
	files, err := index.Build(r)
	if err != nil {
		return 0, err
	}
	if *check {
		stale := index.Stale(r, files)
		for _, p := range stale {
			fmt.Println("stale:", p)
		}
		if len(stale) > 0 {
			fmt.Println("build-index: generated files are out of date; run go -C ci run . build-index")
			return 1, nil
		}
		fmt.Println("build-index: generated files are up to date")
		return 0, nil
	}
	if *fill {
		changed, err := index.FillDecisionCommits(r)
		if err != nil {
			return 0, err
		}
		for _, p := range changed {
			fmt.Println("filled commit:", p)
		}
	}
	changed, err := index.Write(r, files)
	if err != nil {
		return 0, err
	}
	for _, p := range changed {
		fmt.Println("wrote:", p)
	}
	fmt.Printf("build-index: %d generated file(s), %d changed\n", len(files), len(changed))
	return 0, nil
}

func runSnapshot(args []string) error {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	root := fs.String("root", "", "repository root")
	out := fs.String("out", "skills/ag-data-commons/snapshot", "snapshot directory (repository-relative)")
	fs.Parse(args)
	r, err := findRepo(*root)
	if err != nil {
		return err
	}
	m, err := index.BuildSnapshot(r, *out)
	if err != nil {
		return err
	}
	fmt.Printf("snapshot: %d files, %d bytes, built from %s\n", len(m.Files), m.Bytes, m.Commit)
	return nil
}

func runFingerprintID(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: fingerprint-id path/to/fingerprint.json")
	}
	b, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	var fp fingerprint.Fingerprint
	if err := json.Unmarshal(b, &fp); err != nil {
		return err
	}
	fmt.Println(fingerprint.ID(fp))
	return nil
}

func runMatch(args []string) error {
	fs := flag.NewFlagSet("match", flag.ExitOnError)
	root := fs.String("root", "", "repository root")
	fs.Parse(args)
	if fs.NArg() != 1 {
		return fmt.Errorf("usage: match path/to/file.csv")
	}
	r, err := findRepo(*root)
	if err != nil {
		return err
	}
	f, err := os.Open(fs.Arg(0))
	if err != nil {
		return err
	}
	defer f.Close()
	header, err := csv.NewReader(f).Read()
	if err != nil {
		return err
	}
	var cols []fingerprint.Column
	for _, h := range header {
		cols = append(cols, fingerprint.Column{Name: h})
	}
	type hit struct {
		id, crosswalk string
		score         float64
	}
	var hits []hit
	for _, p := range r.Glob("index/*.json") {
		raw, _ := os.ReadFile(r.Abs(p))
		var ix index.Index
		if json.Unmarshal(raw, &ix) != nil {
			continue
		}
		for _, e := range ix.Fingerprints {
			var fpc []fingerprint.Column
			for _, n := range e.Required {
				fpc = append(fpc, fingerprint.Column{Name: n})
			}
			for _, n := range e.Optional {
				fpc = append(fpc, fingerprint.Column{Name: n, Optional: true})
			}
			hits = append(hits, hit{e.ID, e.Crosswalk, fingerprint.Match(cols, fingerprint.Fingerprint{Columns: fpc})})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	cfg, _ := r.Config()
	fmt.Printf("file columns: %s\n", strings.Join(func() []string {
		var n []string
		for _, c := range cols {
			n = append(n, fingerprint.Normalise(c.Name))
		}
		return n
	}(), ", "))
	if len(hits) == 0 {
		fmt.Println("no known fingerprints; map directly to the spec")
	}
	for _, h := range hits {
		use := ""
		if h.score >= cfg.Crosswalks.FingerprintMatchThreshold {
			use = "  <- above threshold, use this crosswalk"
		}
		fmt.Printf("%.2f %s %s%s\n", h.score, h.id, h.crosswalk, use)
	}
	return nil
}

func runFamily(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: family provider/model")
	}
	r, err := findRepo("")
	if err != nil {
		return err
	}
	fmt.Println(checks.FamilyOf(r, args[0]))
	return nil
}
