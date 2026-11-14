package questionbank

import "testing"

func TestCurateReconciliation(t *testing.T) {
	tax := Taxonomy{ID: "google-cloud", Version: "0.1", Technologies: []Entry{{ID: "iam", Label: "IAM"}, {ID: "bigquery", Label: "BigQuery"}}}
	r := Reconciliation{Source: "https://github.com/priyankavergadia/google-cloud-4-words", License: "CC-BY-4.0", Proposed: []CheatSheetCandidate{
		{Label: "/r/bigquery", Categories: []string{"Additional Resources"}},
		{Label: "BigQuery ML", Categories: []string{"Data Analytics"}},
		{Label: "Cloud IAM", Categories: []string{"Identity and Security"}},
		{Label: "Cloud Run", Categories: []string{"Compute"}},
		{Label: "Cloud IoT Core", Categories: []string{"Internet of Things (IoT)"}},
	}}
	out, err := CurateReconciliation(tax, r)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]string{"/r/bigquery": "exclude", "BigQuery ML": "review_parent", "Cloud IAM": "review_alias", "Cloud Run": "review_add", "Cloud IoT Core": "review_add"}
	for _, item := range out.Items {
		if item.Decision != expected[item.Label] {
			t.Errorf("%s: %s", item.Label, item.Decision)
		}
		if item.Label == "Cloud IoT Core" && !item.LifecycleReview {
			t.Error("missing lifecycle review")
		}
	}
	if len(out.Items) != 5 {
		t.Fatalf("got %d items", len(out.Items))
	}
}
func TestCurateRejectsUnexpectedSource(t *testing.T) {
	_, err := CurateReconciliation(Taxonomy{ID: "google-cloud"}, Reconciliation{Source: "untrusted", License: "CC-BY-4.0"})
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestFirestoreSharedCategories(t *testing.T) {
	tax := Taxonomy{ID: "google-cloud", Version: "0.1"}
	r := Reconciliation{Source: "https://github.com/priyankavergadia/google-cloud-4-words", License: "CC-BY-4.0", Proposed: []CheatSheetCandidate{{Label: "Cloud Firestore", Categories: []string{"Database", "Mobile (Firebase)"}}, {Label: "Firebase Authentication", Categories: []string{"Mobile (Firebase)"}}}}
	out, err := CurateReconciliation(tax, r)
	if err != nil {
		t.Fatal(err)
	}
	for _, i := range out.Items {
		if i.Label == "Cloud Firestore" && i.Decision != "review_add" {
			t.Fatalf("shared Firestore classified as %s", i.Decision)
		}
		if i.Label == "Firebase Authentication" && i.Decision != "review_scope" {
			t.Fatalf("Firebase-only classified as %s", i.Decision)
		}
	}
}
