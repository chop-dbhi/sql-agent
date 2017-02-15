#!/bin/bash

export ROOT=/go/src/github.com/chop-dbhi/sql-agent
export LD_LIBRARY_PATH=$ROOT/lib/oracle/instantclient_12_1
export ORACLE_HOME=$ROOT/lib/oracle/instantclient_12_1
export CGO_ENABLED=1

cd /go/src/github.com/chop-dbhi/sql-agent
cp ./lib/oracle/oci8.pc /usr/lib/pkgconfig/

mkdir -p $ROOT/dist/linux-amd64
cd $ROOT/cmd/sql-agent
go build -v -o $ROOT/dist/linux-amd64/sql-agent
