# Build stage: use official Go image (1.25)
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Copy all source code
COPY . .

# Build statically (good for Alpine)
RUN CGO_ENABLED=0 go build -o app ./cmd

# Final stage: minimal runtime image
FROM alpine:latest

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/app .

# Expose app port
EXPOSE 8080

# Run the binary
CMD ["./app"]