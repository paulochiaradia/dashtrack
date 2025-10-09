# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code
COPY . .

# Build the application for production
# -o /app/main: output the binary to /app/main
# CGO_ENABLED=0: disable CGO for a statically linked binary
RUN CGO_ENABLED=0 go build -o /app/main ./cmd/api

# Stage 2: Development stage with Air for hot reload
FROM golang:1.24-alpine AS dev

WORKDIR /app

# Install dependencies for migrations and networking
RUN apk add --no-cache netcat-openbsd curl

# Install migrate tool
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Install air for hot reloading
RUN go install github.com/air-verse/air@v1.60.0

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Copy migration files
COPY migrations/ /app/migrations/

# Copy and make entrypoint script executable
COPY scripts/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 8080
EXPOSE 8080

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

# Command to run air for hot reloading
CMD ["air", "-c", ".air.toml"]

# Stage 3: Production stage
FROM alpine:latest AS production

WORKDIR /app

# Install dependencies for migrations and networking
RUN apk add --no-cache netcat-openbsd curl

# Install migrate tool
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY .env .

# Copy migration files
COPY migrations/ /app/migrations/

# Copy and make entrypoint script executable
COPY scripts/docker-entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose port 8080 to the outside world
EXPOSE 8080

# Use entrypoint script
ENTRYPOINT ["/entrypoint.sh"]

# Command to run the executable
CMD ["/app/main"]
