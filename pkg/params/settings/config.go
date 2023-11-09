package settings

const (
	CliBinaryName         = "knot"
	DebugModeLoggerEnvVar = "KNOT_DEBUG"
)

var (
	RootOptions      = RootFlags{}
	DebugModeEnabled = false
	IsQuiet          = false
)

type RootFlags struct {
	NoColor bool
}
