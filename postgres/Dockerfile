FROM postgres:latest

COPY postgres.key /var/lib/postgresql/server.key
COPY postgres.crt /var/lib/postgresql/server.crt

RUN chmod 600 /var/lib/postgresql/server.key
RUN chown postgres:postgres /var/lib/postgresql/server.key

CMD ["postgres", "-c", "ssl=on", "-c", "ssl_cert_file=/var/lib/postgresql/server.crt", "-c", "ssl_key_file=/var/lib/postgresql/server.key", "-c", "log_connections=on", "-c", "log_disconnections=on"]