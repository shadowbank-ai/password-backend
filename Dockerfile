# Use the official Golang image as a build stage
FROM golang:1.23 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o server .

# Use a minimal image for deployment
FROM debian:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/server .

# Set the environment variable for DigitalOcean
ENV PORT=8080

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./server"]
