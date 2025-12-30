# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o docker-cleanup .

# Run stage
FROM alpine:latest

LABEL org.opencontainers.image.title="Docker Cleanup Tool" \
      org.opencontainers.image.description="Automated tool to clean up unused Docker containers, volumes, networks, and images." \
      org.opencontainers.image.source="https://github.com/rafael-qsantos/Management-go" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.authors="Rafael Queiroz Santos"

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/docker-cleanup .

# Command to run the binary
ENTRYPOINT ["./docker-cleanup"]
