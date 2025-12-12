# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for libvips
RUN apk add --no-cache \
    gcc \
    musl-dev \
    vips-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o ipxpress ./cmd/ipxpress

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    vips \
    ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/ipxpress .

# Expose the default port
EXPOSE 8080

# Run the application
CMD ["./ipxpress"]
