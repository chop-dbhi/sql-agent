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

## Development

### Setup

- [Install Oracle clients libraries](#install-oracle-client-libraries)
- Run `make test-install`

#### Install Oracle client libraries

##### General

In order to install the [go-oci8](https://github.com/mattn/go-oci8) driver, you must install Oracle's client libraries.

Download the **instantclient-basic** and **instantclient-sdk** package from [Oracle's website](http://www.oracle.com/technetwork/database/features/instant-client/index-097480.html) and uncompress to the same directory. Make sure that you selected the platform and architecture.

The installations instructions are listed at the bottom of the page with the download links.

Install `pkg-config`.

Create `oci8.pc` file in your `$PKG_CONFIG_PATH` (such as `/usr/local/lib/pkgconfig`) and add the below contents:

```
prefix=/usr/local/lib/instantclient_11_2
libdir=${prefix}
includedir=${prefix}/sdk/include/

Name: OCI
Description: Oracle database engine
Version: 11.2
Libs: -L${libdir} -lclntsh
Libs.private:
Cflags: -I${includedir}
```

Change the `prefix` to path to location of the Oracle libraries.

##### OS X specific Help

Assuming the `instantclient_11_2` folder is located in `/usr/loca/lib`, link the following files:

```bash
ln /usr/local/lib/instantclient_11_2/libclntsh.dylib /usr/local/lib/libclntsh.dylib
ln /usr/local/lib/instantclient_11_2/libocci.dylib.* /usr/local/lib/libocci.dylib.*
ln /usr/local/lib/instantclient_11_2/libociei.dylib /usr/local/lib/libociei.dylib
ln /usr/local/lib/instantclient_11_2/libnnz11.dylib /usr/local/lib/libnnz11.dylib
```

Install `pkg-config` via [Homebrew](http://brew.sh/), `brew install pkg-config`.
