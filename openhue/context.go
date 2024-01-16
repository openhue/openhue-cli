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
func NewContext(io IOStreams, buildInfo *BuildInfo, api *gen.ClientWithResponses) *Context {
	return &Context{
		Io:        io,
		BuildInfo: buildInfo,
		Api:       api,
	}
}

func NewTestContextWithoutApi() (*Context, *bytes.Buffer) {
	streams, _, out, _ := NewTestIOStreams()
	ctx := NewContext(streams, NewTestBuildInfo(), nil)
	return ctx, out
}
