package main

import (
	"bytes"
	"unicode"
)

func HasPrefix(s, prefix []byte) bool {
	return len(s) >= len(prefix) &&
		bytes.Equal(bytes.ToLower(s[0:len(prefix)]), bytes.ToLower(prefix))
}

func TrimPrefix(s, prefix []byte) []byte {
	if HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

type openFileList []OutputFile

// Len is the number of elements in the collection.
func (f openFileList) Len() int { return len(f) }

// Swap swaps the elements with indexes i and j.
func (f openFileList) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less reports whether the element with
// index i should sort before the element with index j.
func (f openFileList) Less(i, j int) bool {
	return f[i].lastAccess.Before(f[j].lastAccess)
}

// SplitFieldFuncN returns the first n fields of a given s []byte - fields as
//  defined by the second parameter. If not enough fields are fount, the ones
//  found are returned.
func SplitFieldFuncN(s []byte, f func(r rune) bool, n int) (fields [][]byte, remaining []byte) {
	var i int
	remaining = s
	for len(fields) < n {
		i = bytes.IndexFunc(s, unicode.IsSpace)
		if i == -1 {
			return
		}
		fields = append(fields, bytes.TrimSpace(remaining[:i]))
		remaining = bytes.TrimSpace(remaining[i:])
	}
	return
}
