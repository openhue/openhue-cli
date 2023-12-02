package openhue

import (
	"openhue-cli/openhue/gen"
	"openhue-cli/openhue/test/assert"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(NewTestIOStreamsDiscard(), NewTestBuildInfo(), &gen.ClientWithResponses{})
	assert.NotNil(t, ctx)
}

func TestNewTestContextWithoutApi(t *testing.T) {
	ctx, out := NewTestContextWithoutApi()
	assert.NotNil(t, ctx)
	assert.NotNil(t, out)
}
