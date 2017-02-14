package sqlagent

import "github.com/jmoiron/sqlx"

// Record is a database row keyed by column name. This requires the columns to be
// uniquely named.
type Record map[string]interface{}

// Iterator provides a lazy access to the database rows.
type Iterator struct {
	Cols []string
	rows *sqlx.Rows
}

// Close closes the iterator.
func (i *Iterator) Close() {
	i.rows.Close()
}

// Next returns true if another row is available.
func (i *Iterator) Next() bool {
	return i.rows.Next()
}

// Scan takes a record and scans the values of a row into the record.
func (i *Iterator) Scan(r Record) error {
	if err := i.rows.MapScan(r); err != nil {
		return err
	}

	mapBytesToString(r)

	return nil
}

func (i *Iterator) ScanRow(r []interface{}) error {
	return i.rows.Scan(r...)
}

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
func Execute(db *sqlx.DB, sql string, params map[string]interface{}) (*Iterator, error) {
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

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	return &Iterator{
		Cols: cols,
		rows: rows,
	}, nil
}
