package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/chop-dbhi/sql-agent"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
)

var usage = `SQL Agent - HTTP interface

This is an HTTP interface for the SQL Agent.

Run:

	sql-agent [-host=<host>] [-port=<port>]

Example:

	POST /
	Content-Type application/json

	{
		"driver": "postgres",
		"connection": {
			"host": "pghost.org",
			"port": 5432,
		},
		"sql": "SELECT * FROM users WHERE zipcode = :zipcode",
		"parameters": {
			"zipcode": 19104
		}
	}
`

const StatusUnprocessableEntity = 422

var (
	defaultMimetype = "application/json"

	mimetypeFormats = map[string]string{
		"*/*":                  "json",
		"text/csv":             "csv",
		"application/json":     "json",
		"application/x-ldjson": "ldjson",
	}
)

// parseMimetype parses a mimetype from the Accept header.
func parseMimetype(mimetype string) string {
	mimetype, params, err := mime.ParseMediaType(mimetype)

	if err != nil {
		return ""
	}

	// No Accept header passed.
	if mimetype == "" {
		return defaultMimetype
	}

	switch mimetype {
	case "application/json":
		if params["boundary"] == "NL" {
			return "application/x-ldjson"
		}
	default:
		if _, ok := mimetypeFormats[mimetype]; !ok {
			return ""
		}
	}

	return mimetype
}

func init() {
	flag.Usage = func() {
		fmt.Println(usage)
		flag.PrintDefaults()
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var (
		host string
		port int
	)

	flag.StringVar(&host, "host", "localhost", "Host of the agent.")
	flag.IntVar(&port, "port", 5000, "Port of the agent.")

	flag.Parse()

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("* Listening on %s...\n", addr)

	http.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(addr, nil)
	sqlagent.Shutdown()
	log.Fatal(err)
}

type Payload struct {
	Driver     string
	Connection map[string]interface{}
	SQL        string
	Params     map[string]interface{}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Validate the Accept header and parse it to ensure it is
	// supported.
	mimetype := r.Header.Get("Accept")

	if mimetype = parseMimetype(mimetype); mimetype == "" {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	var payload Payload

	// Decode the body.
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("could not decode JSON: %s", err)))
		return
	}

	if _, ok := sqlagent.Drivers[payload.Driver]; !ok {
		w.WriteHeader(StatusUnprocessableEntity)
		w.Write([]byte(fmt.Sprintf("unknown driver: %v", payload.Driver)))
		return
	}

	db, err := sqlagent.PersistentConnect(payload.Driver, payload.Connection)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf("problem connecting to database: %s", err)))
		return
	}

	iter, err := sqlagent.Execute(db, payload.SQL, payload.Params)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf("error executing query: %s", err)))
		return
	}

	defer iter.Close()

	w.Header().Set("content-type", mimetype)

	switch mimetypeFormats[mimetype] {
	case "csv":
		err = sqlagent.EncodeCSV(w, iter)
	case "json":
		err = sqlagent.EncodeJSON(w, iter)
	case "ldjson":
		err = sqlagent.EncodeLDJSON(w, iter)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error encoding data: %s", err)))
		return
	}
}
