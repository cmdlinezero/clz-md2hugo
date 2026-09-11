package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var allowedTypes = map[string]bool{
	"collection":      true,
	"learning-path":   true,
	"certification":   true,
	"tutorial":        true,
	"coding-tutorial": true,
	"quiz":            true,
	"question":        true,
	"topic":           true,
	"flashcard-deck":  true,
	"challenge-lab":   true,
}

func Build(input, output string, includeUnpublished bool) error {
	contents, err := compile(input)
	if err != nil {
		return err
	}
	if !includeUnpublished {
		contents = publishedOnly(contents)
	}
	if err := validateRelationships(contents); err != nil {
		return err
	}
	graph, err := graphFrom(contents)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if dir := filepath.Dir(output); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(output, data, 0644)
}

func Validate(input string) error {
	_, err := compile(input)
	return err
}

func compile(input string) ([]Content, error) {
	files, err := FindMarkdown(input)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no markdown files found under %s", input)
	}

	contents := make([]Content, 0, len(files))
	seen := map[string]string{}
	for _, file := range files {
		c, err := ParseFile(file)
		if err != nil {
			return nil, err
		}
		if rel, err := filepath.Rel(input, file); err == nil {
			c.Source = filepath.ToSlash(rel)
		}
		if err := validateContent(c, file); err != nil {
			return nil, err
		}
		if prev, ok := seen[c.ID]; ok {
			return nil, fmt.Errorf("duplicate content id %q in %s and %s", c.ID, prev, file)
		}
		seen[c.ID] = file
		contents = append(contents, c)
	}
	if err := validateRelationships(contents); err != nil {
		return nil, err
	}
	return contents, nil
}

func validateContent(c Content, file string) error {
	if strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("%s: missing id", file)
	}
	if strings.TrimSpace(c.Title) == "" {
		return fmt.Errorf("%s: missing title", file)
	}
	if !allowedTypes[c.Type] {
		return fmt.Errorf("%s: unsupported type %q", file, c.Type)
	}
	if c.Status != "draft" && c.Status != "review" && c.Status != "published" && c.Status != "archived" {
		return fmt.Errorf("%s: invalid status %q", file, c.Status)
	}
	if c.Type == "question" {
		if len(c.Options) < 2 {
			return fmt.Errorf("%s: question requires at least 2 options", file)
		}
		if c.Answer < 0 || c.Answer >= len(c.Options) {
			return fmt.Errorf("%s: answer index %d is outside options range", file, c.Answer)
		}
	}
	if c.Type == "flashcard-deck" {
		if len(c.Cards) == 0 {
			return fmt.Errorf("%s: flashcard-deck requires at least 1 flashcard", file)
		}
		for _, card := range c.Cards {
			if card.ID == "" || card.Front == "" || card.Back == "" {
				return fmt.Errorf("%s: flashcards require id, front, and back", file)
			}
		}
	}
	if c.Type == "challenge-lab" {
		if len(c.Challenges) == 0 {
			return fmt.Errorf("%s: challenge-lab requires at least 1 challenge", file)
		}
		for _, ch := range c.Challenges {
			if ch.ID == "" || ch.Title == "" {
				return fmt.Errorf("%s: challenge steps require id and title", file)
			}
		}
	}
	if c.Type == "coding-tutorial" {
		if c.Runtime != "python3" && c.Runtime != "go" && c.Runtime != "nodejs" {
			return fmt.Errorf("%s: coding-tutorial runtime must be one of python3, go, nodejs", file)
		}
		if len(c.Steps) == 0 {
			return fmt.Errorf("%s: coding-tutorial requires at least 1 step", file)
		}
		for i, step := range c.Steps {
			if strings.TrimSpace(step.ID) == "" || strings.TrimSpace(step.Label) == "" {
				return fmt.Errorf("%s: coding-tutorial step %d requires id and label", file, i+1)
			}
			if strings.TrimSpace(step.ExpectedOutput) == "" {
				return fmt.Errorf("%s: coding-tutorial step %q requires a check block with output=...", file, step.ID)
			}
		}
		if strings.TrimSpace(c.Steps[0].StarterCode) == "" {
			return fmt.Errorf("%s: first coding-tutorial step requires an editor=true code block", file)
		}
	}
	return nil
}

func validateRelationships(contents []Content) error {
	ids := make(map[string]Content, len(contents))
	for _, c := range contents {
		ids[c.ID] = c
	}
	for _, c := range contents {
		refs := []struct {
			field  string
			values []string
		}{
			{"collections", c.Collections},
			{"certifications", c.Certifications},
			{"prerequisites", c.Prerequisites},
			{"children", c.Children},
			{"activities", c.Activities},
		}
		for _, ref := range refs {
			for _, id := range ref.values {
				if _, ok := ids[id]; !ok {
					return fmt.Errorf("%s %q references missing %s %q", c.Type, c.ID, ref.field, id)
				}
			}
		}
	}
	return nil
}

func graphFrom(contents []Content) (ContentGraph, error) {
	graph := ContentGraph{Version: 2, Generated: time.Now().UTC().Format(time.RFC3339)}
	for _, c := range contents {
		graph.Stats.Documents++
		graph.Stats.Steps += len(c.Steps)
		for _, s := range c.Steps {
			graph.Stats.Blocks += len(s.Blocks)
		}
		switch c.Type {
		case "collection":
			graph.Collections = append(graph.Collections, c)
			graph.Stats.Collections++
		case "learning-path":
			graph.LearningPaths = append(graph.LearningPaths, c)
			graph.Stats.LearningPaths++
		case "certification":
			graph.Certifications = append(graph.Certifications, c)
			graph.Stats.Certifications++
		case "tutorial":
			graph.Tutorials = append(graph.Tutorials, c)
			graph.Stats.Tutorials++
		case "coding-tutorial":
			graph.CodingTutorials = append(graph.CodingTutorials, c)
			graph.Stats.CodingTutorials++
		case "quiz":
			graph.Quizzes = append(graph.Quizzes, c)
			graph.Stats.Quizzes++
		case "question":
			graph.Questions = append(graph.Questions, c)
			graph.Stats.Questions++
		case "topic":
			graph.Topics = append(graph.Topics, c)
			graph.Stats.Topics++
		case "flashcard-deck":
			graph.FlashcardDecks = append(graph.FlashcardDecks, c)
			graph.Stats.FlashcardDecks++
		case "challenge-lab":
			graph.ChallengeLabs = append(graph.ChallengeLabs, c)
			graph.Stats.ChallengeLabs++
		default:
			graph.Other = append(graph.Other, c)
		}
	}
	return graph, nil
}

func publishedOnly(contents []Content) []Content {
	result := make([]Content, 0, len(contents))
	for _, c := range contents {
		if c.Status == "published" {
			result = append(result, c)
		}
	}
	return result
}
