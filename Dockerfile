FROM golang

ENV LD_LIBRARY_PATH /go/src/github.com/chop-dbhi/sql-agent/lib/oracle/instantclient_12_1
ENV ORACLE_HOME /go/src/github.com/chop-dbhi/sql-agent/lib/oracle/instantclient_12_1

RUN apt-get update -qq
RUN apt-get install libaio1 pkg-config -y

RUN mkdir -p /go/src/github.com/chop-dbhi/sql-agent

WORKDIR /go/src/github.com/chop-dbhi/sql-agent
ADD . /go/src/github.com/chop-dbhi/sql-agent
RUN cp ./lib/oracle/oci8.pc /usr/lib/pkgconfig/

WORKDIR ./lib/oracle
RUN tar zxf instantclient_12_1.tar.gz

WORKDIR instantclient_12_1
RUN ln -s libclntsh.so.12.1 libclntsh.so
RUN ln -s libocci.so.12.1 libocci.so

WORKDIR /go/src/github.com/chop-dbhi/sql-agent
RUN make cmd-install
RUN make build

CMD ["/go/bin/sql-agent", "-host", "0.0.0.0"]
