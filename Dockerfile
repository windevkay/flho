# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache ca-certificates git

# Copy only dependency files first to leverage caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-w -s -extldflags "-static"' -o api ./cmd/api

# Run stage - using distroless for smaller, more secure image
FROM gcr.io/distroless/static:nonroot

# Add metadata
LABEL maintainer="Kola Adebayo <kolabayo360@proton.me>"
LABEL version="1.0"
LABEL description="FLHO State Manager"

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/api /app/
# Copy any additional required files (if needed)
# COPY --from=builder /app/configs /app/configs

# Use non-root user for security (already set in distroless/nonroot)
USER nonroot:nonroot

# Expose the port the application runs on
EXPOSE 4000

# Command to run the application
ENTRYPOINT ["/app/api"]