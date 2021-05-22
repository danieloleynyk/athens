package commands

import (
	"context"
	"github.com/gomods/athens/pkg/dumper"
	"github.com/spf13/cobra"
	"log"
)

var (
	outDir     string
	OutDirFlag = "output"
)

func NewDumpCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "dump <go.mod dir>",
		Short: "dumps go modules into a tar.gz",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if _, err := cmd.Flags().GetString(OutDirFlag); err != nil {
				log.Println(err)
				return err
			}

			return dumper.DumpModules(ctx, args[0], outDir)
		},
	}
	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(
		&outDir,
		OutDirFlag,
		"o",
		".",
		"output directory of the packed tar.gz",
	)

	return &cmd
}
