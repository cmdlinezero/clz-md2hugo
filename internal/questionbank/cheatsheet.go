package questionbank

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// CheatSheetCandidate is a source observation, not an automatically approved taxonomy entry.
type CheatSheetCandidate struct {
	Label       string   `json:"label"`
	Categories  []string `json:"categories"`
	Description string   `json:"description,omitempty"`
	SourceLine  int      `json:"source_line"`
}
type Reconciliation struct {
	Source       string                `json:"source"`
	SourceCommit string                `json:"source_commit,omitempty"`
	License      string                `json:"license"`
	Existing     int                   `json:"existing"`
	Candidates   int                   `json:"candidates"`
	Matched      []ReconcileMatch      `json:"matched"`
	Proposed     []CheatSheetCandidate `json:"proposed"`
	Review       []ReconcileMatch      `json:"review"`
}
type ReconcileMatch struct {
	SourceLabel    string   `json:"source_label"`
	CanonicalID    string   `json:"canonical_id,omitempty"`
	CanonicalLabel string   `json:"canonical_label,omitempty"`
	Categories     []string `json:"categories"`
	Reason         string   `json:"reason"`
	SourceLine     int      `json:"source_line"`
}

var sheetEntry = regexp.MustCompile(`^\s*[-*]\s+(.+?):\s*(.*?)\s*$`)
var sheetHeading = regexp.MustCompile(`^###\s+(.+?)\s*$`)
var sheetLink = regexp.MustCompile(`\s*\[[^]]+\]\([^)]*\)`)
var sheetFootnote = regexp.MustCompile(`\s*\[\^[^]]+\]`)

// ParseCheatSheet reads the README's "All Products" section. It intentionally
// ignores surrounding prose, markdown links and non-product sections.
func ParseCheatSheet(path string) ([]CheatSheetCandidate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	inProducts, category, line := false, "", 0
	byLabel := map[string]int{}
	result := []CheatSheetCandidate{}
	for scanner.Scan() {
		line++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "# All Products" {
			inProducts = true
			continue
		}
		if !inProducts {
			continue
		}
		if strings.HasPrefix(raw, "# ") {
			break
		}
		if h := sheetHeading.FindStringSubmatch(raw); h != nil {
			category = strings.TrimSpace(h[1])
			continue
		}
		if category == "" {
			continue
		}
		m := sheetEntry.FindStringSubmatch(raw)
		if m == nil {
			continue
		}
		label := strings.TrimSpace(strings.ReplaceAll(m[1], "**", ""))
		if label == "" {
			continue
		}
		description := strings.TrimSpace(sheetFootnote.ReplaceAllString(sheetLink.ReplaceAllString(m[2], ""), ""))
		k := key(label)
		if i, ok := byLabel[k]; ok {
			found := false
			for _, c := range result[i].Categories {
				if c == category {
					found = true
				}
			}
			if !found {
				result[i].Categories = append(result[i].Categories, category)
			}
		} else {
			byLabel[k] = len(result)
			result = append(result, CheatSheetCandidate{Label: label, Categories: []string{category}, Description: description, SourceLine: line})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if !inProducts || len(result) == 0 {
		return nil, fmt.Errorf("%s: no products found under # All Products", path)
	}
	sort.Slice(result, func(i, j int) bool { return key(result[i].Label) < key(result[j].Label) })
	for i := range result {
		sort.Strings(result[i].Categories)
	}
	return result, nil
}
func slug(s string) string {
	var b strings.Builder
	hyphen := false
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if hyphen && b.Len() > 0 {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			hyphen = false
		} else {
			hyphen = true
		}
	}
	return b.String()
}

// ReconcileCheatSheet does not mutate a curated taxonomy. Unmatched labels are
// proposals, and probable aliases require review rather than silent acceptance.
func ReconcileCheatSheet(t Taxonomy, candidates []CheatSheetCandidate, sourceCommit string) Reconciliation {
	exact, _ := index(t.Technologies, "technology")
	r := Reconciliation{Source: "https://github.com/priyankavergadia/google-cloud-4-words", SourceCommit: sourceCommit, License: "CC-BY-4.0", Existing: len(t.Technologies), Candidates: len(candidates), Matched: []ReconcileMatch{}, Proposed: []CheatSheetCandidate{}, Review: []ReconcileMatch{}}
	bySlug := map[string]Entry{}
	for _, e := range t.Technologies {
		bySlug[slug(e.Label)] = e
		for _, a := range e.Aliases {
			bySlug[slug(a)] = e
		}
	}
	for _, c := range candidates {
		m := ReconcileMatch{SourceLabel: c.Label, Categories: c.Categories, SourceLine: c.SourceLine}
		if id, ok := exact[key(c.Label)]; ok {
			m.CanonicalID = id
			m.Reason = "existing label or alias"
			for _, e := range t.Technologies {
				if e.ID == id {
					m.CanonicalLabel = e.Label
					break
				}
			}
			r.Matched = append(r.Matched, m)
			continue
		}
		if e, ok := bySlug[slug(c.Label)]; ok {
			m.CanonicalID = e.ID
			m.CanonicalLabel = e.Label
			m.Reason = "punctuation or spacing variant; review alias"
			r.Review = append(r.Review, m)
			continue
		}
		r.Proposed = append(r.Proposed, c)
	}
	return r
}
