# httpbin

A simple HTTP request and response service in Go, similar to [httpbin.org](https://httpbin.org). Useful for testing HTTP clients, inspecting requests, and trying status codes and auth.

## Features

- **Health** — `GET /health` for liveness/readiness (includes version from build)
- **Prometheus** — `GET /metrics` for request count and duration
- **HTTP methods** — GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
- **Request inspection** — headers, IP, User-Agent
- **Status codes** — fixed or weighted random
- **Delays** — configurable response delay (max 10s)
- **Auth** — Basic, Bearer token, Digest (simplified)

## Links

- **[Docker Hub Tags](https://hub.docker.com/r/olamilekan001/httpbin/tags)**
- **[GitHub Container Registry](https://github.com/olamilekan000/httpbin-ops/pkgs/container/httpbin-ops/versions)**
- **[Latest Release (v0.1.27)](https://github.com/olamilekan000/httpbin-ops/releases/tag/v0.1.27)**
- **[All Releases](https://github.com/olamilekan000/httpbin-ops/releases)**
- **[Helm Chart](./httpbin-ops-charts)**

## Quick start

```bash
make run
```

Server listens on `http://0.0.0.0:8088` by default. Try: `curl http://localhost:8088/get`.

## Requirements

- Go 1.25+ (see `go.mod` for the exact version)

## Building and testing

| Command             | Description                                |
| ------------------- | ------------------------------------------ |
| `make build`        | Build dynamic binary → `bin/httpbin`       |
| `make build-static` | Build static binary → `bin/httpbin-static` |
| `make test`         | Run tests (with race detector)             |
| `make run`          | Build and run with defaults                |

Lint (optional):

```bash
golangci-lint run ./... --timeout=5m
```

## Configuration

| Flag       | Default   | Description            |
| ---------- | --------- | ---------------------- |
| `-host`    | `0.0.0.0` | Host to bind to        |
| `-port`    | `8088`    | Port to bind to        |
| `-version` | —         | Print version and exit |

Example: `./bin/httpbin -port 9000`

## Docker

Per the task specification, there are two image variants: **statically linked** (CGO_ENABLED=0) and **dynamically linked** (CGO_ENABLED=1).

**Local builds (using the unified `Dockerfile`):**

- **Dynamic (default):**

  ```bash
  docker build -t httpbin:latest .
  ```

- **Static (using Alpine):**
  ```bash
   docker build --build-arg CGO_ENABLED=0 --build-arg BUILDER_IMAGE=golang:1.25-alpine --build-arg RUN_IMAGE=alpine:3.19 -t httpbin:static .
  ```

Port is configurable via `-port`:

```bash
docker run -p 9000:9000 httpbin:latest -port 9000
```

Run with Docker Compose (httpbin + Prometheus scraping `/metrics`):

```bash
docker compose up -d
```

Port is configurable with the `HTTPBIN_PORT` env var (default 8088). If you change it, update `prometheus.yml` so the scrape target port matches (e.g. `httpbin:9000`).

```bash
HTTPBIN_PORT=9000 docker compose up -d
```

- **httpbin** — http://localhost:8088 (or `http://localhost:${HTTPBIN_PORT}`)
- **Prometheus UI** — http://localhost:9090 (scrapes httpbin at the configured port)

### Monitoring & Metrics

The application exports Prometheus metrics on `/metrics`. When running with Docker Compose, Prometheus is pre-configured to scrape these metrics every 15 seconds.

Key metrics include:

- `http_request_duration_seconds`: Histogram of request latencies.
- `http_requests_total`: Counter of total HTTP requests.
- Standard Go process metrics (memory, GC, etc.).

## Helm

A Helm chart is available in the `httpbin-ops-charts` directory.

To install the chart:

```bash
helm install httpbin ./httpbin-ops-charts
```

By default, it uses the `latest-static` image. You can specify a different variant using `--set image.tag`:

```bash
helm install httpbin ./httpbin-ops-charts --set image.tag=latest-fips
```

## Releases

On push of a version tag (`v*`), the [release workflow](.github/workflows/release.yml) runs [GoReleaser](https://goreleaser.com), which produces:

- **Archives** — `tar.gz` for linux amd64/arm64 (static, dynamic, FIPS).
- **DEB / RPM** — `.deb` and `.rpm` packages (Debian/Ubuntu and RHEL/Fedora). RPM archs use `x86_64` and `aarch64`. Three package variants: `httpbin` (static), `httpbin-dynamic`, `httpbin-fips`. **RPMs are built in a RHEL 8–compatible environment so they are installable and runnable on RHEL 8, RHEL 9, and latest Fedora.**
- **Container images** (multi-arch `linux/amd64`, `linux/arm64`) on GitHub Container Registry (built via `Dockerfile.release`):
  - `ghcr.io/<owner>/<repo>:<version>-static` and `latest-static` (Alpine based)
  - `ghcr.io/<owner>/<repo>:<version>-dynamic` and `latest-dynamic` (Debian based)
  - `ghcr.io/<owner>/<repo>:<version>-fips` and `latest-fips` (Alpine based)

**RPM (RHEL 8 / RHEL 9 / Fedora):** The release workflow builds binaries and RPMs inside a CentOS Stream 8 container so that the dynamic binary links against glibc 2.28. All generated RPMs (static, dynamic, FIPS) are installable and runnable on RHEL 8, RHEL 9, and latest Fedora.

To test this manually on your machine:

1.  Generate a local snapshot release: `make snapshot`
2.  Run the smoke test: `make smoke-test-rpm`

**FIPS:** The FIPS build is compiled with **Go BoringCrypto** (`GOEXPERIMENT=boringcrypto`, Go 1.19+ on Linux amd64/arm64) and imports `crypto/tls/fipsonly` so TLS uses only FIPS-approved settings. The release workflow produces FIPS-oriented binaries, packages, and container images; for full FIPS 140-2 validation you must follow your organization’s certification process.

#### Verifying FIPS compliance

> [!IMPORTANT]
> BoringCrypto symbols are only generated when building for **Linux** (amd64 or arm64). On macOS or Windows, the build will succeed but use standard Go crypto.

To verify compliance:

1. **Build for Linux** (using Docker if you are on macOS):

   ```bash
   docker run --rm -v $(pwd):/app -w /app golang:1.25 \
     sh -c "GOEXPERIMENT=boringcrypto CGO_ENABLED=1 go build -tags fips -o httpbin-fips-linux ./cmd/httpbin && \
            go tool nm httpbin-fips-linux | grep _Cfunc__goboringcrypto_"
   ```

2. **Check for symbols**:
   If the binary is successfully linked with the FIPS-validated BoringSSL module, you will see output like this:

   ```text
   92ba28 D crypto/internal/boring._cgo_32f3ac20d4c4_Cfunc__goboringcrypto_BORINGSSL_bcm_power_on_self_test
   92bbb0 D crypto/internal/boring._cgo_32f3ac20d4c4_Cfunc__goboringcrypto_HMAC_CTX_cleanup
   92bc60 D crypto/internal/boring._cgo_32f3ac20d4c4_Cfunc__goboringcrypto_SHA256_Init
   ...
   ```

3. **Runtime Verification**:
   When running the FIPS binary, it will strictly enforce FIPS-only TLS algorithms. You can also check if the binary is using BoringCrypto at runtime by checking the `crypto/tls` behavior or using `GODEBUG=fipsdebug=1` (on supported Go versions) to see FIPS initialization details.

## API endpoints

### Health

- **`GET /health`** — Returns 200 and JSON with `status`, `version`, `commit`, `build_time` (injected at build). Use for liveness/readiness.

```bash
curl http://localhost:8088/health
```

- **`GET /metrics`** — Prometheus metrics (request count, request duration, process metrics).

```bash
curl http://localhost:8088/metrics
```

### HTTP methods

Return request details (method, headers, query, body, origin).

- `GET /get`, `POST /post`, `PUT /put`, `PATCH /patch`, `DELETE /delete`, `HEAD /head`, `OPTIONS /options`

### Request inspection

- **`GET /headers`** — All request headers
- **`GET /ip`** — Origin IP
- **`GET /user-agent`** — User-Agent header

### Status codes

- **`GET /status/{code}`** — Return the given HTTP status (e.g. 200, 404, 500).
- **`GET /status/{code1}:{weight1},{code2}:{weight2}`** — Weighted random (e.g. 90% 200, 10% 500).

```bash
curl http://localhost:8088/status/404
curl http://localhost:8088/status/200:0.9,500:0.1
```

### Delays

- **`GET /delay/{seconds}`** — Delay response (max 10 seconds).

```bash
curl http://localhost:8088/delay/2
```

### Authentication

- **`GET /basic-auth/{user}/{passwd}`** — HTTP Basic auth.

```bash
curl -u user:passwd http://localhost:8088/basic-auth/user/passwd
```

- **`GET /bearer`** — Bearer token (any non-empty token accepted).

```bash
curl -H "Authorization: Bearer my-token" http://localhost:8088/bearer
```

- **`GET /digest-auth/{qop}/{user}/{passwd}`** — HTTP Digest (simplified).

```bash
curl --digest -u user:passwd http://localhost:8088/digest-auth/auth/user/passwd
```

## License

See [LICENSE.md](LICENSE.md).
