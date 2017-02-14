package main

import (
	"bytes"
	"errors"
	"io"
	"time"
)

type Pipeline struct {
	MaxSize int
	buf     bytes.Buffer
	closed  bool
}

// IsClosed reports if the pipeline is closed.
func (p *Pipeline) IsClosed() bool {
	return p.closed
}

// Close changes the pipeline so that it's in a closed state.
// Further writes into the pipeline will not be accepted.
// The pipeline can still be read from.
func (p *Pipeline) Close() {
	p.closed = true
}

// Read will return some back.
// If there is nothing in the buffer, nothing is returned.
// Only if the Pipeline is closed and empty will it return io.EOF.
func (p *Pipeline) Read(out []byte) (int, error) {
	// if the pipeline is not closed, return it
	if p.closed && p.buf.Len() < 1 {
		return 0, io.EOF
	}
	if p.buf.Len() < 1 {
		return 0, nil
	}
	return p.buf.Read(out)
}

// Write allows a writing into the Pipeline.
// Writing into a closed pipeline will return an error.
// If the pipeline has a non-zero MaxSize (so a limit on it's size), the Write
// will be blocked until it will not push it over that size.
func (p *Pipeline) Write(in []byte) (int, error) {
	if p.closed {
		return 0, errors.New("cannot write to closed pipeline")
	}

	// block until there is space in the buffer to write
	for p.buf.Len()+len(in) > p.MaxSize && p.MaxSize > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	return p.buf.Write(in)
}
