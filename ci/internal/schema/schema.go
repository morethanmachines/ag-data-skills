// Package schema compiles commons JSON Schemas (2020-12) and validates instances,
// resolving $ids under repo.BaseURL to files in the repository so nothing is fetched.
package schema

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/morethanmachines/ag-data-skills/ci/internal/repo"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type loader struct{ r *repo.Repo }

func (l loader) Load(u string) (any, error) {
	if strings.HasPrefix(u, repo.BaseURL) {
		rel := strings.TrimPrefix(u, repo.BaseURL)
		if i := strings.IndexByte(rel, '#'); i >= 0 {
			rel = rel[:i]
		}
		return l.r.ReadJSON(rel)
	}
	if strings.HasPrefix(u, "file://") {
		p, err := url.Parse(u)
		if err != nil {
			return nil, err
		}
		return jsonschema.FileLoader{}.Load(p.String())
	}
	return nil, fmt.Errorf("refusing to fetch %s: commons schemas must resolve inside the repository", u)
}

// Compiler compiles schemas from a repository.
type Compiler struct {
	r     *repo.Repo
	cache map[string]*jsonschema.Schema
}

// New returns a compiler for the repository.
func New(r *repo.Repo) *Compiler {
	return &Compiler{r: r, cache: map[string]*jsonschema.Schema{}}
}

func (c *Compiler) newCompiler() *jsonschema.Compiler {
	jc := jsonschema.NewCompiler()
	jc.DefaultDraft(jsonschema.Draft2020)
	jc.AssertFormat()
	jc.UseLoader(loader{c.r})
	return jc
}

// URLFor returns the canonical URL of a repository file.
func URLFor(rel string) string { return repo.BaseURL + rel }

// Compile compiles the schema stored at a repository-relative path.
func (c *Compiler) Compile(rel string) (*jsonschema.Schema, error) {
	if s, ok := c.cache[rel]; ok {
		return s, nil
	}
	doc, err := c.r.ReadJSON(rel)
	if err != nil {
		return nil, err
	}
	u := URLFor(rel)
	if m, ok := doc.(map[string]any); ok {
		if id, ok := m["$id"].(string); ok && id != u {
			return nil, fmt.Errorf("$id %q does not match its path (want %q)", id, u)
		}
	}
	jc := c.newCompiler()
	if err := jc.AddResource(u, doc); err != nil {
		return nil, err
	}
	s, err := jc.Compile(u)
	if err != nil {
		return nil, err
	}
	c.cache[rel] = s
	return s, nil
}

// Validate validates an instance against the schema at rel and returns a list of
// readable error messages (empty if valid).
func (c *Compiler) Validate(rel string, inst any) ([]string, error) {
	s, err := c.Compile(rel)
	if err != nil {
		return nil, err
	}
	if err := s.Validate(inst); err != nil {
		if ve, ok := err.(*jsonschema.ValidationError); ok {
			return flatten(ve), nil
		}
		return []string{err.Error()}, nil
	}
	return nil, nil
}

var printer = message.NewPrinter(language.English)

func flatten(ve *jsonschema.ValidationError) []string {
	var out []string
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) == 0 {
			loc := "/" + strings.Join(e.InstanceLocation, "/")
			out = append(out, fmt.Sprintf("at %s: %s", loc, e.ErrorKind.LocalizedString(printer)))
			return
		}
		for _, c := range e.Causes {
			walk(c)
		}
	}
	walk(ve)
	if len(out) > 10 {
		out = append(out[:10], fmt.Sprintf("and %d more", len(out)-10))
	}
	return out
}
