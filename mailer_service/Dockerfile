# Build
FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o api ./cmd/api

# Run
FROM alpine:3.14

WORKDIR /app

COPY --from=builder /app/api .

EXPOSE 4000

ENTRYPOINT ["./api"]