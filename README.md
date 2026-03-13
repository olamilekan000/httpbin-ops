# httpbin

A simple HTTP request and response service in Go, similar to [httpbin.org](https://httpbin.org). Useful for testing HTTP clients, inspecting requests, and trying status codes and auth.

## Features

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

- Go 1.21+ (see `go.mod` for the exact version)

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

## API endpoints

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
