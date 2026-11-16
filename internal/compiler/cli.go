package compiler

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cmdlinezero/clz-md2hugo/internal/generator"
	"github.com/cmdlinezero/clz-md2hugo/internal/questionbank"
)

const CompilerVersion = "portable-markdown-v3"

func Execute() {
	root := &cobra.Command{Use: "certin-content", Short: "Compile Certin Markdown content into JSON", Version: CompilerVersion}
	build := &cobra.Command{Use: "build", Short: "Validate and compile Markdown content", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		output, _ := cmd.Flags().GetString("output")
		includeUnpublished, _ := cmd.Flags().GetBool("include-unpublished")
		if err := Build(input, output, includeUnpublished); err != nil {
			return err
		}
		fmt.Printf("built %s\n", output)
		return nil
	}}
	build.Flags().StringP("input", "i", "examples", "Markdown content directory")
	build.Flags().StringP("output", "o", "dist/content.json", "Generated JSON path")
	build.Flags().Bool("include-unpublished", false, "Include draft, review, and archived content in the artifact")
	project := &cobra.Command{Use: "project", Short: "Build public catalogue and tier payload projections", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		catalogue, _ := cmd.Flags().GetString("catalogue")
		payloadDir, _ := cmd.Flags().GetString("payload-dir")
		includeUnpublished, _ := cmd.Flags().GetBool("include-unpublished")
		if err := Project(input, catalogue, payloadDir, includeUnpublished); err != nil {
			return err
		}
		fmt.Printf("built catalogue %s and tier payloads under %s\n", catalogue, payloadDir)
		return nil
	}}
	project.Flags().StringP("input", "i", "examples", "Markdown content directory")
	project.Flags().String("catalogue", "dist/catalogue.json", "Public-safe catalogue JSON path")
	project.Flags().String("payload-dir", "private/payloads", "Tier payload directory; keep member/pro outside Hugo public output")
	project.Flags().Bool("include-unpublished", false, "Include draft, review, and archived content in projections")
	validate := &cobra.Command{Use: "validate", Short: "Validate Markdown content", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		if err := Validate(input); err != nil {
			return err
		}
		fmt.Printf("valid: %s\n", input)
		return nil
	}}
	validate.Flags().StringP("input", "i", "examples", "Markdown content directory")
	taxonomy := &cobra.Command{Use: "taxonomy", Short: "Curated taxonomy tooling"}
	taxonomyValidate := &cobra.Command{Use: "validate", Short: "Validate a taxonomy pack", RunE: func(cmd *cobra.Command, args []string) error {
		path, _ := cmd.Flags().GetString("file")
		t, e := questionbank.LoadTaxonomy(path)
		if e != nil {
			return e
		}
		fmt.Printf("valid taxonomy %s (%d domains, %d technologies)\n", t.ID, len(t.Domains), len(t.Technologies))
		return nil
	}}
	taxonomyValidate.Flags().String("file", "taxonomies/google-cloud/taxonomy.json", "Taxonomy JSON file")
	taxonomy.AddCommand(taxonomyValidate)
	taxonomyImport := &cobra.Command{Use: "import", Short: "Reconcile a local Google Cloud in 4 Words README with a curated taxonomy", RunE: func(cmd *cobra.Command, args []string) error {
		source, _ := cmd.Flags().GetString("source")
		taxonomyPath, _ := cmd.Flags().GetString("taxonomy")
		output, _ := cmd.Flags().GetString("output")
		commit, _ := cmd.Flags().GetString("source-commit")
		t, err := questionbank.LoadTaxonomy(taxonomyPath)
		if err != nil {
			return err
		}
		if t.ID != "google-cloud" {
			return fmt.Errorf("cheat sheet importer requires google-cloud taxonomy, got %q", t.ID)
		}
		candidates, err := questionbank.ParseCheatSheet(source)
		if err != nil {
			return err
		}
		report := questionbank.ReconcileCheatSheet(t, candidates, commit)
		if err := questionbank.Write(output, report); err != nil {
			return err
		}
		fmt.Printf("reconciled %d entries: %d matched, %d proposed, %d review; report %s\n", report.Candidates, len(report.Matched), len(report.Proposed), len(report.Review), output)
		return nil
	}}
	taxonomyImport.Flags().String("source", "README.md", "Locally downloaded upstream README.md")
	taxonomyImport.Flags().String("taxonomy", "taxonomies/google-cloud/taxonomy.json", "Existing curated taxonomy")
	taxonomyImport.Flags().String("output", "dist/google-cloud-reconciliation.json", "Review report JSON")
	taxonomyImport.Flags().String("source-commit", "", "Optional upstream commit SHA for provenance")
	taxonomy.AddCommand(taxonomyImport)
	taxonomyCurate := &cobra.Command{Use: "curate", Short: "Produce a conservative review queue from a cheat-sheet reconciliation", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		taxonomyPath, _ := cmd.Flags().GetString("taxonomy")
		output, _ := cmd.Flags().GetString("output")
		t, err := questionbank.LoadTaxonomy(taxonomyPath)
		if err != nil {
			return err
		}
		var reconciliation questionbank.Reconciliation
		if err := questionbank.ReadReconciliation(input, &reconciliation); err != nil {
			return err
		}
		report, err := questionbank.CurateReconciliation(t, reconciliation)
		if err != nil {
			return err
		}
		if err := questionbank.Write(output, report); err != nil {
			return err
		}
		fmt.Printf("curation queue: %d items, %d matched; report %s\n", len(report.Items), report.MatchedCount, output)
		return nil
	}}
	taxonomyCurate.Flags().String("input", "dist/google-cloud-reconciliation.json", "Reconciliation JSON")
	taxonomyCurate.Flags().String("taxonomy", "taxonomies/google-cloud/taxonomy.json", "Curated taxonomy JSON")
	taxonomyCurate.Flags().String("output", "dist/google-cloud-curation.json", "Review queue JSON")
	taxonomy.AddCommand(taxonomyCurate)

	taxonomyPromote := &cobra.Command{Use: "promote", Short: "Generate an approval-gated taxonomy promotion dry run", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		taxonomyPath, _ := cmd.Flags().GetString("taxonomy")
		approvalsPath, _ := cmd.Flags().GetString("approvals")
		output, _ := cmd.Flags().GetString("output")
		t, err := questionbank.LoadTaxonomy(taxonomyPath)
		if err != nil {
			return err
		}
		c, err := questionbank.ReadCuration(input)
		if err != nil {
			return err
		}
		a, err := questionbank.ReadPromotionApprovals(approvalsPath)
		if err != nil {
			return err
		}
		plan, err := questionbank.PlanPromotion(t, c, a)
		if err != nil {
			return err
		}
		apply, _ := cmd.Flags().GetBool("apply")
		if apply {
			backup, _ := cmd.Flags().GetString("backup")
			plan, err = questionbank.ApplyPromotion(taxonomyPath, backup, c, a)
			if err != nil {
				return err
			}
		}
		if err := questionbank.Write(output, plan); err != nil {
			return err
		}
		fmt.Printf("promotion (applied=%t): %d additions, %d aliases, %d parent edges; %s\n", apply, len(plan.Add), len(plan.Aliases), len(plan.Parents), output)
		return nil
	}}
	taxonomyPromote.Flags().String("input", "dist/google-cloud-curation.json", "Curation report JSON")
	taxonomyPromote.Flags().String("taxonomy", "taxonomies/google-cloud/taxonomy.json", "Authoritative taxonomy JSON (read only)")
	taxonomyPromote.Flags().String("approvals", "taxonomies/google-cloud/promotion-approvals.json", "Explicit curator selections")
	taxonomyPromote.Flags().String("output", "dist/google-cloud-promotion-plan.json", "Dry-run plan JSON")
	taxonomyPromote.Flags().Bool("apply", false, "Explicitly apply approved changes to authoritative taxonomy")
	taxonomyPromote.Flags().String("backup", "", "Required unique backup file path when applying")
	taxonomyPromote.Flags().Bool("dry-run", true, "Legacy compatibility flag; use --apply to write")
	_ = taxonomyPromote.Flags().MarkHidden("dry-run")
	taxonomy.AddCommand(taxonomyPromote)

	quiz := &cobra.Command{Use: "quiz", Short: "Legacy JSON question-bank tooling"}
	for _, action := range []string{"validate", "ingest"} {
		action := action
		c := &cobra.Command{Use: action, RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := cmd.Flags().GetString("input")
			taxonomyPath, _ := cmd.Flags().GetString("taxonomy")
			bankID, _ := cmd.Flags().GetString("bank")
			out, _ := cmd.Flags().GetString("output")
			bank, e := questionbank.Ingest(input, taxonomyPath, bankID)
			if e != nil {
				return e
			}
			if action == "validate" {
				fmt.Printf("valid bank %s: %d questions\n", bank.ID, len(bank.Questions))
				return nil
			}
			if e := questionbank.Write(out, bank); e != nil {
				return e
			}
			fmt.Printf("ingested %d questions to %s\n", len(bank.Questions), out)
			return nil
		}}
		c.Flags().String("input", "questions", "Quiz JSON file or directory")
		c.Flags().String("taxonomy", "taxonomies/google-cloud/taxonomy.json", "Taxonomy JSON file")
		c.Flags().String("bank", "question-bank", "Canonical bank ID")
		c.Flags().String("output", "dist/question-bank.json", "Output file (ingest only)")
		quiz.AddCommand(c)
	}
	root.AddCommand(build, project, validate, taxonomy, quiz, generator.Command())
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
