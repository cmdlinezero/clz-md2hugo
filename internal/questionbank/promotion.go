package questionbank

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// PromotionApprovals is an explicit, reviewed selection. An empty file selects nothing.
// IDs refer to curation suggested_id, not source labels.
type PromotionApprovals struct {
	SchemaVersion int      `json:"schema_version"`
	Add           []string `json:"add"`
	Aliases       []string `json:"aliases"`
	Parents       []string `json:"parents"`
}
type PromotionAlias struct {
	Label       string `json:"label"`
	CanonicalID string `json:"canonical_id"`
}
type PromotionParent struct {
	ChildID  string `json:"child_id"`
	ParentID string `json:"parent_id"`
	Relation string `json:"relation"`
}
type PromotionPlan struct {
	SchemaVersion   int               `json:"schema_version"`
	DryRun          bool              `json:"dry_run"`
	TaxonomyID      string            `json:"taxonomy_id"`
	TaxonomyVersion string            `json:"taxonomy_version"`
	Source          string            `json:"source"`
	SourceCommit    string            `json:"source_commit,omitempty"`
	License         string            `json:"license"`
	Add             []Entry           `json:"add"`
	Aliases         []PromotionAlias  `json:"aliases"`
	Parents         []PromotionParent `json:"parents"`
	Warnings        []string          `json:"warnings"`
	Notice          string            `json:"notice"`
}

// PlanPromotion never writes to or mutates the authoritative taxonomy.
func PlanPromotion(t Taxonomy, c CurationReport, a PromotionApprovals) (PromotionPlan, error) {
	p := PromotionPlan{SchemaVersion: 1, DryRun: true, TaxonomyID: t.ID, TaxonomyVersion: t.Version, Source: c.Source, SourceCommit: c.SourceCommit, License: c.License, Add: []Entry{}, Aliases: []PromotionAlias{}, Parents: []PromotionParent{}, Warnings: []string{}, Notice: "Review-only projection; authoritative taxonomy, quiz data and domain mappings remain unchanged. Source categories are not exam domains."}
	if a.SchemaVersion != 1 || c.SchemaVersion != 1 || c.TaxonomyID != t.ID || c.TaxonomyVersion != t.Version || t.ID != "google-cloud" {
		return p, fmt.Errorf("approval/curation schema or taxonomy mismatch")
	}
	if c.Source != "https://github.com/priyankavergadia/google-cloud-4-words" || c.License != "CC-BY-4.0" {
		return p, fmt.Errorf("unrecognized source or license")
	}
	byID := map[string]CurationItem{}
	for _, i := range c.Items {
		if i.SuggestedID == "" {
			continue
		}
		if _, ok := byID[i.SuggestedID]; ok {
			return p, fmt.Errorf("duplicate curation ID %q", i.SuggestedID)
		}
		byID[i.SuggestedID] = i
	}
	existing := map[string]bool{}
	names := map[string]string{}
	for _, e := range t.Technologies {
		existing[e.ID] = true
		for _, s := range append([]string{e.Label}, e.Aliases...) {
			names[key(s)] = e.ID
		}
	}
	used := map[string]bool{}
	selectItem := func(id, action string) (CurationItem, error) {
		k := action + ":" + id
		if used[k] {
			return CurationItem{}, fmt.Errorf("duplicate %s approval %q", action, id)
		}
		used[k] = true
		item, ok := byID[id]
		if !ok {
			return item, fmt.Errorf("unknown curation ID %q", id)
		}
		expected := map[string]string{"add": "review_add", "alias": "review_alias", "parent": "review_parent"}[action]
		if item.Decision != expected {
			return item, fmt.Errorf("%q has decision %q, not %q", id, item.Decision, expected)
		}
		return item, nil
	}
	for _, id := range a.Add {
		i, err := selectItem(id, "add")
		if err != nil {
			return p, err
		}
		if existing[id] {
			return p, fmt.Errorf("ID collision %q", id)
		}
		if old := names[key(i.Label)]; old != "" {
			return p, fmt.Errorf("label %q already belongs to %s", i.Label, old)
		}
		existing[id] = true
		names[key(i.Label)] = id
		p.Add = append(p.Add, Entry{ID: id, Label: i.Label})
		if i.LifecycleReview {
			p.Warnings = append(p.Warnings, fmt.Sprintf("%s: lifecycle verification still required", id))
		}
	}
	for _, id := range a.Aliases {
		i, err := selectItem(id, "alias")
		if err != nil {
			return p, err
		}
		if !existing[i.AliasOf] {
			return p, fmt.Errorf("alias target %q missing", i.AliasOf)
		}
		if old := names[key(i.Label)]; old != "" && old != i.AliasOf {
			return p, fmt.Errorf("alias collision %q", i.Label)
		}
		names[key(i.Label)] = i.AliasOf
		p.Aliases = append(p.Aliases, PromotionAlias{Label: i.Label, CanonicalID: i.AliasOf})
	}
	for _, id := range a.Parents {
		i, err := selectItem(id, "parent")
		if err != nil {
			return p, err
		}
		if !existing[i.ParentID] {
			return p, fmt.Errorf("parent %q missing", i.ParentID)
		}
		if existing[id] {
			return p, fmt.Errorf("feature ID %q collides with existing technology", id)
		}
		p.Parents = append(p.Parents, PromotionParent{ChildID: id, ParentID: i.ParentID, Relation: "feature_of"})
	}
	sort.Slice(p.Add, func(i, j int) bool { return p.Add[i].ID < p.Add[j].ID })
	sort.Slice(p.Aliases, func(i, j int) bool { return p.Aliases[i].Label < p.Aliases[j].Label })
	sort.Slice(p.Parents, func(i, j int) bool { return p.Parents[i].ChildID < p.Parents[j].ChildID })
	sort.Strings(p.Warnings)
	return p, nil
}
func ReadPromotionApprovals(path string) (PromotionApprovals, error) {
	var a PromotionApprovals
	b, e := os.ReadFile(path)
	if e != nil {
		return a, e
	}
	e = json.Unmarshal(b, &a)
	return a, e
}
func ReadCuration(path string) (CurationReport, error) {
	var c CurationReport
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	e = json.Unmarshal(b, &c)
	return c, e
}
