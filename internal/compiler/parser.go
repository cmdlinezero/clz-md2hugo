package compiler

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var stepStart = regexp.MustCompile(`^\{\{<\s*step\s+(.+?)\s*>\}\}$`)
var stepEnd = regexp.MustCompile(`^\{\{<\s*/step\s*>\}\}$`)
var attrRE = regexp.MustCompile(`([A-Za-z0-9_-]+)=(?:"([^"]*)"|'([^']*)'|([^\s]+))`)

func ParseFile(path string) (Content, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Content{}, err
	}
	return Parse(string(b), path)
}

func Parse(src, path string) (Content, error) {
	front, body, err := splitFrontmatter(src)
	if err != nil {
		return Content{}, fmt.Errorf("%s: %w", path, err)
	}
	meta, err := parseFrontmatter(front)
	if err != nil {
		return Content{}, fmt.Errorf("%s: %w", path, err)
	}

	contentType := stringValue(meta["type"])
	if contentType == "" {
		contentType = "tutorial"
	}
	c := Content{
		ID: stringValue(meta["id"]), Type: contentType, Title: stringValue(meta["title"]),
		Description: stringValue(meta["description"]), Date: stringValue(meta["date"]),
		Categories: stringSlice(meta["categories"]), Tags: stringSlice(meta["tags"]),
		Topics: stringSlice(meta["topics"]), Collections: stringSlice(meta["collections"]),
		Certifications: stringSlice(meta["certifications"]), Prerequisites: stringSlice(meta["prerequisites"]),
		Children: stringSlice(meta["children"]), Activities: stringSlice(meta["activities"]),
		Duration: stringValue(meta["duration"]), HeroTitle: stringValue(meta["hero_title"]),
		HeroImage: stringValue(meta["hero_image"]), Issue: intValue(meta["issue"]),
		Kind: stringValue(meta["kind"]), Provider: stringValue(meta["provider"]),
		RatingID: stringValue(meta["rating_id"]), ProductID: stringValue(meta["product_id"]),
		Options: stringSlice(meta["options"]), Answer: intValue(meta["answer"]), Explanation: stringValue(meta["explanation"]),
		Volume: intValue(meta["volume"]), Special: boolValue(meta["special_edition"]),
	}
	c.Draft = boolValue(meta["draft"])
	c.Status = stringValue(meta["status"])
	if c.Status == "" {
		if c.Draft {
			c.Status = "draft"
		} else {
			c.Status = "published"
		}
	}
	if c.Status == "draft" {
		c.Draft = true
	}
	if c.ID == "" {
		c.ID = slugify(c.Title)
	}
	c.Slug = c.ID
	if c.Title == "" {
		return Content{}, errors.New("missing required frontmatter: title")
	}

	steps, conclusion, err := parseSteps(body)
	if err != nil {
		return Content{}, fmt.Errorf("%s: %w", path, err)
	}
	c.Steps, c.Conclusion = steps, conclusion
	if c.Type == "flashcard-deck" || c.Type == "challenge-lab" {
		blocks, err := parseBlocks(body)
		if err != nil {
			return Content{}, fmt.Errorf("%s: %w", path, err)
		}
		if c.Type == "flashcard-deck" {
			for _, b := range blocks {
				if b.Type != "flashcard" {
					continue
				}
				c.Cards = append(c.Cards, Flashcard{ID: b.Attributes["id"], Front: b.Attributes["front"], Back: b.Content, Hint: b.Attributes["hint"], Explanation: b.Attributes["explanation"], Topics: parseInlineList(b.Attributes["topics"])})
			}
		} else {
			for _, b := range blocks {
				if b.Type != "challenge" {
					continue
				}
				c.Challenges = append(c.Challenges, Challenge{ID: b.Attributes["id"], Title: b.Attributes["title"], Objective: b.Attributes["objective"], Instructions: b.Content, Hint: b.Attributes["hint"], Success: b.Attributes["success"], Command: b.Attributes["command"], Attributes: b.Attributes})
			}
		}
	}
	return c, nil
}

func splitFrontmatter(src string) (string, string, error) {
	s := strings.ReplaceAll(src, "\r\n", "\n")
	if !strings.HasPrefix(s, "---\n") {
		return "", s, errors.New("frontmatter must start with ---")
	}
	rest := s[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", errors.New("unterminated frontmatter")
	}
	front := rest[:idx]
	body := rest[idx+4:]
	body = strings.TrimPrefix(body, "\n")
	return front, body, nil
}

func parseFrontmatter(src string) (map[string]any, error) {
	out := map[string]any{}
	sc := bufio.NewScanner(strings.NewReader(src))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid frontmatter line %q", line)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		out[key] = parseScalar(val)
	}
	return out, sc.Err()
}

