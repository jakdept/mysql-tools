package main

var db string

// Process takes lines in via a reader, and processes them to write them out to a file
// It removes from start (byte slice) to a end (byte slice).
//
// Example usage to remove go comments from source code:
//     reader := Sanitize("/*", "*/", r)
// func Process(eol [][]byte, r io.Reader) {
// 	// create a scanner that breaks on end
// 	s := bufio.NewScanner(r)
// 	s.Split(SplitBytesSet(eol))

// 	for s.Scan() {
// 		line := bytes.TrimSpace(s.Bytes())
// 		switch {

// 		}

// 	}
// }
