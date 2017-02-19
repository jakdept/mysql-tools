package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	PerTable = 1 << iota
	DropDb
	CreateDb
	DropTable
	CreateTable
	InsertOverwrite
	InsertIgnore
)

var openFiles openFileList
var filePermissions os.FileMode
var maxOpenFiles = 256

type OutputFile struct {
	data       *os.File
	create     *os.File
	db         string
	table      string
	options    int
	lastAccess time.Time
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

func OpenFile(dbName, tableName, path string, options int) (OutputFile, error) {
	// prune a file if needed to make sure we've got room for this one
	if err := pruneFiles(); err != nil {
		return OutputFile{}, err
	}

	var filename string
	if options&PerTable != 0 {
		filename = filepath.Join(path, ".", dbName, ".", tableName, ".sql")
	} else {
		filename = filepath.Join(path, ".", dbName, ".sql")
	}

	var newFile OutputFile

	var err error
	newFile.data, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, filePermissions)
	if err != nil {
		return OutputFile{}, fmt.Errorf("problem opening the data output file: %v", err)
	}
	if options&PerTable != 0 && options&CreateTable != 0 {
		newFile.create, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, filePermissions)
		if err != nil {
			newFile.data.Close()
			return OutputFile{}, fmt.Errorf("problem opening the table info file: %v", err)
		}
	}

	// set the other stuff
	newFile.db = dbName
	newFile.table = tableName
	newFile.options = options
	newFile.lastAccess = time.Now()

	if err := newFile.startingLines(); err != nil {
		newFile.Close()
		return OutputFile{}, err
	}

	openFiles = append(openFiles, newFile)
	return newFile, nil
}

func pruneFiles() error {
	if len(openFiles) < maxOpenFiles {
		return nil
	}
	sort.Sort(openFiles)
	return openFiles[len(openFiles)-1].Close()
}

func (f *OutputFile) startingLines() error {
	padLines := []string{
		"/**Disable Keys **/",
	}
	var err error
	for _, line := range padLines {
		_, err = f.data.Write([]byte(line))
		if err != nil {
			return fmt.Errorf("problem prepping output file: %v", err)
		}
	}
	return nil
}

func (f *OutputFile) endingLines() error {
	padLines := []string{
		"/**Disable Keys **/",
	}
	var err error
	for _, line := range padLines {
		_, err = f.data.Write([]byte(line))
		if err != nil {
			return fmt.Errorf("problem prepping output file: %v", err)
		}
	}
	return nil
}

func (f *OutputFile) Close() error {
	var myId int
	for i, file := range openFiles {
		if f.db == file.db && f.table == file.table {
			myId = i
		}
	}

	openFiles = append(openFiles[:myId], openFiles[myId+1:len(openFiles)-1]...)

	var errs []error
	errs = append(errs, f.endingLines())
	errs = append(errs, f.data.Close())
	if f.options&PerTable != 0 && f.options&CreateTable != 0 {
		errs = append(errs, f.create.Close())
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return fmt.Errorf("problem closing file: %v", errs[0])
	default:
		return fmt.Errorf(`
    problem closing table data: %v\n
    problem closing closing table schema: %v
    `, errs[0], errs[1])
	}
}
