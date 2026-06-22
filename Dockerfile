# --- Build Stage ---
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go applications
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shrimpy ./cmd/shrimpy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /shrimpy-db ./cmd/db

# --- Run Stage ---
FROM alpine:3.19

# Install certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binaries from builder
COPY --from=builder /shrimpy .
COPY --from=builder /shrimpy-db .

# Copy migration files if needed
COPY migrations/ ./migrations/

# Expose API port
EXPOSE 8080

# Run the app
CMD ["./shrimpy"]
