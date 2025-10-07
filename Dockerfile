FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

ARG GITHUB_TOKEN
RUN --mount=type=secret,id=GIT_AUTH_TOKEN \
    TOKEN=${GITHUB_TOKEN:-$(cat /run/secrets/GIT_AUTH_TOKEN 2>/dev/null || echo "")} && \
    if [ -n "$TOKEN" ]; then \
      git config --global url."https://${TOKEN}@github.com/".insteadOf "https://github.com/" && \
      echo "machine github.com login ${TOKEN}" > ~/.netrc; \
    fi && \
    go env -w GOPRIVATE=github.com/alphabatem/* && \
    go mod download && \
    if [ -n "$TOKEN" ]; then \
      git config --global --unset url."https://${TOKEN}@github.com/".insteadOf 2>/dev/null || true; \
    fi

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./runtime/

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
