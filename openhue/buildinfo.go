package openhue

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// NewBuildInfo creates a new BuildInfo container
func NewBuildInfo(version string, commit string, date string) *BuildInfo {
	return &BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
}

// NewTestBuildInfo returns a valid openhue.BuildInfo just for testing
func NewTestBuildInfo() *BuildInfo {
	return NewBuildInfo("1.0.0", "1234", "now")
}
