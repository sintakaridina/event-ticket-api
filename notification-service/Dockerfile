FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a smaller image for the final stage
FROM alpine:latest

WORKDIR /app

# Install necessary packages
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy any config files if needed
COPY --from=builder /app/config ./config

# Expose the service port
EXPOSE 8084

# Command to run the executable
CMD ["./main"]