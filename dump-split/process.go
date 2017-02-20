package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"unicode"
)

var db string

// Process takes lines in via a reader, and processes them to split inserts from
// creates from drops. eol allows you to specify multiple line terminating
//  characters - any of those will signify a line.
func Process(eol [][]byte, r io.Reader) {
	// create a scanner that breaks on end
	s := bufio.NewScanner(r)
	s.Split(SplitBytesSet(eol))

	for s.Scan() {
		line := bytes.TrimSpace(s.Bytes())
		switch {

		}
	}
}

func ParseInsert(line []byte) error {
	line = TrimPrefix(line, []byte("insert"))
	line = TrimPrefix(line, []byte("delayed"))
	line = TrimPrefix(line, []byte("low_priority"))
	line = TrimPrefix(line, []byte("high_priority"))
	line = TrimPrefix(line, []byte("ignore"))
	line = TrimPrefix(line, []byte("into"))

	// get the tableName/DB name
	tableLoc, line := SplitFieldFuncN(line, unicode.IsSpace, 1)
	tableChunks := bytes.Split(tableLoc[0], []byte("."))

	var db, table, partition, colList []byte

	if len(tableChunks) > 0 {
		db = bytes.Trim(tableChunks[0], "`")
		table = bytes.Trim(tableChunks[1], "`")
	} else {
		table = bytes.Trim(tableChunks[0], "`")
	}

	if HasPrefix(line, []byte("partition")) {
		line = TrimPrefix(line, []byte("partition"))
		var partitionChunks [][]byte
		partitionChunks, line = SplitFieldFuncN(line, unicode.IsSpace, 1)
		partition = bytes.Trim(bytes.TrimSpace(partitionChunks[0]), "`")
	}

	if line[0] == '(' {
		i := bytes.IndexRune(line, ')') + 1
		if len(line) <= i {
			return errors.New("Undefined error") // todo
			// kick an error
		}
		colList = bytes.TrimSpace(line[:i])
		line = bytes.TrimSpace(line[i:])
	}

	line = TrimPrefix(line, []byte("values"))
	line = TrimPrefix(line, []byte("value"))

	chunks := bytes.Split(line, []byte(","))

	// need to write this yet
	file, err := GetFile(string(db), string(table))
	if err != nil {
		return err
	}

	for _, row := range chunks {
		err := file.WriteRow(db, table, partition, colList, row)
		if err != nil {
			return err
		}
	}
	return nil
}
