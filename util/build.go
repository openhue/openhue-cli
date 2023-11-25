package util

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
