package questionbank

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// CurationItem is a review recommendation, never an automatic taxonomy change.
type CurationItem struct {
	Label           string   `json:"label"`
	SuggestedID     string   `json:"suggested_id,omitempty"`
	Kind            string   `json:"kind"`
	Decision        string   `json:"decision"`
	Reasons         []string `json:"reasons"`
	Categories      []string `json:"categories"`
	Description     string   `json:"description,omitempty"`
	SourceLine      int      `json:"source_line"`
	ParentID        string   `json:"parent_id,omitempty"`
	AliasOf         string   `json:"alias_of,omitempty"`
	LifecycleReview bool     `json:"lifecycle_review,omitempty"`
}
type CurationReport struct {
	SchemaVersion   int            `json:"schema_version"`
	Source          string         `json:"source"`
	SourceCommit    string         `json:"source_commit,omitempty"`
	License         string         `json:"license"`
	TaxonomyID      string         `json:"taxonomy_id"`
	TaxonomyVersion string         `json:"taxonomy_version"`
	MatchedCount    int            `json:"matched_count"`
	Items           []CurationItem `json:"items"`
	Counts          map[string]int `json:"counts"`
	Notice          string         `json:"notice"`
}

var excludedCategories = map[string]bool{
	"Additional Resources": true, "Language Specific Documentation": true, "Platform Comparisons": true,
}
var ecosystemCategories = map[string]bool{
	"Google Cloud Foundational Open Source Projects": true, "Google Maps Platform": true,
	"Workspace Platform": true, "Mobile (Firebase)": true,
}
var lifecycleLabels = map[string]bool{
	"Anthos": true, "Anthos clusters": true, "Anthos Config Management": true,
	"Anthos Service Mesh": true, "Cloud Functions": true, "Cloud IoT Core": true,
	"Container Registry": true, "Google Data Studio": true, "Cloud Source Repositories": true,
	"Cloud Deployment Manager": true, "Dataprep by Trifacta": true,
	"Firebase Dynamic Links": true, "Firebase Predictions": true,
}
var aliasHints = map[string]string{
	"Cloud IAM": "iam", "Google Kubernetes Engine (GKE)": "gke",
	"Cloud Memorystore": "memorystore", "Cloud Filestore": "filestore",
	"Cloud Key Management Service": "cloud-kms",
}
var featureParents = map[string]string{
	"BigQuery BI Engine": "bigquery", "BigQuery GIS": "bigquery", "BigQuery ML": "bigquery",
	"BigQuery Data Transfer Service": "bigquery", "Cloud SQL Insights": "cloud-sql",
}

// CurateReconciliation produces a conservative, deterministic review queue.
// It does not modify the taxonomy, approve candidates or assert product lifecycle status.
func CurateReconciliation(t Taxonomy, r Reconciliation) (CurationReport, error) {
	if t.ID != "google-cloud" {
		return CurationReport{}, fmt.Errorf("curation requires google-cloud taxonomy")
	}
	if r.Source != "https://github.com/priyankavergadia/google-cloud-4-words" || r.License != "CC-BY-4.0" {
		return CurationReport{}, fmt.Errorf("unexpected reconciliation source or license")
	}
	result := CurationReport{SchemaVersion: 1, Source: r.Source, SourceCommit: r.SourceCommit, License: r.License, TaxonomyID: t.ID, TaxonomyVersion: t.Version, MatchedCount: len(r.Matched), Items: []CurationItem{}, Counts: map[string]int{}, Notice: "Suggestions only. No taxonomy entries, aliases, domain mappings, lifecycle statuses, or quiz data are changed. Verify all lifecycle and alias suggestions against current official documentation; source categories are not exam domains."}
	existing := map[string]bool{}
	for _, e := range t.Technologies {
		existing[e.ID] = true
	}
	seen := map[string]bool{}
	for _, c := range r.Proposed {
		k := key(c.Label)
		if seen[k] {
			return CurationReport{}, fmt.Errorf("duplicate proposed label %q", c.Label)
		}
		seen[k] = true
		item := CurationItem{Label: c.Label, SuggestedID: slug(c.Label), Categories: c.Categories, Description: c.Description, SourceLine: c.SourceLine, Kind: "technology", Decision: "review_add", Reasons: []string{"unmatched source entry; curator approval required"}}
		allExcluded := len(c.Categories) > 0
		ecosystem := false
		core := false
		for _, cat := range c.Categories {
			if !excludedCategories[cat] {
				allExcluded = false
			}
			if ecosystemCategories[cat] {
				ecosystem = true
			} else if !excludedCategories[cat] {
				core = true
			}
		}
		switch {
		case allExcluded:
			item.Kind = "reference"
			item.Decision = "exclude"
			item.Reasons = []string{"documentation, community or reference category; not a cloud product"}
		case featureParents[c.Label] != "":
			item.Kind = "feature"
			item.ParentID = featureParents[c.Label]
			item.Decision = "review_parent"
			item.Reasons = []string{"possible subservice or feature; review parent relationship"}
		case ecosystem && !core:
			item.Kind = "ecosystem"
			item.Decision = "review_scope"
			item.Reasons = []string{"outside core Google Cloud product scope; review provider/technology ownership"}
		}
		if hint := aliasHints[c.Label]; hint != "" && existing[hint] {
			item.AliasOf = hint
			item.Decision = "review_alias"
			item.Reasons = append(item.Reasons, "possible alias of existing stable ID; confirm before accepting")
		}
		if lifecycleLabels[c.Label] {
			item.LifecycleReview = true
			item.Reasons = append(item.Reasons, "historical or renamed product candidate; verify current lifecycle")
		}
		result.Items = append(result.Items, item)
		result.Counts[item.Decision]++
	}
	for _, m := range r.Review {
		k := key(m.SourceLabel)
		if seen[k] {
			return CurationReport{}, fmt.Errorf("duplicate review label %q", m.SourceLabel)
		}
		seen[k] = true
		result.Items = append(result.Items, CurationItem{Label: m.SourceLabel, Kind: "alias", Decision: "review_alias", AliasOf: m.CanonicalID, Categories: m.Categories, SourceLine: m.SourceLine, Reasons: []string{m.Reason}})
		result.Counts["review_alias"]++
	}
	sort.Slice(result.Items, func(i, j int) bool {
		return strings.ToLower(result.Items[i].Label) < strings.ToLower(result.Items[j].Label)
	})
	return result, nil
}

// ReadReconciliation loads the report emitted by taxonomy import.
func ReadReconciliation(path string, out *Reconciliation) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
