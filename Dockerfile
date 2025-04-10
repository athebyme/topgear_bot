FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application as a fully static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o forza-bot ./cmd/bot/main.go

# Use a minimal base image
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Copy compiled application
COPY --from=builder /app/forza-bot /app/
COPY --from=builder /app/configs /app/configs

# Run as non-root user
USER nonroot:nonroot

# Run the application
CMD ["/app/forza-bot"]