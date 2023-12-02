package openhue

import (
	"io"
	"os"
	"testing"
)

func TestNewIOSteams(t *testing.T) {
	streams := NewIOStreams()
	if streams.In != os.Stdin {
		t.Fatalf("In should be os.Stdin")
	}
	if streams.Out != os.Stdout {
		t.Fatalf("Out should be os.Stdout")
	}
	if streams.ErrOut != os.Stderr {
		t.Fatalf("ErrOut should be os.Stderr")
	}
}

func TestNewTestIOStreams(t *testing.T) {

	streams, in, out, err := NewTestIOStreams()

	if streams.In != in {
		t.Fatalf("invalid in")
	}
	if streams.Out != out {
		t.Fatalf("invalid out")
	}
	if streams.ErrOut != err {
		t.Fatalf("invalid err")
	}
}

func TestNewTestIOStreamsDiscard(t *testing.T) {
	streams := NewTestIOStreamsDiscard()
	if streams.In == nil {
		t.Fatalf("In should be set")
	}
	if streams.Out != io.Discard {
		t.Fatalf("Out should be io.Discard")
	}
	if streams.ErrOut != io.Discard {
		t.Fatalf("ErrOut should be io.Discard")
	}
}
