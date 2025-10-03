FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates openssh-client

# Configure SSH for private repositories
RUN mkdir -p /root/.ssh && \
    chmod 0700 /root/.ssh && \
    ssh-keyscan github.com > /root/.ssh/known_hosts

# Copy go mod files
COPY go.mod go.sum ./

# Add SSH key and download dependencies (mount as secret)
RUN --mount=type=secret,id=ssh_key,target=/tmp/ssh_key \
    cp /tmp/ssh_key /root/.ssh/id_rsa && \
    chmod 600 /root/.ssh/id_rsa && \
    git config --global url."git@github.com:".insteadOf "https://github.com/" && \
    go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./runtime/

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Expose port
EXPOSE 8000

# Run the application
CMD ["./main"]
