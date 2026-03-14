ARG BUILDER_IMAGE=golang:1.25-bookworm
ARG RUN_IMAGE=debian:bookworm-slim

FROM ${BUILDER_IMAGE} AS builder
WORKDIR /app

RUN if command -v apk >/dev/null; then \
        apk add --no-cache git ca-certificates; \
    else \
        apt-get update && apt-get install -y --no-install-recommends ca-certificates git && rm -rf /var/lib/apt/lists/*; \
    fi

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CGO_ENABLED=1
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME
ARG GOFLAGS=""

RUN CGO_ENABLED=${CGO_ENABLED} go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.commit=${COMMIT}" \
    -o /httpbin ./cmd/httpbin

FROM ${RUN_IMAGE}

RUN if command -v apk >/dev/null; then \
        apk add --no-cache ca-certificates wget; \
        adduser -D -u 65532 appuser; \
    else \
        apt-get update && apt-get install -y --no-install-recommends ca-certificates wget && rm -rf /var/lib/apt/lists/*; \
        adduser --disabled-password --gecos '' --uid 65532 appuser; \
    fi

COPY --from=builder /httpbin /httpbin

EXPOSE 8088
USER appuser
ENTRYPOINT ["/httpbin"]
CMD ["-port", "8088"]
