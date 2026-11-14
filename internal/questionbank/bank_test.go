package questionbank

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIngest(t *testing.T) {
	d := t.TempDir()
	tax := filepath.Join(d, "tax.json")
	quiz := filepath.Join(d, "q.json")
	os.WriteFile(tax, []byte(`{"id":"test","version":"1","domains":[{"id":"analytics","label":"Analytics"}],"technologies":[{"id":"pubsub","label":"Pub/Sub","aliases":["Cloud Pub/Sub"]}]}`), 0600)
	s := `{"id":"q1","provider":"Google Cloud","subject":"PDE","tasks":[{"tag":1,"title":"Question","domain":"Analytics","explanation":"Because","answer":[{"answerText":"A"}],"options":[{"option":"A"},{"option":"B"}],"products":["Cloud Pub/Sub"]}]}`
	os.WriteFile(quiz, []byte(s), 0600)
	b, e := Ingest(quiz, tax, "pde")
	if e != nil {
		t.Fatal(e)
	}
	if len(b.Questions) != 1 || b.Questions[0].TechnologyIDs[0] != "pubsub" {
		t.Fatal(b)
	}
	os.WriteFile(quiz, []byte(strings.Replace(s, "Cloud Pub/Sub", "Unknown", 1)), 0600)
	if _, e = Ingest(quiz, tax, "pde"); e == nil {
		t.Fatal("unknown technology accepted")
	}
}
func TestFixture(t *testing.T) {
	tax := "../../taxonomies/google-cloud/taxonomy.json"
	source := "../../../pde-quiz-1.json"
	if _, e := os.Stat(source); e != nil {
		t.Skip("external sample not present")
	}
	_, e := Ingest(source, tax, "pde")
	if e != nil {
		t.Fatal(e)
	}
}
