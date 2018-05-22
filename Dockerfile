FROM debian:jessie-slim

ENV LD_LIBRARY_PATH /usr/lib/instantclient_12_1
ENV ORACLE_HOME /usr/lib/instantclient_12_1

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

COPY ./lib/oracle/instantclient_12_1 /usr/lib/instantclient_12_1
COPY ./lib/oracle/oci8.pc /usr/lib/pkgconfig/
RUN ln -s /usr/lib/instantclient_12_1/libclntsh.so.12.1 /usr/lib/instantclient_12_1/libclntsh.so
RUN ln -s /usr/lib/instantclient_12_1/libocci.so.12.1 /usr/lib/instantclient_12_1/libocci.so

COPY ./lib/unixODBC-2.3.1.tar.gz /opt/
RUN cd /opt && \
  tar xf unixODBC-2.3.1.tar.gz && \
  cd /opt/unixODBC-2.3.1 && \
  ./configure --disable-gui && \
  make && \
  make install && \
  echo '/usr/local/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf && \
  ldconfig && \
  rm -rf /opt/unixODBC-2.3.1*

# Install Netezza ODBC driver
COPY ./lib/netezza /opt/netezza/
RUN /opt/netezza/unpack -f /usr/local/nz && \
  odbcinst -i -d -f /opt/netezza/netezza.driver && \
  echo '/usr/local/nz/lib' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf && \
  echo '/usr/local/nz/lib64' >> /etc/ld.so.conf.d/x86_64-linux-gnu.conf && \
  ldconfig && \
  rm -rf /opt/netezza && \
  ln -s /usr/local/nz/bin64/nzodbcsql /usr/local/bin/nzodbcsql

RUN apt-get remove -y aptitude g++ libc6-dev gcc && \
  apt-get -y autoremove && apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY ./dist/linux-amd64/sql-agent /usr/bin/sql-agent

CMD ["sql-agent", "-host", "0.0.0.0"]
