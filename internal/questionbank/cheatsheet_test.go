package questionbank

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheatSheetParseAndReconcile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "README.md")
	fixture := `# Introduction
* Ignore: this
# All Products
### Data Analytics
* Pub/Sub: Global real-time messaging [docs](https://example.com)
* BigQuery: Data warehouse and analytics
### Application Integration
* Pub/Sub: Global real-time messaging
### Database
* Cloud Bigtable: Petabyte-scale database
* Future Product: New thing
# Other
* Another: Ignore
`
	if err := os.WriteFile(source, []byte(fixture), 0600); err != nil {
		t.Fatal(err)
	}
	entries, err := ParseCheatSheet(source)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 4 {
		t.Fatalf("got %d candidates", len(entries))
	}
	var pub CheatSheetCandidate
	for _, e := range entries {
		if e.Label == "Pub/Sub" {
			pub = e
		}
	}
	if len(pub.Categories) != 2 {
		t.Fatalf("categories: %#v", pub.Categories)
	}
	tax := Taxonomy{Technologies: []Entry{{ID: "pubsub", Label: "Pub/Sub"}, {ID: "bigtable", Label: "Bigtable", Aliases: []string{"Cloud Bigtable"}}, {ID: "bigquery", Label: "BigQuery"}}}
	report := ReconcileCheatSheet(tax, entries, "abc123")
	if len(report.Matched) != 3 || len(report.Proposed) != 1 || report.SourceCommit != "abc123" {
		t.Fatalf("report: %#v", report)
	}
	if report.Proposed[0].Label != "Future Product" {
		t.Fatal(report.Proposed)
	}
}
func TestCheatSheetMissingProducts(t *testing.T) {
	p := filepath.Join(t.TempDir(), "README.md")
	_ = os.WriteFile(p, []byte("# No products"), 0600)
	if _, err := ParseCheatSheet(p); err == nil {
		t.Fatal("expected error")
	}
}
