package commands

import (
	"context"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/dumper"
	"github.com/gomods/athens/pkg/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	outDir     string
	OutDirFlag = "output"
)

func NewDumpCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "dump <go.mod dir>",
		Short: "dumps go modules into a tar.gz",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if _, err := cmd.Flags().GetString(ConfigDirFlag); err != nil {
				return err
			}

			cfg, err := config.Load(configDir)
			if err != nil {
				return err
			}

			if _, err := cmd.Flags().GetString(OutDirFlag); err != nil {
				return err
			}

			logLvl, err := logrus.ParseLevel(cfg.LogLevel)
			if err != nil {
				return err
			}
			lggr := log.New(cfg.CloudRuntime, logLvl)

			return dumper.DumpModules(ctx, args[0], outDir, cfg, lggr)
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

	cmd.Flags().StringVarP(
		&configDir,
		ConfigDirFlag,
		"c",
		"",
		"athens config directory",
	)

	return &cmd
}
