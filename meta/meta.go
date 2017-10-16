package meta

// Meta variables set dynamically by the linker at build time
var (
	Version   string
	GitCommit string
)

// FullVersion returns the version string in the form of
// ([major].[minor].[patch]+[commit])
func FullVersion() string {
	return Version + "+" + GitCommit
}
