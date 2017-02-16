package main

import (
	"bufio"
	"bytes"
)

func sanitizeInput(p Pipeline) Pipeline {
	var output Pipeline
	s := bufio.NewScanner(p)
	go func() {
		for s.Scan() {
			if _, err := output.Write(bytes.TrimSpace(bytes.Split(s.Bytes(), []byte("\\!"))[0])); err != nil {
				p.Error = err
				return
			}
		}
	}()
	return output
}
