# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and other necessary tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Assuming main.go is in cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Add ca-certificates for HTTPS calls if needed
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port (adjust if your backend uses a different port)
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
