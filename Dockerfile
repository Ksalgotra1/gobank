# Build stage
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first and download them to cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application, specifically targeting our new cmd/gobank entrypoint
RUN CGO_ENABLED=0 GOOS=linux go build -o gobank ./cmd/gobank

# Run stage (Using a completely fresh, tiny Alpine image just for running)
FROM alpine:latest

WORKDIR /app

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/gobank .
# Also copy .env if you plan on using it inside the container
COPY .env .env

# Expose our API port so the outside world can hit it
EXPOSE 3000

# When the container turns on, boot the api and automatically seed it!
CMD ["./gobank", "--seed"]
