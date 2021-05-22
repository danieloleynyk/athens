package commands

import "github.com/spf13/cobra"

// NewDefaultCommand creates a new default command for when
// the user does not provide a command
func NewDefaultCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "athens <subcommand>",
		Short: "Command line tool to assist with packaging Go modules for athens",
	}

	cmd.AddCommand(NewDumpCommand())

	return &cmd
}
