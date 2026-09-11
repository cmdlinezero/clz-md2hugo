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
	validate := &cobra.Command{Use: "validate", Short: "Validate Markdown content", RunE: func(cmd *cobra.Command, args []string) error {
		input, _ := cmd.Flags().GetString("input")
		if err := Validate(input); err != nil {
			return err
		}
		fmt.Printf("valid: %s\n", input)
		return nil
	}}
	validate.Flags().StringP("input", "i", "examples", "Markdown content directory")
	root.AddCommand(build, validate)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
