# Use the official Golang image (Debian-based) as a build stage
FROM golang:1.24 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 for static build, GOOS=linux for Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o lookatthatmongo .

# Use a minimal Debian-based image for the final stage
FROM debian:stable-slim

# Install necessary CA certificates for TLS verification
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && rm -rf /var/lib/apt/lists/*

# Set the Current Working Directory inside the container
WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/lookatthatmongo .

# Create a non-root user for security
RUN useradd -m appuser
USER appuser

# Expose port if the application listens on one (adjust if needed)
# EXPOSE 8080

# Command to run the executable
CMD ["./lookatthatmongo"] 