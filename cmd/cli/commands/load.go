package commands

import (
	"context"
	"github.com/gomods/athens/cmd/cli/actions"
	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/loader"
	"github.com/spf13/cobra"
)

func NewLoadCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "load <athens_dump.tar.gz dir>",
		Short: "loads into athens",
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

			app, err := actions.NewApp(cfg)
			if err != nil {
				return err
			}

			return loader.LoadModules(ctx, args[0], app.Store, app.Lggr)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().StringVarP(
		&configDir,
		ConfigDirFlag,
		"c",
		"",
		"athens config directory",
	)

	return &cmd
}
