# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Copy mod files first for better build caching
COPY go.mod go.sum ./

# ARG for GitHub token (passed from CI)
ARG GITHUB_TOKEN

# Configure git auth + Go private modules
RUN if [ -n "$GITHUB_TOKEN" ]; then \
      git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"; \
    fi && \
    go env -w GOPRIVATE=github.com/alphabatem/*,github.com/lac-hong-legacy/* && \
    go mod download

# Copy the rest of the app
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./runtime/

# ---- Runtime Stage ----
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]
