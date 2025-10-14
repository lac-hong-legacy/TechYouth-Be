# ---- Builder Stage ----
FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Copy mod files first for better build caching
COPY go.mod go.sum ./

# Configure git auth + Go private modules
RUN go mod download

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
