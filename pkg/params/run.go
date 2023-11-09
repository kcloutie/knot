package params

import (
	"strings"
)

const (
	PACConfigmapName        = "pipelines-as-code"
	StartingPipelineRunText = `Starting Pipelinerun <b>%s</b> in namespace
  <b>%s</b><br><br>You can follow the execution on the [%s](%s) PipelineRun viewer or via
  the command line with :
	<br><code>%s pr logs -n %s %s -f</code>`
	QueuingPipelineRunText = `PipelineRun <b>%s</b> has been queued Queuing in namespace
  <b>%s</b><br><br>`
)

type Run struct {
	// Clients clients.Clients
	// Info    info.Info
}

func StringToBool(s string) bool {
	if strings.ToLower(s) == "true" ||
		strings.ToLower(s) == "yes" || s == "1" {
		return true
	}
	return false
}

func New() *Run {
	return &Run{
		// Info: info.Info{
		// 	Pac: &info.PacOpts{
		// 		Settings: &settings.Settings{
		// 			ApplicationName: settings.PACApplicationNameDefaultValue,
		// 			HubURL:          settings.HubURLDefaultValue,
		// 		},
		// 	},
		// },
	}
}
