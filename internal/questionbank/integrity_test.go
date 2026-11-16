package questionbank

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTaxonomyIntegrity(t *testing.T) {
	base := `{"id":"google-cloud","version":"1","domains":[{"id":"d","label":"Domain"}],"technologies":[{"id":"firestore","label":"Firestore","aliases":["Cloud Firestore"],"categories":["Database","Mobile (Firebase)"],"provenance":[{"source":"https://example.org","license":"CC-BY-4.0","modification":"reviewed"}]},{"id":"bigquery","label":"BigQuery"},{"id":"bigquery-ml","label":"BigQuery ML"}],"technology_relationships":[{"child_id":"bigquery-ml","parent_id":"bigquery","relation":"feature_of"}]}`
	cases := []struct{ name, data, want string }{
		{"valid", base, ""},
		{"missing endpoint", strings.Replace(base, `"parent_id":"bigquery"`, `"parent_id":"missing"`, 1), "missing endpoint"},
		{"self edge", strings.Replace(base, `"parent_id":"bigquery"`, `"parent_id":"bigquery-ml"`, 1), "self-referential"},
		{"duplicate edge", strings.Replace(base, `"technology_relationships":[{`, `"technology_relationships":[{"child_id":"bigquery-ml","parent_id":"bigquery","relation":"feature_of"},{`, 1), "duplicate technology relationship"},
		{"cycle", strings.Replace(base, `"technology_relationships":[{`, `"technology_relationships":[{"child_id":"bigquery","parent_id":"bigquery-ml","relation":"feature_of"},{`, 1), "cycle"},
		{"unknown relation", strings.Replace(base, `"relation":"feature_of"`, `"relation":"prerequisite"`, 1), "unsupported"},
		{"duplicate alias", strings.Replace(base, `"Cloud Firestore"`, `"Cloud Firestore","cloud firestore"`, 1), "duplicate technology label"},
		{"duplicate category", strings.Replace(base, `"Mobile (Firebase)"`, `"database"`, 1), "duplicate category"},
		{"invalid provenance", strings.Replace(base, `"license":"CC-BY-4.0"`, `"license":""`, 1), "invalid provenance"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "taxonomy.json")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := LoadTaxonomy(path)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				if len(got.TechnologyRelationships) != 1 || len(got.Technologies[0].Categories) != 2 {
					t.Fatal("graph or categories lost")
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
}
