package sqlagent

import (
	"encoding/json"
	"io"
	"net/http"
)

// Encoder provides an interface for encoding an iterator into
// various formats.
type Encoder struct {
	w io.Writer
}

// EncodeJSON encodes the iterator as a JSON array of records.
func (e *Encoder) EncodeJSON(i *Iterator) error {
	var (
		err error
		r   Record
		rs  []Record
	)

	for i.Next() {
		// Initial new record for each
		r = make(Record)

		if err = i.Scan(r); err != nil {
			return err
		}

		rs = append(rs, r)
	}

	return json.NewEncoder(e.w).Encode(rs)
}

// EncodeLDJSON encodes the iterator as a line delimited stream
// of records.
func (e *Encoder) EncodeLDJSON(i *Iterator) error {
	var (
		err error
		r   = make(Record)
		n   int
	)

	fw, ok := e.w.(http.Flusher)

	enc := json.NewEncoder(e.w)

	for i.Next() {
		if err = i.Scan(r); err != nil {
			return err
		}

		if err = enc.Encode(r); err != nil {
			return err
		}

		// Flusher
		if ok {
			n++

			if n%1000 == 0 {
				fw.Flush()
			}
		}
	}

	return nil
}

// NewEncoder returns an Encoder bound to the io.Writer.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}
