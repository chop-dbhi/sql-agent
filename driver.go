package sqlagent

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Connector takes a map of connection parameters and converts them into a
// connection string. In order to support a uniform way of specifying connection
// options, a connector must be used to conver the map of options to a
// connection the underlying driver supports.
type connector func(map[string]interface{}) string

// cleanParams removes any key with empty string values.
func cleanParams(p map[string]interface{}) map[string]interface{} {
	c := make(map[string]interface{})

	for k, v := range p {
		switch x := v.(type) {
		case string:
			if x == "" {
				continue
			}
		}

		c[k] = v
	}

	return c
}

// Set of connectors for each internal driver.
var connectors = map[string]connector{
	// Postgres supports space-delimited key=value pairs.
	// See http://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
	"postgres": func(params map[string]interface{}) string {
		toks := make([]string, len(params))
		i := 0

		for k, v := range params {
			toks[i] = fmt.Sprintf("%s=%v", k, v)
			i++
		}

		return strings.Join(toks, " ")
	},

	// MySQL has a more complex format.
	// See: https://github.com/go-sql-driver/mysql/#dsn-data-source-name
	"mysql": func(params map[string]interface{}) string {
		var (
			user, pass, db interface{}

			host interface{} = "localhost"
			port interface{} = 3306

			query []string
		)

		for k, v := range params {
			switch k {
			case "user":
				user = v
			case "password":
				pass = v
			case "host":
				host = v
			case "port":
				port = v
			case "database":
				db = v
			default:
				query = append(query, fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
			}
		}

		var conn string

		if user != nil {
			conn += fmt.Sprintf("%s", user)
		}

		if pass != nil {
			conn += fmt.Sprintf(":%s", pass)
		}

		if conn != "" {
			conn += "@"
		}

		// Only TCP is supported.
		conn += "tcp"

		// If a host is supplied a port must as well or vice versa.
		if host != nil || port != nil {
			conn += fmt.Sprintf("(%s:%v)", host, port)
		}

		conn += fmt.Sprintf("/%v", db)

		if len(query) > 0 {
			conn += fmt.Sprintf("?%s", strings.Join(query, "&"))
		}

		return conn
	},

	// SQLite3 requires the path and supports other query parameters.
	// See http://godoc.org/github.com/mattn/go-sqlite3#SQLiteDriver.Open
	"sqlite3": func(params map[string]interface{}) string {
		var (
			db    interface{} = ":memory:"
			query []string
		)

		for k, v := range params {
			if k == "database" {
				db = v
			} else {
				query = append(query, fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
			}
		}

		if len(query) == 0 {
			return fmt.Sprint(db)
		}

		return fmt.Sprintf("%s?%s", db, strings.Join(query, "&"))
	},

	// MSSQL supports semicolon delimited key=value parameters.
	// See https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn
	"mssql": func(params map[string]interface{}) string {
		toks := make([]string, len(params))
		i := 0

		for k, v := range params {
			switch k {
			case "host":
				k = "server"
			case "user":
				k = "user id"
			}

			toks[i] = fmt.Sprintf("%s=%v", k, v)
			i++
		}

		return strings.Join(toks, ";")
	},

	// Oracle supports a standard URI-based connection string.
	// See http://godoc.org/github.com/mattn/go-oci8#ParseDSN
	"oci8": func(params map[string]interface{}) string {
		var (
			user, pass, db interface{}

			host interface{} = "localhost"
			port interface{} = 1521

			query []string
		)

		for k, v := range params {
			switch k {
			case "user":
				user = v
			case "password":
				pass = v
			case "host":
				host = v
			case "port":
				port = v
			case "database":
				db = v
			default:
				query = append(query, fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
			}
		}

		var conn string

		if user != nil {
			conn += fmt.Sprintf("%s", user)
		}

		if pass != nil {
			conn += fmt.Sprintf("/%s", pass)
		}

		if conn != "" {
			conn += "@"
		}

		// If a host is supplied a port must as well or vice versa.
		if host != nil {
			conn += fmt.Sprint(host)
		}

		if port != nil {
			conn += fmt.Sprintf(":%v", port)
		}

		conn += fmt.Sprintf("/%s", db)

		if len(query) > 0 {
			conn += fmt.Sprintf("?%s", strings.Join(query, "&"))
		}

		return conn
	},
}

// mapBytesToString ensures byte slices that were returned from the database
// are represented as strings.
// See https://github.com/jmoiron/sqlx/issues/135
func mapBytesToString(m map[string]interface{}) {
	for k, v := range m {
		if b, ok := v.([]byte); ok {
			m[k] = string(b)
		}
	}
}

// ErrUnknownDriver is returned when an unknown driver is used when attempting to connect.
var ErrUnknownDriver = errors.New("sqlagent: Unknown driver")

// Drivers contains a map of public driver names to registered driver names.
var Drivers = map[string]string{
	"postgresql": "postgres",
	"postgres":   "postgres",
	"mysql":      "mysql",
	"mariadb":    "mysql",
	"sqlite":     "sqlite3",
	"mssql":      "mssql",
	"sqlserver":  "mssql",
	"oracle":     "oci8",
}
