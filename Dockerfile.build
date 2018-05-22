FROM golang:1.10

ENV LD_LIBRARY_PATH /usr/lib/instantclient_12_1
ENV ORACLE_HOME /usr/lib/instantclient_12_1

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
      libaio1 \
      pkg-config

COPY ./lib/oracle/instantclient_12_1 /usr/lib/instantclient_12_1
COPY ./lib/oracle/oci8.pc /usr/lib/pkgconfig/
RUN ln -s /usr/lib/instantclient_12_1/libclntsh.so.12.1 /usr/lib/instantclient_12_1/libclntsh.so
RUN ln -s /usr/lib/instantclient_12_1/libocci.so.12.1 /usr/lib/instantclient_12_1/libocci.so

COPY ./lib/unixODBC-2.3.1.tar.gz /opt/
WORKDIR /opt
RUN tar xf unixODBC-2.3.1.tar.gz
WORKDIR /opt/unixODBC-2.3.1
RUN ./configure --disable-gui
RUN make
RUN make install
RUN echo '/usr/local/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
RUN ldconfig
RUN rm -rf /opt/unixODBC-2.3.1*

# Install Netezza ODBC driver
COPY ./lib/netezza /opt/netezza/
RUN /opt/netezza/unpack -f /usr/local/nz
RUN odbcinst -i -d -f /opt/netezza/netezza.driver
RUN echo '/usr/local/nz/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
RUN echo '/usr/local/nz/lib64' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf
RUN ldconfig
RUN rm -rf /opt/netezza
RUN ln -s /usr/local/nz/bin64/nzodbcsql /usr/local/bin/nzodbcsql

WORKDIR /go/src/app
COPY . .

RUN mkdir -p ./cmd/sql-agent/vendor/github.com/chop-dbhi
RUN rm -rf ./cmd/sql-agent/vendor/github.com/chop-dbhi/sql-agent
RUN ln -s /go/src/app ./cmd/sql-agent/vendor/github.com/chop-dbhi/sql-agent

ENTRYPOINT go build -v -o ./dist/linux-amd64/sql-agent ./cmd/sql-agent
