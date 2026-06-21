# --- Build Stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shrimpy ./cmd/shrimpy

# --- Run Stage ---
FROM alpine:3.19

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /shrimpy .

# Copy migration files if needed
COPY migrations/ ./migrations/

# Expose API port
EXPOSE 8080

# Run the app
CMD ["./shrimpy"]
