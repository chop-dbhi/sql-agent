package sqlagent

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
)

// Encoder provides an satisfies the encoder type.
type Encoder func(io.Writer, *Iterator) error

func EncodeCSV(w io.Writer, i *Iterator) error {
	r := make([]interface{}, len(i.Cols), len(i.Cols))
	o := make([]string, len(i.Cols), len(i.Cols))

	// Allocate string pointers.
	for i := range r {
		r[i] = &sql.NullString{}
	}

	enc := csv.NewWriter(w)

	if err := enc.Write(i.Cols); err != nil {
		return err
	}

	for i.Next() {
		if err := i.ScanRow(r); err != nil {
			return err
		}

		for i, v := range r {
			x := v.(*sql.NullString)
			if x.Valid {
				o[i] = x.String
			} else {
				o[i] = ""
			}
		}

		if err := enc.Write(o); err != nil {
			return err
		}
	}

	enc.Flush()
	return enc.Error()
}

// EncodeJSON encodes the iterator as a JSON array of records.
func EncodeJSON(w io.Writer, i *Iterator) error {
	r := make(Record)

	// Open paren.
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}

	var c int
	enc := json.NewEncoder(w)

	delim := []byte{',', '\n'}

	for i.Next() {
		if c > 0 {
			if _, err := w.Write(delim); err != nil {
				return err
			}
		}

		c++

		if err := i.Scan(r); err != nil {
			return err
		}

		if err := enc.Encode(r); err != nil {
			return err
		}
	}

	// Close paren.
	if _, err := w.Write([]byte{']'}); err != nil {
		return err
	}

	return nil
}

// EncodeLDJSON encodes the iterator as a line delimited stream
// of records.
func EncodeLDJSON(w io.Writer, i *Iterator) error {
	r := make(Record)

	enc := json.NewEncoder(w)

	for i.Next() {
		if err := i.Scan(r); err != nil {
			return err
		}

		if err := enc.Encode(r); err != nil {
			return err
		}
	}

	return nil
}
