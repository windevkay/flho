FROM golang:1.22

WORKDIR /app

COPY bin/linux_amd64/api /app/api

CMD /app/api -smtp-host=${SMTP_HOST} -smtp-username=${SMTP_USERNAME} -smtp-password=${SMTP_PASSWORD}