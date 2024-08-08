FROM golang:1.22

WORKDIR /app

COPY bin/linux_amd64/api /app/api

CMD /app/api -db-dsn=${DB_DSN}