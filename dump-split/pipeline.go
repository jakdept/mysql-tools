package main

import (
	"bytes"
	"errors"
	"time"
)

type Pipeline struct {
	MaxSize int
	buf     bytes.Buffer
	closed  bool
}

func (p *Pipeline) IsClosed() bool {
	return p.closed
}

// Closes the Pipeline so it will not accept any further input
func (p *Pipeline) Close() {
	p.closed = true
}

func (p *Pipeline) Read(out []byte) (int, error) {
	// if the pipeline is not closed, return it
	if p.buf.Len() < 1 {
		return 0, nil
	}
	return p.buf.Read(out)
}

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
