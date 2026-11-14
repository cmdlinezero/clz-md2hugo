package questionbank

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ApplyPromotion rechecks the approved plan against the on-disk taxonomy and
// preserves unknown curator metadata using raw JSON objects.
// It creates an exclusive backup before atomically replacing the source.
func ApplyPromotion(taxonomyPath, backupPath string, c CurationReport, approvals PromotionApprovals) (PromotionPlan, error) {
	var zero PromotionPlan
	if backupPath == "" {
		return zero, errors.New("explicit backup path required")
	}
	abs, err := filepath.Abs(taxonomyPath)
	if err != nil {
		return zero, err
	}
	bak, err := filepath.Abs(backupPath)
	if err != nil {
		return zero, err
	}
	if abs == bak {
		return zero, errors.New("backup cannot overwrite taxonomy")
	}
	if len(approvals.Add)+len(approvals.Aliases)+len(approvals.Parents) == 0 {
		return zero, errors.New("apply requires nonempty explicit approvals")
	}
	// Exclusive lock prevents concurrent invocations of this writer.
	lock := abs + ".promotion.lock"
	lf, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return zero, fmt.Errorf("promotion lock: %w", err)
	}
	lf.Close()
	defer os.Remove(lock)
	info, err := os.Stat(abs)
	if err != nil {
		return zero, err
	}
	if !info.Mode().IsRegular() {
		return zero, errors.New("taxonomy must be a regular file")
	}
	original, err := os.ReadFile(abs)
	if err != nil {
		return zero, err
	}
	tax, err := LoadTaxonomy(abs)
	if err != nil {
		return zero, err
	}
	plan, err := PlanPromotion(tax, c, approvals)
	if err != nil {
		return zero, err
	}
	if len(plan.Warnings) > 0 {
		return zero, fmt.Errorf("unresolved promotion warnings: %s", strings.Join(plan.Warnings, "; "))
	}
	var root map[string]json.RawMessage
	if err = json.Unmarshal(original, &root); err != nil {
		return zero, err
	}
	var technologies []map[string]json.RawMessage
	if err = json.Unmarshal(root["technologies"], &technologies); err != nil {
		return zero, err
	}
	byID := map[string]int{}
	for i, e := range technologies {
		var id string
		if err = json.Unmarshal(e["id"], &id); err != nil {
			return zero, err
		}
		byID[id] = i
	}
	for _, e := range plan.Add {
		if _, ok := byID[e.ID]; ok {
			return zero, fmt.Errorf("duplicate ID %s", e.ID)
		}
		obj := map[string]json.RawMessage{}
		put(obj, "id", e.ID)
		put(obj, "label", e.Label)
		put(obj, "aliases", []string{})
		put(obj, "review_status", "curator_approved")
		if item := findCuration(c.Items, e.ID); item != nil {
			put(obj, "categories", item.Categories)
			put(obj, "provenance", []EntryProvenance{curationEvidence(c, *item, "curator-approved technology addition")})
		}
		byID[e.ID] = len(technologies)
		technologies = append(technologies, obj)
	}
	for _, a := range plan.Aliases {
		idx, ok := byID[a.CanonicalID]
		if !ok {
			return zero, fmt.Errorf("missing alias target %s", a.CanonicalID)
		}
		var aliases []string
		if raw := technologies[idx]["aliases"]; len(raw) > 0 {
			if err = json.Unmarshal(raw, &aliases); err != nil {
				return zero, err
			}
		}
		found := false
		for _, v := range aliases {
			if strings.EqualFold(strings.TrimSpace(v), strings.TrimSpace(a.Label)) {
				found = true
			}
		}
		if !found {
			aliases = append(aliases, a.Label)
		}
		put(technologies[idx], "aliases", aliases)
		if item := findCuration(c.Items, slug(a.Label)); item != nil {
			appendEvidence(technologies[idx], curationEvidence(c, *item, "curator-approved alias: "+a.Label))
		}
	}
	// Features are first-class technology nodes; feature_of edges are not prerequisites.
	var relations []map[string]json.RawMessage
	if raw := root["technology_relationships"]; len(raw) > 0 {
		if err = json.Unmarshal(raw, &relations); err != nil {
			return zero, err
		}
	}
	for _, edge := range plan.Parents {
		item := findCuration(c.Items, edge.ChildID)
		if item == nil {
			return zero, fmt.Errorf("missing feature %s", edge.ChildID)
		}
		if _, ok := byID[edge.ChildID]; ok {
			return zero, fmt.Errorf("feature ID collision %s", edge.ChildID)
		}
		obj := map[string]json.RawMessage{}
		put(obj, "id", edge.ChildID)
		put(obj, "label", item.Label)
		put(obj, "aliases", []string{})
		put(obj, "review_status", "curator_approved")
		put(obj, "kind", "feature")
		put(obj, "categories", item.Categories)
		put(obj, "provenance", []EntryProvenance{curationEvidence(c, *item, "curator-approved feature_of relationship to "+edge.ParentID)})
		byID[edge.ChildID] = len(technologies)
		technologies = append(technologies, obj)
		rel := map[string]json.RawMessage{}
		put(rel, "child_id", edge.ChildID)
		put(rel, "parent_id", edge.ParentID)
		put(rel, "relation", "feature_of")
		relations = append(relations, rel)
	}
	put(root, "technologies", technologies)
	if len(plan.Parents) > 0 {
		put(root, "technology_relationships", relations)
	}
	// Preserve existing attribution fields and document this specific adaptation.
	var attr map[string]json.RawMessage
	if err = json.Unmarshal(root["attribution"], &attr); err != nil {
		return zero, err
	}
	if attr == nil {
		attr = map[string]json.RawMessage{}
	}
	put(attr, "source_commit", c.SourceCommit)
	var notice string
	_ = json.Unmarshal(attr["adaptation_notice"], &notice)
	notice += " Promotion: curator-approved Google Cloud in 4 Words labels and feature relationships; source commit " + c.SourceCommit + "."
	put(attr, "adaptation_notice", strings.TrimSpace(notice))
	put(root, "attribution", attr)
	// Version remains unchanged intentionally: curator must version the pack in a separate reviewed commit.
	output, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return zero, err
	}
	output = append(output, '\n')
	// Validate the exact proposed bytes before touching the source.
	temp, err := os.CreateTemp(filepath.Dir(abs), ".taxonomy-promotion-*.json")
	if err != nil {
		return zero, err
	}
	tmp := temp.Name()
	defer os.Remove(tmp)
	if err = temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return zero, err
	}
	if _, err = temp.Write(output); err != nil {
		temp.Close()
		return zero, err
	}
	if err = temp.Sync(); err != nil {
		temp.Close()
		return zero, err
	}
	if err = temp.Close(); err != nil {
		return zero, err
	}
	proposed, err := LoadTaxonomy(tmp)
	if err != nil {
		return zero, fmt.Errorf("proposed taxonomy invalid: %w", err)
	}
	if len(proposed.Technologies) != len(tax.Technologies)+len(plan.Add)+len(plan.Parents) {
		return zero, errors.New("unexpected technology count")
	}
	// Refuse to clobber a previous backup, including on repeated apply.
	bf, err := os.OpenFile(bak, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return zero, fmt.Errorf("backup: %w", err)
	}
	if _, err = bf.Write(original); err != nil {
		bf.Close()
		os.Remove(bak)
		return zero, err
	}
	if err = bf.Sync(); err != nil {
		bf.Close()
		os.Remove(bak)
		return zero, err
	}
	if err = bf.Close(); err != nil {
		os.Remove(bak)
		return zero, err
	}
	if err = os.Rename(tmp, abs); err != nil {
		return zero, fmt.Errorf("atomic replace failed; original remains and backup is %s: %w", bak, err)
	}
	plan.DryRun = false
	plan.Notice = "Applied to authoritative taxonomy; backup: " + bak + ". Source categories are not exam domains. Review and commit taxonomy version separately."
	return plan, nil
}
func put(obj map[string]json.RawMessage, k string, v any) { b, _ := json.Marshal(v); obj[k] = b }
func findCuration(items []CurationItem, id string) *CurationItem {
	for i := range items {
		if items[i].SuggestedID == id {
			return &items[i]
		}
	}
	return nil
}

func curationEvidence(c CurationReport, item CurationItem, modification string) EntryProvenance {
	return EntryProvenance{Source: c.Source, SourceCommit: c.SourceCommit, License: c.License, SourceLine: item.SourceLine, Modification: modification}
}
func appendEvidence(obj map[string]json.RawMessage, evidence EntryProvenance) {
	var previous []EntryProvenance
	_ = json.Unmarshal(obj["provenance"], &previous)
	previous = append(previous, evidence)
	put(obj, "provenance", previous)
}
