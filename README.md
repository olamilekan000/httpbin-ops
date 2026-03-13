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

## Quick start

```bash
make run
```

Server listens on `http://0.0.0.0:8088` by default. Try: `curl http://localhost:8088/get`.

## Requirements

- Go 1.25+ (see `go.mod` for the exact version)

## Building and testing

| Command           | Description                    |
|-------------------|--------------------------------|
| `make build`      | Build dynamic binary → `bin/httpbin` |
| `make build-static` | Build static binary → `bin/httpbin-static` |
| `make test`       | Run tests (with race detector)  |
| `make run`        | Build and run with defaults    |

Lint (optional):

```bash
golangci-lint run ./... --timeout=5m
```

## Configuration

| Flag      | Default   | Description              |
|-----------|-----------|--------------------------|
| `-host`   | `0.0.0.0` | Host to bind to          |
| `-port`   | `8088`    | Port to bind to          |
| `-version`| —         | Print version and exit   |

Example: `./bin/httpbin -port 9000`

## Docker

Per the task specification, there are two image variants: **statically linked** (CGO_ENABLED=0) and **dynamically linked** (CGO_ENABLED=1).

**Static (default Dockerfile):**
```bash
docker build -t httpbin:latest .
docker run -p 8088:8088 httpbin:latest
```

**Dynamic (links against glibc; use Debian-based image):**
```bash
docker build -f Dockerfile.dynamic -t httpbin:dynamic .
docker run -p 8088:8088 httpbin:dynamic
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
- **Prometheus** — http://localhost:9090 (scrapes httpbin at the configured port)

## Releases

On push of a version tag (`v*`), the [release workflow](.github/workflows/release.yml) runs [GoReleaser](https://goreleaser.com), which produces:

- **Archives** — `tar.gz` for linux amd64/arm64 (static, dynamic, FIPS).
- **DEB / RPM** — `.deb` and `.rpm` packages (Debian/Ubuntu and RHEL/Fedora). RPM archs use `x86_64` and `aarch64`. Three package variants: `httpbin` (static), `httpbin-dynamic`, `httpbin-fips`. **RPMs are built in a RHEL 8–compatible environment so they are installable and runnable on RHEL 8, RHEL 9, and latest Fedora.**
- **Container images** (multi-arch `linux/amd64`, `linux/arm64`) on GitHub Container Registry:
  - `ghcr.io/<owner>/<repo>:<version>-static` and `latest-static`
  - `ghcr.io/<owner>/<repo>:<version>-dynamic` and `latest-dynamic`
  - `ghcr.io/<owner>/<repo>:<version>-fips` and `latest-fips`

**RPM (RHEL 8 / RHEL 9 / Fedora):** The release workflow builds binaries and RPMs inside a CentOS Stream 8 container so that the dynamic binary links against glibc 2.28. All generated RPMs (static, dynamic, FIPS) are installable and runnable on RHEL 8, RHEL 9, and latest Fedora.

**FIPS:** The FIPS build uses the same source as the static build. For FIPS-validated compliance you must build with a FIPS-approved Go toolchain (e.g. BoringCrypto/Go FIPS) in your own pipeline; the release workflow provides the packaging and image layout.

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
