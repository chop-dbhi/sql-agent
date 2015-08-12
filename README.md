# SQL Agent

[![GoDoc](https://godoc.org/github.com/chop-dbhi/sql-agent?status.svg)](https://godoc.org/github.com/chop-dbhi/sql-agent)

SQL Agent is an HTTP service for executing ad-hoc queries on remote databases. The motivation for this service is to be part of a data monitoring process or system in which the query results will be evaluated against previous snapshots of the results.

The supported databases are:

- PostgreSQL
- MySQL, MariaDB
- Oracle
- Microsoft SQL Server
- SQLite

In addition to the service, this repo also defines a `sqlagent` package for using in other Go programs.

## Install

At the moment, it is recommended to run the service using Docker because there are no pre-built binaries yet.

```
docker run -d -p 5000:5000 dbhi/sql-agent
```

## Usage

To execute a query, simply send a POST request with a payload containing the driver name of the database, connection information to with, and the SQL statement with optional parameters. The service will connect to the database, execute the query and return the results as a JSON-encoded array of maps (see [details](#Details) below).

**Request**

```json
{
    "driver": "postgres",
    "connection": {
        "host": "localhost",        
        "user": "postgres"
    },
    "sql": "SELECT name FROM users WHERE zipcode = :zipcode",
    "params": {
        "zipcode": 18019
    }
}
```

**Response**

```json
[
    {
        "name": "George"
    },
    ...
]
```

### Connection Options

The core option names are standardized for ease of use.

- `host` - The host of the database.
- `port` - The port of the database.
- `user` - The user to connect with.
- `password` - The password to authenticate with.
- `database` - The name of the database to connect to. For SQLite, this will be a filesystem path. For Oracle, this would be the SID.

Other options that are supplied are passed query options if they are known, otherwise they are they ignored.

## Details

- Only `SELECT` statements are supported.
- Statements using parameters must use the `:param` syntax and must have a corresponding entry in the `params` map.
- The only standard 

### Constraints

- Columns must be uniquely named, otherwise the conversion into a map will include only one of the values.
