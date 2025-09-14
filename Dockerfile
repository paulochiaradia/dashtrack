# Stage 1: Builder
FROM golang:1.23-alpine AS builder

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

# Stage 2: Development stage (simplified without air for now)
FROM builder AS dev

# Expose port 8080
EXPOSE 8080

# Command to run the application directly
CMD ["/app/main"]

# Stage 3: Production stage
FROM alpine:latest AS production

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/main .
COPY .env .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]
