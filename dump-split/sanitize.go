package main

import (
	"bufio"
	"bytes"
	"io"
)

// SplitBytes satasifies bufio.SplitFunc while allowing splitting on a []byte
// slice that is specified at calling time. It mostly returns the number of
// characers to burn, the next token for the scanner, and any relevant error.
func SplitBytes(endToken []byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.Index(data, endToken); i >= 0 {
			return i + len(endToken), data[0:i], nil
		}
		// If we're at EOF, we have a final, non-terminated line. Return it.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
}

// Sanitize removes stuff from an io.Reader, returning a bytes.Buffer.
// It removes from start (byte slice) to a end (byte slice).
//
// Example usage to remove go comments from source code:
//     reader := Sanitize("/*", "*/", r)
//     reader := Sanitize("//", "*\n, r)
func Sanitize(start, end []byte, r io.Reader) bytes.Buffer {
	// create my output io.Reader
	var buf bytes.Buffer
	// var discard bool

	// create a scanner that breaks on end
	s := bufio.NewScanner(r)
	s.Split(SplitBytes(end))

	go func() {
		for s.Scan() {
			// somewhere in here, I need to look for start - and if i find it, start discarding instead
			if _, err := buf.Write(bytes.TrimSpace(bytes.Split(s.Bytes(), start)[0])); err != nil {
				panic(err)
			}
		}
	}()
	return buf
}

var bufferSize = 1024 * 1024 * 16
