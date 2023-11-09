package knot

import (
	"context"
	"fmt"

	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/kcloutie/knot/pkg/cli"
	"github.com/kcloutie/knot/pkg/cmd"
	"github.com/kcloutie/knot/pkg/cmd/knot/run"
	"github.com/kcloutie/knot/pkg/cmd/knot/version"
	"github.com/kcloutie/knot/pkg/logger"
	"github.com/kcloutie/knot/pkg/params"
	"github.com/kcloutie/knot/pkg/params/settings"
	"github.com/spf13/cobra"
)

var (
	showVersion = false
	ioStreams   = cli.NewIOStreams()
)

func Root(cliParams *params.Run) *cobra.Command {
	cCmd := &cobra.Command{
		Use:   "knot",
		Short: "knot is a cli/api tool for sending notifications",
		Long: heredoc.Doc(`
			knot is a cli/api tool for sending notifications
		`),
		SilenceUsage: false,
		PersistentPreRun: func(cCmd *cobra.Command, args []string) {
			lgr := logger.Get()
			lgr.Info("Starting application")
			if settings.DebugModeEnabled || os.Getenv(settings.DebugModeLoggerEnvVar) != "" {
				lgr.Info("Debugging has been enabled!")
			}

		},
		RunE: func(cCmd *cobra.Command, args []string) error {
			if showVersion {
				vopts := version.VersionCmdOptions{
					IoStreams: ioStreams,
					CliOpts:   cli.NewCliOptions(),
					Output:    "",
				}
				vopts.IoStreams.SetColorEnabled(!settings.RootOptions.NoColor)
				vopts.PrintVersion(context.Background())
				return nil
			}
			return fmt.Errorf("no command was specified")
		},
		Annotations: map[string]string{
			"commandType": "main",
		},
	}

	cCmd.PersistentFlags().BoolVar(&settings.DebugModeEnabled, "debug", false, "When set, additional output around debugging is output to the screen")
	cCmd.PersistentFlags().BoolVarP(&settings.RootOptions.NoColor, cmd.NoColorFlag, "C", false, "Disable coloring")
	cCmd.PersistentFlags().BoolVar(&showVersion, "version", false, "Show the version")
	cCmd.AddCommand(version.VersionCommand(ioStreams))
	cCmd.AddCommand(run.Root(cliParams, ioStreams))

	return cCmd
}
