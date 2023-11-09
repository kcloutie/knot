package run

import (
	"github.com/kcloutie/knot/pkg/cli"
	"github.com/kcloutie/knot/pkg/params"
	"github.com/spf13/cobra"
)

func Root(cliParams *params.Run, ioStreams *cli.IOStreams) *cobra.Command {
	cCmd := &cobra.Command{
		Use:     "run",
		Aliases: []string{},
		Short:   "Runs the web/api server",
	}
	cCmd.AddCommand(ServerCommand(cliParams, ioStreams))
	return cCmd
}
