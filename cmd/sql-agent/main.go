package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/chop-dbhi/sql-agent"
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

	http.HandleFunc("/", handlerRequest)

	log.Fatal(http.ListenAndServe(addr, nil))
}

type Payload struct {
	Driver     string
	Connection map[string]interface{}
	SQL        string
	Params     map[string]interface{}
}

func handlerRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

	db, err := sqlagent.Connect(payload.Driver, payload.Connection)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf("problem connecting to database: %s", err)))
		return
	}

	records, err := sqlagent.Execute(db, payload.SQL, payload.Params)

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(fmt.Sprintf("error executing query: %s", err)))
		return
	}

	if err = json.NewEncoder(w).Encode(records); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error encoding records: %s", err)))
		return
	}
}
