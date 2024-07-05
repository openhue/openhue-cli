package openhue

import (
	"github.com/openhue/openhue-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext(NewTestIOStreamsDiscard(), NewTestBuildInfo(), &openhue.Home{}, nil)
	assert.NotNil(t, ctx, "Context should not be nil")
}

func TestNewTestContextWithoutApi(t *testing.T) {
	ctx, out := NewTestContext(nil)
	assert.NotNil(t, ctx, "Context should not be nil")
	assert.NotNil(t, out, "Out buffer should not be nil")
}