func parseScalar(v string) any {
	v = strings.TrimSpace(v)
	if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
		return v[1 : len(v)-1]
	}
	if v == "true" {
		return true
	}
	if v == "false" {
		return false
	}
	if i, err := strconv.Atoi(v); err == nil {
		return i
	}
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		return parseInlineList(v[1 : len(v)-1])
	}
	return v
}

func parseInlineList(v string) []string {
	var result []string
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, strings.Trim(p, "\"'"))
		}
	}
	return result
}

func parseSteps(body string) ([]Step, string, error) {
	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	var steps []Step
	var current *Step
	var buf []string
	flush := func() error {
		if current == nil {
			return nil
		}
		blocks, err := parseBlocks(strings.Join(buf, "\n"))
		if err != nil {
			return err
		}
		current.Blocks = blocks
		steps = append(steps, *current)
		current = nil
		buf = nil
		return nil
	}
	var conclusion []string
	inConclusion := false
	for _, line := range lines {
		if stepStart.MatchString(strings.TrimSpace(line)) {
			if err := flush(); err != nil {
				return nil, "", err
			}
			attrs := attributes(stepStart.FindStringSubmatch(strings.TrimSpace(line))[1])
			label := attrs["label"]
			if label == "" {
				return nil, "", errors.New("step missing label")
			}
			current = &Step{ID: slugify(label), Label: label, Duration: attrs["duration"]}
			inConclusion = false
			continue
		}
		if stepEnd.MatchString(strings.TrimSpace(line)) {
			if err := flush(); err != nil {
				return nil, "", err
			}
			continue
		}
		if current != nil {
			buf = append(buf, line)
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "#") && strings.Contains(strings.ToLower(line), "conclusion") {
			inConclusion = true
			continue
		}
		if inConclusion {
			conclusion = append(conclusion, line)
		}
	}
	if err := flush(); err != nil {
		return nil, "", err
	}
	return steps, strings.TrimSpace(strings.Join(conclusion, "\n")), nil
}

func parseBlocks(src string) ([]Block, error) {
	lines := strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n")
	var blocks []Block
	var text []string
	flushText := func() {
		t := strings.TrimSpace(strings.Join(text, "\n"))
		if t != "" {
			blocks = append(blocks, Block{Type: "markdown", Content: t})
		}
		text = nil
	}
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "```") {
			text = append(text, lines[i])
			continue
		}
		flushText()
		head := strings.TrimSpace(strings.TrimPrefix(line, "```"))
		j := i + 1
		var code []string
		for ; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == "```" {
				break
			}
			code = append(code, lines[j])
		}
		if j >= len(lines) {
			return nil, errors.New("unterminated fenced code block")
		}
		b := Block{Type: "code", Content: strings.TrimSuffix(strings.Join(code, "\n"), "\n")}
		language, attrText := splitFenceHead(head)
		b.Language = language
		attrs := attributes(attrText)
		b.Attributes = attrs
		b.Title = attrs["title"]
		b.Level = attrs["level"]
		b.Copy = boolString(attrs["copy"])
		b.Terminal = boolString(attrs["terminal"])
		b.FileID = attrs["file-id"]
		b.CmdID = attrs["cmd-id"]
		if language == "info" {
			b.Type = "callout"
			b.Attributes = nil
		}
		if language == "warn" {
			b.Type = "callout"
			b.Attributes = nil
		}
		if language == "hint" {
			b.Type = "hint"
			b.Attributes = attrs
		}
		if language == "check" {
			b.Type = "check"
			b.Attributes = attrs
		}
		if language == "mermaid" {
			b.Type = "diagram"
		}
		if language == "flashcard" {
			b.Type = "flashcard"
		}
		if language == "challenge" {
			b.Type = "challenge"
		}
		i = j
		blocks = append(blocks, b)
	}
	flushText()
	return blocks, nil
}

func splitFenceHead(head string) (string, string) {
	idx := strings.Index(head, "{")
	if idx < 0 {
		return head, ""
	}
	lang := strings.TrimSpace(head[:idx])
	attrs := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(head[idx:], "{"), "}"))
	return lang, attrs
}

func attributes(s string) map[string]string {
	out := map[string]string{}
	for _, m := range attrRE.FindAllStringSubmatch(s, -1) {
		v := m[2]
		if v == "" {
			v = m[3]
		}
		if v == "" {
			v = m[4]
		}
		out[m[1]] = v
	}
	return out
}

func stringValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}
func boolValue(v any) bool     { b, _ := v.(bool); return b }
func boolString(v string) bool { return strings.EqualFold(v, "true") }
func intValue(v any) int       { i, _ := v.(int); return i }
func stringSlice(v any) []string {
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	dash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}
func FindMarkdown(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
