package main

import (
	"bufio"
	"bytes"
	"io"
)

func removeToEOL(token []byte, r io.Reader) bytes.Buffer {
	var buf bytes.Buffer
	s := bufio.NewScanner(r)
	go func() {
		for s.Scan() {
			if _, err := buf.Write(bytes.TrimSpace(bytes.Split(s.Bytes(), token)[0])); err != nil {
				panic(err)
			}
		}
	}()
	return buf
}

var bufferSize = 1024 * 1024 * 16

func removeToToken(start, end []byte, r io.Reader) bytes.Buffer {
	buf := make([]byte, 0, bufferSize)
	var readTail []byte
	var out bytes.Buffer
	var discard bool
	var n int
	var err error

	go func() {
		for {

			n, err = r.Read(buf)
		}
	}()
	return output
}
