package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGenerateCheckAndCleanup(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "data.json"), `{"projects":[{"slug":"alpha","title":"Alpha"},{"slug":"beta","title":"Beta"}]}`)
	mustWrite(t, filepath.Join(root, "page.md.tmpl"), "---\ntitle: {{ json .title }}\n---\n\n{{ .title }}\n")
	mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"projects","source":"data.json","template":"page.md.tmpl","output":"content","items":"projects","path":"{{ .slug }}.md"}]}`)

	result, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != 2 {
		t.Fatalf("changed=%d, want 2", len(result.Changed))
	}
	alpha, err := os.ReadFile(filepath.Join(root, "content", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := "---\ntitle: \"Alpha\"\n---\n\n" + generatedMarker + "\n"
	if !strings.HasPrefix(string(alpha), wantPrefix) {
		t.Fatalf("front matter or generated marker misplaced: %s", alpha)
	}

	result, err = Run(Options{ConfigPath: filepath.Join(root, "config.json"), Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != 0 {
		t.Fatalf("check changed=%d, want 0", len(result.Changed))
	}

	mustWrite(t, filepath.Join(root, "data.json"), `{"projects":[{"slug":"alpha","title":"Alpha"}]}`)
	result, err = Run(Options{ConfigPath: filepath.Join(root, "config.json"), Check: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != 1 {
		t.Fatalf("stale check changed=%d, want 1", len(result.Changed))
	}
	if _, err := os.Stat(filepath.Join(root, "content", "beta.md")); err != nil {
		t.Fatalf("check removed file: %v", err)
	}

	if _, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "content", "beta.md")); !os.IsNotExist(err) {
		t.Fatalf("stale generated file still exists")
	}
}

func TestRunPreservesHumanOwnedFiles(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "data.json"), `[]`)
	mustWrite(t, filepath.Join(root, "page.md.tmpl"), "{{ . }}\n")
	mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"empty","source":"data.json","template":"page.md.tmpl","output":"content","path":"{{ . }}.md"}]}`)
	mustWrite(t, filepath.Join(root, "content", "_index.md"), "human owned\n")
	if _, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "content", "_index.md")); err != nil {
		t.Fatal("human-owned file was removed")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRunMergesSecondaryDataUnderAlias(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "projects.json"), `{"projects":[{"repo":"alpha","slug":"alpha","title":"Editorial Alpha"}]}`)
	mustWrite(t, filepath.Join(root, "facts.json"), `{"alpha":{"name":"alpha","language":"Go","archived":false}}`)
	mustWrite(t, filepath.Join(root, "page.md.tmpl"), "---\ntitle: {{ json .title }}\nlanguage: {{ json .github.language }}\n---\n")
	mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"projects","source":"projects.json","template":"page.md.tmpl","output":"content","items":"projects","path":"{{ .slug }}.md","merge":[{"source":"facts.json","localKey":"repo","foreignKey":"name","as":"github"}]}]}`)

	result, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Changed) != 1 {
		t.Fatalf("changed=%d, want 1", len(result.Changed))
	}
	page, err := os.ReadFile(filepath.Join(root, "content", "alpha.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(page), `language: "Go"`) {
		t.Fatalf("merged data missing from output: %s", page)
	}
}

func TestRunMergeRejectsMissingAndDuplicateMatches(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		root := t.TempDir()
		mustWrite(t, filepath.Join(root, "projects.json"), `[{"repo":"alpha","slug":"alpha"}]`)
		mustWrite(t, filepath.Join(root, "facts.json"), `[{"name":"beta"}]`)
		mustWrite(t, filepath.Join(root, "page.md.tmpl"), "{{ .slug }}\n")
		mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"projects","source":"projects.json","template":"page.md.tmpl","output":"content","path":"{{ .slug }}.md","merge":[{"source":"facts.json","localKey":"repo","foreignKey":"name","as":"facts"}]}]}`)

		_, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")})
		if err == nil || !strings.Contains(err.Error(), "found no match") {
			t.Fatalf("err=%v, want missing match error", err)
		}
	})

	t.Run("duplicate", func(t *testing.T) {
		root := t.TempDir()
		mustWrite(t, filepath.Join(root, "projects.json"), `[{"repo":"alpha","slug":"alpha"}]`)
		mustWrite(t, filepath.Join(root, "facts.json"), `[{"name":"alpha","value":1},{"name":"alpha","value":2}]`)
		mustWrite(t, filepath.Join(root, "page.md.tmpl"), "{{ .slug }}\n")
		mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"projects","source":"projects.json","template":"page.md.tmpl","output":"content","path":"{{ .slug }}.md","merge":[{"source":"facts.json","localKey":"repo","foreignKey":"name","as":"facts"}]}]}`)

		_, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")})
		if err == nil || !strings.Contains(err.Error(), "duplicate foreign key") {
			t.Fatalf("err=%v, want duplicate key error", err)
		}
	})
}

func TestRunMergeRejectsAliasCollision(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "projects.json"), `[{"repo":"alpha","slug":"alpha","github":{"manual":true}}]`)
	mustWrite(t, filepath.Join(root, "facts.json"), `[{"name":"alpha"}]`)
	mustWrite(t, filepath.Join(root, "page.md.tmpl"), "{{ .slug }}\n")
	mustWrite(t, filepath.Join(root, "config.json"), `{"version":1,"jobs":[{"name":"projects","source":"projects.json","template":"page.md.tmpl","output":"content","path":"{{ .slug }}.md","merge":[{"source":"facts.json","localKey":"repo","foreignKey":"name","as":"github"}]}]}`)

	_, err := Run(Options{ConfigPath: filepath.Join(root, "config.json")})
	if err == nil || !strings.Contains(err.Error(), "collides") {
		t.Fatalf("err=%v, want alias collision error", err)
	}
}

func TestGeneratedMarkerPlacementAndDetection(t *testing.T) {
	tests := []struct {
		name     string
		rendered string
		prefix   string
	}{
		{
			name:     "yaml front matter",
			rendered: "---\ntitle: Example\n---\n\nBody\n",
			prefix:   "---\ntitle: Example\n---\n\n" + generatedMarker + "\n",
		},
		{
			name:     "toml front matter",
			rendered: "+++\ntitle = \"Example\"\n+++\n\nBody\n",
			prefix:   "+++\ntitle = \"Example\"\n+++\n\n" + generatedMarker + "\n",
		},
		{
			name:     "no front matter",
			rendered: "Body\n",
			prefix:   generatedMarker + "\nBody\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := addGeneratedMarker(tt.rendered)
			if !strings.HasPrefix(string(content), tt.prefix) {
				t.Fatalf("content=%q, want prefix %q", content, tt.prefix)
			}
			if !hasGeneratedMarker(content) {
				t.Fatal("generated content was not recognised as owned")
			}
		})
	}
}

func TestGeneratedMarkerDetectionSupportsLegacyFiles(t *testing.T) {
	legacy := []byte(generatedMarker + "\n---\ntitle: Legacy\n---\n")
	if !hasGeneratedMarker(legacy) {
		t.Fatal("legacy prefix marker was not recognised")
	}
}

func TestGeneratedMarkerDetectionRejectsHumanOwnedFrontMatter(t *testing.T) {
	human := []byte("---\ntitle: Human\n---\n\nBody\n")
	if hasGeneratedMarker(human) {
		t.Fatal("human-owned file was recognised as generated")
	}
}
