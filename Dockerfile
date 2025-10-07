FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies using GitHub token
# The token is passed as a build arg or secret
ARG GITHUB_TOKEN
RUN --mount=type=secret,id=GIT_AUTH_TOKEN \
    TOKEN=${GITHUB_TOKEN:-$(cat /run/secrets/GIT_AUTH_TOKEN 2>/dev/null || echo "")} && \
    if [ -n "$TOKEN" ]; then \
      git config --global url."https://${TOKEN}@github.com/".insteadOf "https://github.com/" && \
      echo "machine github.com login ${TOKEN}" > ~/.netrc; \
    fi && \
    go env -w GOPRIVATE=github.com/alphabatem/* && \
    go mod download && \
    git config --global --unset url."https://${TOKEN}@github.com/".insteadOf 2>/dev/null || true

# Copy source code
COPY . .

# Build the application
RUN GOTOOLCHAIN=auto CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./runtime/

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
