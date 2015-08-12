package sqlagent

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-oci8"
	_ "github.com/mattn/go-sqlite3"
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
	// See https://github.com/denisenkom/go-mssqldb#connection-parameters
	"mssql": func(params map[string]interface{}) string {
		toks := make([]string, len(params))
		i := 0

		for k, v := range params {
			toks[i] = fmt.Sprintf("%s=%v", k, v)
			i++
		}

		return strings.Join(toks, " ")
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
	"postgres":  "postgres",
	"mysql":     "mysql",
	"mariadb":   "mysql",
	"sqlite":    "sqlite3",
	"mssql":     "mssql",
	"sqlserver": "mssql",
	"oracle":    "oci8",
}

// Record is a database row keyed by column name. This requires the columns to be
// uniquely named.
type Record map[string]interface{}

// Connect connects to a database given a driver name and set of connection parameters.
// Each database supports a different set of connection parameters, however the few
// that are common are standardized.
//
// - `host` - The database host.
// - `port` - The database port.
// - `user` - The username to authenticate with.
// - `password` - The password to authenticate with.
// - `database` - The database to connect to.
//
// Other known database-specific parameters will be appended to the connection string and the remaining will be ignored.
func Connect(driver string, params map[string]interface{}) (*sqlx.DB, error) {
	// Select the driver.
	driver, ok := Drivers[driver]

	if !ok {
		return nil, ErrUnknownDriver
	}

	// Connect to the database.
	connector := connectors[driver]

	params = cleanParams(params)
	connstr := connector(params)

	return sqlx.Connect(driver, connstr)
}

// Execute takes a database instance, SQL statement, and parameters and executes the query
// returning the resulting rows.
func Execute(db *sqlx.DB, sql string, params map[string]interface{}) ([]Record, error) {
	var (
		err  error
		rows *sqlx.Rows
	)

	// Execute the query.
	if params != nil && len(params) > 0 {
		rows, err = db.NamedQuery(sql, params)
	} else {
		rows, err = db.Queryx(sql)
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var (
		record  Record
		records []Record
	)

	for rows.Next() {
		record = make(Record)

		if err = rows.MapScan(record); err != nil {
			break
		}

		mapBytesToString(record)

		records = append(records, record)
	}

	if err != nil {
		return nil, err
	}

	return records, nil
}
