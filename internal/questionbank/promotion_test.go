package questionbank

import "testing"

func TestPromotionApprovalGate(t *testing.T) {
	tax := Taxonomy{ID: "google-cloud", Version: "0.1", Technologies: []Entry{{ID: "iam", Label: "IAM"}, {ID: "bigquery", Label: "BigQuery"}}}
	c := CurationReport{SchemaVersion: 1, TaxonomyID: "google-cloud", TaxonomyVersion: "0.1", Source: "https://github.com/priyankavergadia/google-cloud-4-words", License: "CC-BY-4.0", Items: []CurationItem{{Label: "Cloud IAM", SuggestedID: "cloud-iam", Decision: "review_alias", AliasOf: "iam"}, {Label: "Cloud Run", SuggestedID: "cloud-run", Decision: "review_add"}, {Label: "BigQuery ML", SuggestedID: "bigquery-ml", Decision: "review_parent", ParentID: "bigquery"}, {Label: "Blog", SuggestedID: "blog", Decision: "exclude"}}}
	empty, err := PlanPromotion(tax, c, PromotionApprovals{SchemaVersion: 1})
	if err != nil || len(empty.Add) != 0 || len(empty.Aliases) != 0 {
		t.Fatalf("empty approval must select nothing: %v", err)
	}
	a := PromotionApprovals{SchemaVersion: 1, Add: []string{"cloud-run"}, Aliases: []string{"cloud-iam"}, Parents: []string{"bigquery-ml"}}
	p, err := PlanPromotion(tax, c, a)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Add) != 1 || len(p.Aliases) != 1 || len(p.Parents) != 1 || len(tax.Technologies) != 2 {
		t.Fatalf("unexpected plan %+v", p)
	}
	a.Add = []string{"blog"}
	if _, err := PlanPromotion(tax, c, a); err == nil {
		t.Fatal("excluded reference approved")
	}
	a.Add = []string{"cloud-run", "cloud-run"}
	if _, err := PlanPromotion(tax, c, a); err == nil {
		t.Fatal("duplicate approval")
	}
}
