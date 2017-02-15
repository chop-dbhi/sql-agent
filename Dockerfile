FROM debian:jessie-slim

ENV LD_LIBRARY_PATH /usr/lib/instantclient_12_1
ENV ORACLE_HOME /usr/lib/instantclient_12_1

RUN apt-get update -qq && apt-get install -y --no-install-recommends \
      libaio1 \
      pkg-config \
      && rm -rf /var/lib/apt/lists/*

COPY ./lib/oracle/instantclient_12_1 /usr/lib/instantclient_12_1
COPY ./lib/oracle/oci8.pc /usr/lib/pkgconfig/
COPY ./dist/linux-amd64/sql-agent /usr/bin/sql-agent

CMD ["sql-agent", "-host", "0.0.0.0"]
