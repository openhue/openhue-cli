package openhue

import (
	"bytes"
	"github.com/openhue/openhue-go"
)

// Context contains the common objects for the different commands.
type Context struct {
	Io        IOStreams
	BuildInfo *BuildInfo
	H         *openhue.Home
	Home      *HomeModel
	Config    *Config
}

// NewContext returns an initialized Context from a given openhue.ClientWithResponses API with default IOStreams
func NewContext(io IOStreams, buildInfo *BuildInfo, home *openhue.Home, config *Config) *Context {
	return &Context{
		Io:        io,
		BuildInfo: buildInfo,
		H:         home,
		Config:    config,
	}
}

// NewTestContext returns a Context for testing usage only and the out buffer to validate
// the command output.
// The Api field of the returned Context is not set.
func NewTestContext(home *HomeModel) (*Context, *bytes.Buffer) {
	streams, _, out, _ := NewTestIOStreams()
	ctx := NewContext(streams, NewTestBuildInfo(), nil, nil)
	ctx.Home = home
	return ctx, out
}
