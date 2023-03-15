FROM debian:bullseye-slim

# current versions:
# https://download.oracle.com/otn_software/linux/instantclient/216000/instantclient-sdk-linux.x64-21.6.0.0.0dbru.zip
# http://www.unixodbc.org/unixODBC-2.3.11.tar.gz
ARG INSTANTCLIENT_VERSION=21_6
ARG UNIXODBC_VERSION=2.3.11

ENV LD_LIBRARY_PATH /usr/lib/instantclient_${INSTANTCLIENT_VERSION}
ENV ORACLE_HOME /usr/lib/instantclient_${INSTANTCLIENT_VERSION}

RUN apt-get update -qq && \
    apt-get install -y --no-install-recommends \
      libaio1 \
      g++ \
      gcc \
      libc6-dev \
      make \
      pkg-config \
      ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY ./lib/oracle/instantclient_${INSTANTCLIENT_VERSION} /usr/lib/instantclient_${INSTANTCLIENT_VERSION}
COPY ./lib/oracle/oci8.pc /usr/lib/pkgconfig/

COPY ./lib/unixODBC-${UNIXODBC_VERSION}.tar.gz /opt/
RUN cd /opt && \
  tar xf unixODBC-${UNIXODBC_VERSION}.tar.gz && \
  cd /opt/unixODBC-${UNIXODBC_VERSION} && \
  ./configure --disable-gui && \
  make && \
  make install && \
  echo '/usr/local/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf && \
  ldconfig && \
  rm -rf /opt/unixODBC-${UNIXODBC_VERSION}*

RUN apt-get remove -y aptitude g++ libc6-dev gcc && \
  apt-get -y autoremove && apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY ./dist/linux-amd64/sql-agent /usr/bin/sql-agent

CMD ["sql-agent", "-host", "0.0.0.0"]
