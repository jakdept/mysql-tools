package main

import (
	"bufio"
	"bytes"
	"io"
	"sort"
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

type endTokenLocation struct {
	pos int
	len int
}

type endTokenLocations []endTokenLocation

func (t endTokenLocations) Len() int {
	return len(t)
}
func (t endTokenLocations) Less(i, j int) bool {
	return t[i].pos+t[i].len < t[j].pos+t[i].len
}
func (t endTokenLocations) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func SplitBytesSet(endTokens [][]byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		var endLocations endTokenLocations
		for _, token := range endTokens {
			if i := bytes.Index(data, token); i >= 0 {
				endLocations = append(endLocations, endTokenLocation{pos: i, len: len(token)})
			}
		}
		if len(endLocations) > 0 {
			sort.Sort(endLocations)
			return endLocations[0].pos + endLocations[0].len, data[0:endLocations[0].pos], nil

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
func Sanitize(start, end []byte, r io.Reader) bytes.Buffer {
	// create my output io.Reader
	var buf bytes.Buffer

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

// SanitizeToEOL removes stuff from an io.Reader, returning a bytes.Buffer.
// It removes from start (byte slice) to the end of that line.
//
// Example usage to remove go comments from source code:
//     reader := SanitizeToEOL("//")
func SanitizeToEOL(start []byte, r io.Reader) bytes.Buffer {
	// create my output io.Reader
	var buf bytes.Buffer

	// create a scanner that breaks on end
	s := bufio.NewScanner(r)

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
