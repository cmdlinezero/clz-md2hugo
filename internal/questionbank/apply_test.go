package questionbank

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestApplyPromotionRoundTripAndRepeatFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "taxonomy.json")
	backup := filepath.Join(dir, "taxonomy.backup.json")
	seed := `{"schema_version":"1.0-draft","id":"google-cloud","version":"0.1","status":"draft","custom_metadata":{"keep":true},"attribution":{"source":"Google Cloud in 4 Words","creator":"Priyanka Vergadia","url":"https://github.com/priyankavergadia/google-cloud-4-words","license":"CC-BY-4.0","license_url":"https://creativecommons.org/licenses/by/4.0/","adaptation_notice":"seed"},"domains":[],"technologies":[{"id":"iam","label":"IAM","aliases":[],"review_status":"original"},{"id":"bigquery","label":"BigQuery","aliases":[]}]}`
	if err := os.WriteFile(path, []byte(seed), 0600); err != nil {
		t.Fatal(err)
	}
	c := CurationReport{SchemaVersion: 1, TaxonomyID: "google-cloud", TaxonomyVersion: "0.1", Source: "https://github.com/priyankavergadia/google-cloud-4-words", SourceCommit: "abc123", License: "CC-BY-4.0", Items: []CurationItem{{Label: "Cloud Run", SuggestedID: "cloud-run", Decision: "review_add"}, {Label: "Cloud IAM", SuggestedID: "cloud-iam", Decision: "review_alias", AliasOf: "iam"}, {Label: "BigQuery ML", SuggestedID: "bigquery-ml", Decision: "review_parent", ParentID: "bigquery"}}}
	a := PromotionApprovals{SchemaVersion: 1, Add: []string{"cloud-run"}, Aliases: []string{"cloud-iam"}, Parents: []string{"bigquery-ml"}}
	plan, err := ApplyPromotion(path, backup, c, a)
	if err != nil {
		t.Fatal(err)
	}
	if plan.DryRun || len(plan.Add) != 1 || len(plan.Parents) != 1 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	old, err := os.ReadFile(backup)
	if err != nil || string(old) != seed {
		t.Fatalf("backup mismatch %v", err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err = json.Unmarshal(updated, &root); err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		Keep bool `json:"keep"`
	}
	if err := json.Unmarshal(root["custom_metadata"], &metadata); err != nil || !metadata.Keep {
		t.Fatalf("metadata lost or invalid: %s (%v)", root["custom_metadata"], err)
	}
	var rel []PromotionParent
	if err = json.Unmarshal(root["technology_relationships"], &rel); err != nil || len(rel) != 1 || rel[0].Relation != "feature_of" {
		t.Fatalf("bad relations: %+v %v", rel, err)
	}
	tax, err := LoadTaxonomy(path)
	if err != nil || len(tax.Technologies) != 4 {
		t.Fatalf("invalid taxonomy %v", err)
	}
	if _, err := ApplyPromotion(path, backup, c, a); err == nil {
		t.Fatal("reapply should fail")
	}
	after, _ := os.ReadFile(path)
	if string(after) != string(updated) {
		t.Fatal("reapply changed taxonomy")
	}
}
func TestApplyPromotionEmptyIsRejected(t *testing.T) {
	_, err := ApplyPromotion("missing.json", "backup.json", CurationReport{}, PromotionApprovals{SchemaVersion: 1})
	if err == nil {
		t.Fatal("empty apply accepted")
	}
}
