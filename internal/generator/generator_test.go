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
	if !strings.HasPrefix(string(alpha), generatedMarker+"\n") {
		t.Fatalf("missing generated marker: %s", alpha)
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
