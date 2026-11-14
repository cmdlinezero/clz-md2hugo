package questionbank

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPromotionPerEntryEvidence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "taxonomy.json")
	seed := `{"id":"google-cloud","version":"0.1","attribution":{},"domains":[],"technologies":[{"id":"iam","label":"IAM"},{"id":"bigquery","label":"BigQuery"}]}`
	if err := os.WriteFile(path, []byte(seed), 0600); err != nil {
		t.Fatal(err)
	}
	c := CurationReport{SchemaVersion: 1, TaxonomyID: "google-cloud", TaxonomyVersion: "0.1", Source: "https://github.com/priyankavergadia/google-cloud-4-words", SourceCommit: "abc", License: "CC-BY-4.0", Items: []CurationItem{
		{Label: "Cloud Run", SuggestedID: "cloud-run", Decision: "review_add", Categories: []string{"Compute"}, SourceLine: 51},
		{Label: "Cloud IAM", SuggestedID: "cloud-iam", Decision: "review_alias", AliasOf: "iam", Categories: []string{"Identity and Security"}, SourceLine: 187},
		{Label: "BigQuery ML", SuggestedID: "bigquery-ml", Decision: "review_parent", ParentID: "bigquery", Categories: []string{"Data Analytics"}, SourceLine: 81},
	}}
	a := PromotionApprovals{SchemaVersion: 1, Add: []string{"cloud-run"}, Aliases: []string{"cloud-iam"}, Parents: []string{"bigquery-ml"}}
	if _, err := ApplyPromotion(path, filepath.Join(dir, "backup.json"), c, a); err != nil {
		t.Fatal(err)
	}
	var root struct {
		Technologies []Entry `json:"technologies"`
	}
	b, _ := os.ReadFile(path)
	if err := json.Unmarshal(b, &root); err != nil {
		t.Fatal(err)
	}
	for _, entry := range root.Technologies {
		switch entry.ID {
		case "cloud-run", "bigquery-ml", "iam":
			if len(entry.Provenance) != 1 || entry.Provenance[0].SourceCommit != "abc" || entry.Provenance[0].SourceLine == 0 {
				t.Fatalf("missing evidence for %s: %+v", entry.ID, entry.Provenance)
			}
		}
	}
}
func TestTaxonomySharedCategories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "taxonomy.json")
	seed := `{"id":"google-cloud","version":"0.1","attribution":{},"domains":[],"technologies":[{"id":"firestore","label":"Firestore","aliases":["Cloud Firestore"],"categories":["Database","Mobile (Firebase)"]}]}`
	if err := os.WriteFile(path, []byte(seed), 0600); err != nil {
		t.Fatal(err)
	}
	tax, err := LoadTaxonomy(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(tax.Technologies) != 1 || len(tax.Technologies[0].Categories) != 2 {
		t.Fatal("shared classification not preserved")
	}
}
