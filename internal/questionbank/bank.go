// Package questionbank validates curated taxonomies and ingests legacy quiz JSON.
package questionbank

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Entry struct {
	ID         string            `json:"id"`
	Label      string            `json:"label"`
	Aliases    []string          `json:"aliases,omitempty"`
	Categories []string          `json:"categories,omitempty"`
	Provenance []EntryProvenance `json:"provenance,omitempty"`
}

// EntryProvenance records evidence for a specific curated label or relationship.
// Source categories are classifications, never exam blueprint domains.
type EntryProvenance struct {
	Source       string `json:"source"`
	SourceCommit string `json:"source_commit,omitempty"`
	License      string `json:"license"`
	SourceLine   int    `json:"source_line,omitempty"`
	Modification string `json:"modification"`
}
type Exception struct {
	Label string `json:"label"`
	Kind  string `json:"kind"`
}
type Attribution struct {
	Source           string  `json:"source"`
	Creator          string  `json:"creator"`
	URL              string  `json:"url"`
	License          string  `json:"license"`
	LicenseURL       string  `json:"license_url"`
	AdaptationNotice string  `json:"adaptation_notice"`
	SourceCommit     *string `json:"source_commit,omitempty"`
}
type Taxonomy struct {
	ID           string      `json:"id"`
	Version      string      `json:"version"`
	Attribution  Attribution `json:"attribution"`
	Domains      []Entry     `json:"domains"`
	Technologies []Entry     `json:"technologies"`
	Exceptions   []Exception `json:"classification_exceptions,omitempty"`
}
type Task struct {
	Tag         json.RawMessage `json:"tag"`
	Title       string          `json:"title"`
	Domain      string          `json:"domain"`
	Explanation string          `json:"explanation"`
	Answer      []struct {
		Text string `json:"answerText"`
	} `json:"answer"`
	Options []struct {
		Text string `json:"option"`
	} `json:"options"`
	Products []string `json:"products"`
}
type Quiz struct {
	Author    string `json:"author,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Subject   string `json:"subject"`
	Tasks     []Task `json:"tasks"`
}
type Source struct {
	File   string          `json:"file"`
	QuizID string          `json:"quiz_id"`
	Tag    json.RawMessage `json:"tag"`
}
type Question struct {
	ID            string   `json:"id"`
	Source        Source   `json:"source"`
	DomainID      string   `json:"domain_id"`
	TechnologyIDs []string `json:"technology_ids"`
	TopicLabels   []string `json:"topic_labels,omitempty"`
	Task          Task     `json:"task"`
}
type Bank struct {
	SchemaVersion   int        `json:"schema_version"`
	ID              string     `json:"id"`
	TaxonomyID      string     `json:"taxonomy_id"`
	TaxonomyVersion string     `json:"taxonomy_version"`
	Provider        string     `json:"provider"`
	Subject         string     `json:"subject"`
	Questions       []Question `json:"questions"`
}

func readJSON(path string, dst any) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(dst); e != nil {
		return fmt.Errorf("%s: %w", path, e)
	}
	var extra any
	if e = d.Decode(&extra); e == nil {
		return fmt.Errorf("%s: trailing JSON", path)
	}
	return nil
}
func key(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
func index(entries []Entry, name string) (map[string]string, error) {
	m := map[string]string{}
	ids := map[string]bool{}
	for _, e := range entries {
		if e.ID == "" || e.Label == "" {
			return nil, fmt.Errorf("%s entry requires id and label", name)
		}
		if ids[e.ID] {
			return nil, fmt.Errorf("duplicate %s id %q", name, e.ID)
		}
		ids[e.ID] = true
		categories := map[string]bool{}
		for _, c := range e.Categories {
			k := key(c)
			if k == "" || categories[k] {
				return nil, fmt.Errorf("invalid or duplicate category on %s %q", name, e.ID)
			}
			categories[k] = true
		}
		for _, p := range e.Provenance {
			if p.Source == "" || p.License == "" || p.Modification == "" || p.SourceLine < 0 {
				return nil, fmt.Errorf("invalid provenance on %s %q", name, e.ID)
			}
		}
		for _, s := range append([]string{e.Label}, e.Aliases...) {
			k := key(s)
			if k == "" {
				return nil, fmt.Errorf("empty %s alias", name)
			}
			if old, ok := m[k]; ok && old != e.ID {
				return nil, fmt.Errorf("ambiguous %s label %q", name, s)
			}
			m[k] = e.ID
		}
	}
	return m, nil
}
func LoadTaxonomy(path string) (Taxonomy, error) {
	var t Taxonomy
	if e := readJSON(path, &t); e != nil { // Taxonomy seed includes curator metadata: decode without rejecting its extension fields.
		b, err := os.ReadFile(path)
		if err != nil {
			return t, err
		}
		if err = json.Unmarshal(b, &t); err != nil {
			return t, err
		}
	}
	if t.ID == "" || t.Version == "" {
		return t, errors.New("taxonomy requires id and version")
	}
	if _, e := index(t.Domains, "domain"); e != nil {
		return t, e
	}
	if _, e := index(t.Technologies, "technology"); e != nil {
		return t, e
	}
	seen := map[string]bool{}
	for _, x := range t.Exceptions {
		if key(x.Label) == "" || x.Kind == "" {
			return t, errors.New("invalid classification exception")
		}
		if seen[key(x.Label)] {
			return t, fmt.Errorf("duplicate exception %q", x.Label)
		}
		seen[key(x.Label)] = true
	}
	return t, nil
}
func filesUnder(path string) ([]string, error) {
	info, e := os.Stat(path)
	if e != nil {
		return nil, e
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var out []string
	e = filepath.Walk(path, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !i.IsDir() && strings.EqualFold(filepath.Ext(p), ".json") {
			out = append(out, p)
		}
		return nil
	})
	sort.Strings(out)
	return out, e
}
func Ingest(input, taxonomyPath, bankID string) (Bank, error) {
	var b Bank
	t, e := LoadTaxonomy(taxonomyPath)
	if e != nil {
		return b, e
	}
	if bankID == "" {
		return b, errors.New("bank ID required")
	}
	domains, _ := index(t.Domains, "domain")
	tech, _ := index(t.Technologies, "technology")
	exceptions := map[string]bool{}
	for _, x := range t.Exceptions {
		exceptions[key(x.Label)] = true
	}
	paths, e := filesUnder(input)
	if e != nil {
		return b, e
	}
	if len(paths) == 0 {
		return b, errors.New("no JSON files found")
	}
	b = Bank{SchemaVersion: 1, ID: bankID, TaxonomyID: t.ID, TaxonomyVersion: t.Version, Questions: []Question{}}
	seen := map[string]bool{}
	for _, p := range paths {
		var q Quiz
		if e := readJSON(p, &q); e != nil {
			return b, e
		}
		if q.ID == "" || q.Provider == "" || q.Subject == "" || len(q.Tasks) == 0 {
			return b, fmt.Errorf("%s: missing quiz identity or tasks", p)
		}
		if b.Provider == "" {
			b.Provider = q.Provider
			b.Subject = q.Subject
		} else if b.Provider != q.Provider || b.Subject != q.Subject {
			return b, fmt.Errorf("%s: provider/subject differs from bank", p)
		}
		for i, task := range q.Tasks {
			where := fmt.Sprintf("%s task %d", p, i+1)
			tag := strings.TrimSpace(string(task.Tag))
			if tag == "" || tag == "null" {
				return b, fmt.Errorf("%s: missing tag", where)
			}
			id := q.ID + ":" + tag
			if seen[id] {
				return b, fmt.Errorf("%s: duplicate question ID %s", where, id)
			}
			seen[id] = true
			d, ok := domains[key(task.Domain)]
			if !ok {
				return b, fmt.Errorf("%s: unknown domain %q", where, task.Domain)
			}
			if strings.TrimSpace(task.Title) == "" || strings.TrimSpace(task.Explanation) == "" || len(task.Options) < 2 || len(task.Answer) == 0 {
				return b, fmt.Errorf("%s: incomplete question", where)
			}
			opts := map[string]bool{}
			for _, o := range task.Options {
				if strings.TrimSpace(o.Text) == "" {
					return b, fmt.Errorf("%s: blank option", where)
				}
				opts[strings.TrimSpace(o.Text)] = true
			}
			for _, a := range task.Answer {
				if !opts[strings.TrimSpace(a.Text)] {
					return b, fmt.Errorf("%s: answer does not match an option", where)
				}
			}
			ids := []string{}
			topics := []string{}
			added := map[string]bool{}
			for _, label := range task.Products {
				if tid, ok := tech[key(label)]; ok {
					if !added[tid] {
						ids = append(ids, tid)
						added[tid] = true
					}
				} else if exceptions[key(label)] {
					topics = append(topics, label)
				} else {
					return b, fmt.Errorf("%s: unknown product/technology %q", where, label)
				}
			}
			sort.Strings(ids)
			sort.Strings(topics)
			b.Questions = append(b.Questions, Question{ID: id, Source: Source{File: filepath.Base(p), QuizID: q.ID, Tag: task.Tag}, DomainID: d, TechnologyIDs: ids, TopicLabels: topics, Task: task})
		}
	}
	sort.Slice(b.Questions, func(i, j int) bool { return b.Questions[i].ID < b.Questions[j].ID })
	return b, nil
}
func Write(path string, v any) error {
	data, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
