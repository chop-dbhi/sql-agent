#!/bin/bash

export ROOT=/go/src/github.com/chop-dbhi/sql-agent
export LD_LIBRARY_PATH=$ROOT/lib/oracle/instantclient_12_1
export ORACLE_HOME=$ROOT/lib/oracle/instantclient_12_1
export CGO_ENABLED=1

cd /go/src/github.com/chop-dbhi/sql-agent
cp ./lib/oracle/oci8.pc /usr/lib/pkgconfig/

cp ./lib/unixODBC-2.3.1.tar.gz /opt
cp -r ./lib/netezza /opt/netezza

cd /opt
tar xf unixODBC-2.3.1.tar.gz
cd /opt/unixODBC-2.3.1
./configure --disable-gui
make
make install
echo '/usr/local/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
ldconfig
rm -rf /opt/unixODBC-2.3.1*

# Install Netezza ODBC driver
cd /opt/netezza
./unpack -f /usr/local/nz
odbcinst -i -d -f /opt/netezza/netezza.driver
echo '/usr/local/nz/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
echo '/usr/local/nz/lib64' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
ldconfig
rm -rf /opt/netezza
ln -s /usr/local/nz/bin64/nzodbcsql /usr/local/bin/nzodbcsql

mkdir -p $ROOT/dist/linux-amd64
mkdir -p $ROOT/cmd/sql-agent/vendor/github.com/chop-dbhi
rm -rf $ROOT/cmd/sql-agent/vendor/github.com/chop-dbhi/sql-agent
ln -s $ROOT $ROOT/cmd/sql-agent/vendor/github.com/chop-dbhi/sql-agent
cd $ROOT/cmd/sql-agent

go build -v -o $ROOT/dist/linux-amd64/sql-agent
