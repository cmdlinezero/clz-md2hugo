package compiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Project compiles content once and writes two deliberately separate classes
// of artifact: a public-safe catalogue and tier-specific full payloads.
// payloadDir must be deployed behind entitlement enforcement; it must not be
// copied into Hugo's public/static output for member or pro tiers.
func Project(input, catalogueOutput, payloadDir string, includeUnpublished bool) error {
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

	generated := time.Now().UTC().Format(time.RFC3339)
	catalogue := catalogueFrom(contents, generated)
	if err := writeJSON(catalogueOutput, catalogue); err != nil {
		return err
	}

	for _, tier := range []string{"public", "member", "pro"} {
		projection := payloadProjectionFrom(contents, tier, generated)
		path := filepath.Join(payloadDir, tier+".json")
		if err := writeJSON(path, projection); err != nil {
			return err
		}
	}
	return nil
}

func catalogueFrom(contents []Content, generated string) Catalogue {
	items := make([]CatalogueItem, 0, len(contents))
	for _, c := range contents {
		items = append(items, CatalogueItem{
			ID: c.ID, Type: c.Type, Title: c.Title, Description: c.Description,
			Slug: c.Slug, Kind: c.Kind, Provider: c.Provider, RatingID: c.RatingID,
			ProductID: c.ProductID, Access: c.Access, Runtime: c.Runtime,
			Difficulty: c.Difficulty, Categories: c.Categories, Tags: c.Tags,
			Topics: c.Topics, Collections: c.Collections, Certifications: c.Certifications,
			Prerequisites: c.Prerequisites, Children: c.Children, Activities: c.Activities,
			Duration: c.Duration, HeroTitle: c.HeroTitle, HeroImage: c.HeroImage,
			Issue: c.Issue, Volume: c.Volume, Special: c.Special,
			StepCount: len(c.Steps), CardCount: len(c.Cards), ChallengeCount: len(c.Challenges),
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].ID < items[j].ID
		}
		return items[i].Type < items[j].Type
	})
	return Catalogue{Version: 1, Generated: generated, Items: items}
}

func payloadProjectionFrom(contents []Content, tier, generated string) PayloadProjection {
	items := make([]Content, 0)
	for _, c := range contents {
		if c.Access.Tier == tier {
			items = append(items, c)
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Type == items[j].Type {
			return items[i].ID < items[j].ID
		}
		return items[i].Type < items[j].Type
	})
	return PayloadProjection{Version: 1, Generated: generated, Tier: tier, Items: items}
}

func writeJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("output path is required")
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}
