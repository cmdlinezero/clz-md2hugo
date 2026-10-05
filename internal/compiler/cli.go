package compiler

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func Execute() {
	root := &cobra.Command{Use: "certin-content", Short: "Compile Certin Markdown content into JSON"}
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
	root.AddCommand(build, project, validate)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
