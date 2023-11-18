package settings

const (
	CliBinaryName               = "knot"
	DebugModeLoggerEnvVar       = "KNOT_DEBUG"
	PubSubEndpoint              = "pubsub"
	GoTemplateDefaultDelimLeft  = "{{"
	GoTemplateDefaultDelimRight = "}}"
)

var (
	RootOptions      = RootFlags{}
	DebugModeEnabled = false
	IsQuiet          = false
)

type RootFlags struct {
	NoColor bool
}
