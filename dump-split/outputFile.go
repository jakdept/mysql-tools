package main

import (
	"bytes"
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
	DropTableSafe
	CreateTable
	CreateTableSafe
	InsertOverwrite
	InsertIgnore
	StaticDatabase
)

var semiColon = []byte(";")
var objectQuote = []byte("`")
var objectSep = []byte(".")
var emptySlice = []byte("")
var space = []byte(" ")
var openFiles openFileList
var filePermissions os.FileMode
var maxOpenFiles = 256
var globalOptions = 0
var globalPath = "./"

type OutputFile struct {
	data       *os.File
	create     *os.File
	db         string
	table      string
	options    int
	lastAccess time.Time
}

func GetFile(dbName, tableName string) (OutputFile, error) {
	for _, each := range openFiles {
		if each.db == dbName && each.table == tableName {
			return each, nil
		}
	}

	if err := pruneFiles(); err != nil {
		return OutputFile{}, err
	}

	var filename string
	if globalOptions&PerTable != 0 {
		filename = filepath.Join(globalPath, ".", dbName, ".", tableName, ".sql")
	} else {
		filename = filepath.Join(globalPath, ".", dbName, ".sql")
	}

	return openFile(dbName, tableName, filename, globalOptions)
}

func openFile(dbName, tableName, path string, options int) (OutputFile, error) {
	// prune a file if needed to make sure we've got room for this one

	var newFile OutputFile
	var err error
	newFile.data, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, filePermissions)
	if err != nil {
		return OutputFile{}, fmt.Errorf("problem opening the data output file: %v", err)
	}
	if options&PerTable != 0 && options&CreateTable != 0 {
		newFile.create, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, filePermissions)
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

func (f *OutputFile) WriteRow(db, table, partition, colList, data []byte) error {
	var chunks [][]byte
	chunks = append(chunks, []byte("INSERT"))
	if f.options&InsertIgnore != 0 {
		chunks = append(chunks, []byte("IGNORE"))
	}
	chunks = append(chunks, []byte("INTO"))

	tableSpec := [][]byte{objectQuote, table, objectQuote}
	if f.options&StaticDatabase != 0 {
		tableSpec = append([][]byte{objectQuote, db, objectQuote, objectSep}, tableSpec...)
	}
	chunks = append(chunks, bytes.Join(tableSpec, emptySlice))

	if len(partition) > 0 {
		chunks = append(chunks, []byte("PARTITION"))
		chunks = append(chunks, bytes.Join([][]byte{objectQuote, partition, objectQuote}, emptySlice))
	}

	if len(colList) > 0 {
		chunks = append(chunks, colList, []byte("VALUES"))
	}

	chunks = append(chunks, data, semiColon)

	_, err := f.data.Write(bytes.Join(chunks, space))
	return err
}

func (f *OutputFile) WriteDropTable(db, table []byte, tempoary bool) error {
	if f.options&DropTable == 0 {
		// if you ended up here and are not writing these, abandon
		return nil
	}

	var chunks [][]byte
	chunks = append(chunks, []byte("DROP"))
	if tempoary {
		chunks = append(chunks, []byte("TEMPOARY"))
	}
	chunks = append(chunks, []byte("TABLE"))

	if f.options&DropTableSafe != 0 {
		chunks = append(chunks, []byte("IF NOT EXISTS"))
	}

	tableSpec := [][]byte{objectQuote, table, objectQuote}
	if f.options&StaticDatabase != 0 {
		tableSpec = append([][]byte{objectQuote, db, objectQuote, objectSep}, tableSpec...)
	}

	chunks = append(chunks, bytes.Join(tableSpec, emptySlice))

	_, err := f.create.Write(bytes.Join(chunks, space))
	return err
}

func (f *OutputFile) WriteCloneTable(db, table, srcDb, srcTable []byte, tempoary bool) error {
	if f.options&DropTable == 0 {
		// if you ended up here and are not writing these, abandon
		return nil
	}

	var chunks [][]byte
	chunks = append(chunks, []byte("CREATE"))
	if tempoary {
		chunks = append(chunks, []byte("TEMPOARY"))
	}
	chunks = append(chunks, []byte("TABLE"))

	if f.options&CreateTableSafe != 0 {
		chunks = append(chunks, []byte("IF NOT EXISTS"))
	}

	tableSpec := [][]byte{objectQuote, table, objectQuote}
	if f.options&StaticDatabase != 0 {
		tableSpec = append([][]byte{objectQuote, db, objectQuote, objectSep}, tableSpec...)
	}
	chunks = append(chunks, bytes.Join(tableSpec, emptySlice))

	chunks = append(chunks, []byte("LIKE"))

	tableSpec = [][]byte{objectQuote, srcTable, objectQuote}
	if f.options&StaticDatabase != 0 {
		tableSpec = append([][]byte{objectQuote, srcDb, objectQuote, objectSep}, tableSpec...)
	}
	chunks = append(chunks, bytes.Join(tableSpec, emptySlice))

	_, err := f.create.Write(bytes.Join(chunks, space))
	return err
}
