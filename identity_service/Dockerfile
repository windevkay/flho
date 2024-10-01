FROM alpine:latest

RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY bin/linux_amd64/api /app/api

CMD /app/api -db-dsn=${DB_DSN} -mailer-server=${MAILER_SERVER} -message-queue-server=${MESSAGE_QUEUE_SERVER}