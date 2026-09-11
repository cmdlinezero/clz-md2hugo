package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildMultiTypeGraph(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "collections", "docker.md"), `---
id: docker-tutorials
type: collection
title: Docker Tutorials
---
# Docker Tutorials
`)
	write(t, filepath.Join(dir, "certifications", "cka.md"), `---
id: cka
type: certification
title: CKA
---
# CKA
`)
	write(t, filepath.Join(dir, "tutorials", "docker", "version.md"), `---
id: docker-version
type: tutorial
title: Check Docker Version
collections: [docker-tutorials]
certifications: [cka]
---
# Check Docker Version
{{< step label="Check" >}}
Run Docker.
{{< /step >}}
`)

	contents, err := compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := graphFrom(contents)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Collections) != 1 || len(graph.Certifications) != 1 || len(graph.Tutorials) != 1 {
		t.Fatalf("unexpected graph counts: %+v", graph.Stats)
	}
	if graph.Tutorials[0].Collections[0] != "docker-tutorials" {
		t.Fatalf("collection relationship not preserved")
	}
}

func TestMissingRelationshipFails(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "tutorial.md"), `---
id: t1
type: tutorial
title: Tutorial
collections: [missing]
---
# Tutorial
`)
	if err := Validate(dir); err == nil {
		t.Fatal("expected missing relationship error")
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestFlashcardAndChallengeModalities(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "flashcards", "docker.md"), "---\nid: docker-flashcards\ntype: flashcard-deck\ntitle: Docker Flashcards\n---\n\n```flashcard {id=\"image\" front=\"What is a Docker image?\" hint=\"Think template.\" explanation=\"A read-only template used to create containers.\"}\nA read-only template used to create containers.\n```\n")
	write(t, filepath.Join(dir, "labs", "docker.md"), "---\nid: docker-lab\ntype: challenge-lab\ntitle: Docker Lab\n---\n\n```challenge {id=\"run\" title=\"Run NGINX\" objective=\"Create a running nginx container.\" command=\"docker run -d --name web nginx:latest\"}\nRun the command and verify it with docker ps.\n```\n")
	contents, err := compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	graph, err := graphFrom(contents)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.FlashcardDecks) != 1 || len(graph.FlashcardDecks[0].Cards) != 1 {
		t.Fatalf("unexpected flashcard graph: %+v", graph.Stats)
	}
	if len(graph.ChallengeLabs) != 1 || len(graph.ChallengeLabs[0].Challenges) != 1 {
		t.Fatalf("unexpected challenge graph: %+v", graph.Stats)
	}
	if graph.ChallengeLabs[0].Challenges[0].Command != "docker run -d --name web nginx:latest" {
		t.Fatal("challenge command not preserved")
	}
}

func TestCheckBlock(t *testing.T) {
	dir := t.TempDir()
	body := "---\nid: check-tutorial\ntype: tutorial\ntitle: Check Tutorial\n---\n\n{{< step label=\"Test\" >}}\nDo the thing.\n\n```check {id=\"verify-thing\" message=\"I did the thing\"}\n```\n{{< /step >}}\n"
	write(t, filepath.Join(dir, "tutorial.md"), body)

	contents, err := compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) != 1 || len(contents[0].Steps) != 1 || len(contents[0].Steps[0].Blocks) != 2 {
		t.Fatalf("unexpected parsed check block: %+v", contents)
	}
	b := contents[0].Steps[0].Blocks[1]
	if b.Type != "check" || b.Attributes["id"] != "verify-thing" || b.Attributes["message"] != "I did the thing" {
		t.Fatalf("check block not preserved: %+v", b)
	}
}

func TestRatingAndProductIDsPreserved(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "tutorial.md"), `---
id: docker-import-export
type: tutorial
title: Docker Import Export
rating_id: docker-import-export
product_id: docker
---
# Docker Import Export
`)

	contents, err := compile(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(contents) != 1 {
		t.Fatalf("expected one content item, got %d", len(contents))
	}
	if contents[0].RatingID != "docker-import-export" {
		t.Fatalf("rating_id not preserved: %+v", contents[0])
	}
	if contents[0].ProductID != "docker" {
		t.Fatalf("product_id not preserved: %+v", contents[0])
	}

	graph, err := graphFrom(contents)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Tutorials) != 1 || graph.Tutorials[0].RatingID != "docker-import-export" || graph.Tutorials[0].ProductID != "docker" {
		t.Fatalf("rating/product IDs not preserved in graph: %+v", graph.Tutorials)
	}
}
