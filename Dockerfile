# Statically linked image (CGO_ENABLED=0) per task specification — one of two required variants (static + dynamic).
# Build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME
# Static binary: no CGO, runs on any Linux (e.g. Alpine, distroless)
RUN CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.commit=${COMMIT}" -o /httpbin ./cmd/httpbin

# Run stage (Alpine for small image + wget for healthcheck; run as non-root)
FROM alpine:3.19
RUN apk add --no-cache ca-certificates wget
RUN adduser -D -u 65532 appuser
COPY --from=builder /httpbin /httpbin
# Default port; override with -port or in compose via HTTPBIN_PORT
EXPOSE 8088
USER appuser
ENTRYPOINT ["/httpbin"]
CMD ["-port", "8088"]
