package openhue

import (
	"github.com/stretchr/testify/assert"
	"openhue-cli/openhue/gen"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(NewTestIOStreamsDiscard(), NewTestBuildInfo(), &gen.ClientWithResponses{})
	assert.NotNil(t, ctx, "Context should not be nil")
}

func TestNewTestContextWithoutApi(t *testing.T) {
	ctx, out := NewTestContextWithoutApi()
	assert.NotNil(t, ctx, "Context should not be nil")
	assert.NotNil(t, out, "Out buffer should not be nil")
}
