package openhue

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
)

type Console interface {
	Printf(format string, a ...any)
	Println(a ...any)

	ErrPrintf(format string, a ...any)
	ErrPrintln(a ...any)
}

// IOStreams provides the standard names for io streams.
// This is useful for embedding and for unit testing.
type IOStreams struct {
	Console
	// In think, os.Stdin
	In io.Reader
	// Out think, os.Stdout
	Out io.Writer
	// ErrOut think, os.Stderr
	ErrOut io.Writer
}

// NewIOStreams returns a valid IOStreams with the default os.Stdin, os.Stdout and os.Stderr thinks
func NewIOStreams() IOStreams {
	return IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}
}

// NewTestIOStreams returns a valid IOStreams and in, out, errOut buffers for unit tests
func NewTestIOStreams() (IOStreams, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	in := &bytes.Buffer{}
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	return IOStreams{
		In:     in,
		Out:    out,
		ErrOut: errOut,
	}, in, out, errOut
}

// NewTestIOStreamsDiscard returns a valid IOStreams that just discards
func NewTestIOStreamsDiscard() IOStreams {
	in := &bytes.Buffer{}
	return IOStreams{
		In:     in,
		Out:    io.Discard,
		ErrOut: io.Discard,
	}
}

//
// Console implementation
//

func (io *IOStreams) Printf(format string, a ...any) {
	_, err := fmt.Fprintf(io.Out, format, a...)
	if err != nil {
		log.Error(err)
		return
	}
}

func (io *IOStreams) Println(a ...any) {
	_, err := fmt.Fprintln(io.Out, a...)
	if err != nil {
		log.Error(err)
		return
	}
}

func (io *IOStreams) ErrPrintf(format string, a ...any) {
	_, err := fmt.Fprintf(io.ErrOut, format, a...)
	if err != nil {
		log.Error(err)
		return
	}
}

func (io *IOStreams) ErrPrintln(a ...any) {
	_, err := fmt.Fprintln(io.ErrOut, a...)
	if err != nil {
		log.Error(err)
		return
	}
}
