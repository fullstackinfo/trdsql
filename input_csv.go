package trdsql

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
)

// CSVReader provides methods of the Reader interface.
type CSVReader struct {
	reader  *csv.Reader
	names   []string
	types   []string
	preRead [][]string
}

// NewCSVReader returns CSVReader and error.
func NewCSVReader(reader io.Reader, opts *ReadOpts) (*CSVReader, error) {
	var err error
	r := &CSVReader{}
	if reader == nil {
		return nil, errors.New("nil reader")
	}
	r.reader = csv.NewReader(reader)
	r.reader.LazyQuotes = true
	r.reader.FieldsPerRecord = -1 // no check count
	r.reader.TrimLeadingSpace = true
	r.reader.Comma, err = delimiter(opts.InDelimiter)

	if opts.InSkip > 0 {
		skip := make([]interface{}, 1)
		for i := 0; i < opts.InSkip; i++ {
			row, err := r.ReadRow(skip)
			if err != nil {
				log.Printf("ERROR: skip error %s", err)
				break
			}
			debug.Printf("Skip row:%s\n", row)
		}
	}

	// Header
	if opts.InHeader {
		row, err := r.reader.Read()
		if err != nil && err != io.EOF {
			return nil, err
		}
		r.names = make([]string, len(row))
		for i, col := range row {
			if col == "" {
				r.names[i] = "c" + strconv.Itoa(i+1)
			} else {
				r.names[i] = col
			}
		}
		opts.InPreRead--
	}

	for n := 0; n < opts.InPreRead; n++ {
		row, err := r.reader.Read()
		if err != nil {
			if err != io.EOF {
				return r, err
			}
			return r, nil
		}
		rows := make([]string, len(row))
		for i, col := range row {
			rows[i] = col
			if len(r.names) < i+1 {
				r.names = append(r.names, "c"+strconv.Itoa(i+1))
			}
		}
		r.preRead = append(r.preRead, rows)
	}

	return r, err
}

func delimiter(sepString string) (rune, error) {
	if sepString == "" {
		return 0, nil
	}
	sepRunes, err := strconv.Unquote(`'` + sepString + `'`)
	if err != nil {
		return ',', fmt.Errorf("can not get separator: %w:\"%s\"", err, sepString)
	}
	sepRune := ([]rune(sepRunes))[0]
	return sepRune, err
}

// Names returns column names.
func (r *CSVReader) Names() ([]string, error) {
	if len(r.names) == 0 {
		return r.names, fmt.Errorf("no rows")
	}
	return r.names, nil
}

// Types returns column types.
// All CSV types return the DefaultDBType.
func (r *CSVReader) Types() ([]string, error) {
	r.types = make([]string, len(r.names))
	for i := 0; i < len(r.names); i++ {
		r.types[i] = DefaultDBType
	}
	return r.types, nil
}

// PreReadRow is returns only columns that store preread rows.
func (r *CSVReader) PreReadRow() [][]interface{} {
	rowNum := len(r.preRead)
	rows := make([][]interface{}, rowNum)
	for n := 0; n < rowNum; n++ {
		rows[n] = make([]interface{}, len(r.names))
		for i, f := range r.preRead[n] {
			rows[n][i] = f
		}
	}
	return rows
}

// ReadRow is read the rest of the row.
func (r *CSVReader) ReadRow(row []interface{}) ([]interface{}, error) {
	record, err := r.reader.Read()
	if err != nil {
		return row, err
	}
	for i := 0; len(row) > i; i++ {
		if len(record) > i {
			row[i] = record[i]
		} else {
			row[i] = nil
		}
	}
	return row, nil
}
