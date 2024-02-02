package openhue

import (
	"bytes"
	"openhue-cli/openhue/gen"
)

// Context contains the common objects for the different commands.
type Context struct {
	Io        IOStreams
	BuildInfo *BuildInfo
	Api       *gen.ClientWithResponses
	Home      *Home
	Config    *Config
}

// NewContext returns an initialized Context from a given gen.ClientWithResponses API with default IOStreams
func NewContext(io IOStreams, buildInfo *BuildInfo, api *gen.ClientWithResponses, config *Config) *Context {
	return &Context{
		Io:        io,
		BuildInfo: buildInfo,
		Api:       api,
		Config:    config,
	}
}

// NewTestContext returns a Context for testing usage only and the out buffer to validate
// the command output.
// The Api field of the returned Context is not set.
func NewTestContext(home *Home) (*Context, *bytes.Buffer) {
	streams, _, out, _ := NewTestIOStreams()
	ctx := NewContext(streams, NewTestBuildInfo(), nil, nil)
	ctx.Home = home
	return ctx, out
}
