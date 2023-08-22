# Build stage using the Go image
FROM golang:alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Build the Go app for a smaller binary size
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage using Alpine
FROM alpine:latest

# Install ca-certificates to ensure HTTPS works universally
RUN apk --no-cache add ca-certificates

# Copy the binary and .env file from the builder stage
COPY --from=builder /app/main /main
COPY --from=builder /app/.env /.env

# Expose port 8890
EXPOSE 8890

# Command to run the application
CMD ["/main"]

