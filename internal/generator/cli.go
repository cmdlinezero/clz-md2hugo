package generator

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var configPath string
	var check bool
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate deterministic Hugo content from structured data",
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := Run(Options{ConfigPath: configPath, Check: check})
			if err != nil {
				return err
			}
			if check {
				if len(result.Changed) > 0 {
					return fmt.Errorf("generated content is stale: %d file(s) would change", len(result.Changed))
				}
				fmt.Fprintln(cmd.OutOrStdout(), "generated content is current")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "updated %d generated file(s)\n", len(result.Changed))
			return nil
		},
	}
	cmd.Flags().StringVarP(&configPath, "config", "c", "md2hugo.generate.json", "Generator configuration file")
	cmd.Flags().BoolVar(&check, "check", false, "Check generated content without writing files")
	return cmd
}
