# Current State Discovery Report

A baseline report of what exists and what was discovered in the repository **before** applying the fixes.

---

## Summary of discoveries

| Area | Finding |
|------|--------|
| **API** | No health or readiness endpoints; no Prometheus metrics |
| **Security** | 1 reachable stdlib vuln (net/url); MD5 in auth; weak RNG; Slowloris risk; unhandled errors |
| **FIPS** | MD5 and current RNG usage are not FIPS-compliant |
| **Code quality** | Multiple errcheck and one deprecation (rand.Seed); lint timeout |
| **Build** | Makefile has wrong `ls` target after build-static; hardcoded Go path |
| **Release** | No CI/CD, no GoReleaser, no DEB/RPM/containers, no GitHub release artefacts |
| **Repo** | Module name typo; no SECURITY.md or release runbook |

---

## 1. Application & API

### Missing behaviour

- **No health check route** — There is no `/health`, `/ready`, or similar endpoint. Kubernetes liveness/readiness probes and load balancers cannot check service health.
- **No readiness route** — No way to signal that the server is ready to accept traffic (e.g. after startup).
- **No Prometheus metrics** — No `/metrics` or instrumentation for request counts, latency, or status codes. Observability is limited to logs.

### What exists

- HTTP methods, request inspection, status codes, delays, and auth (basic, bearer, digest) as described in the README.
- Server binds to configurable host/port; graceful shutdown with timeout.

---

## 2. Security & compliance

### Vulnerability scanning (govulncheck)

- **Reachable vulnerability:** **GO-2026-4601** — Incorrect parsing of IPv6 host literals in `net/url`. Present in Go 1.25.7; fixed in 1.25.8. Call chain: `server.Server.Start` → `http.Server.ListenAndServe` → `url.Parse` / `url.ParseRequestURI` ([details](https://pkg.go.dev/vuln/GO-2026-4601)).
- Two further vulns were reported in dependencies/modules that the code does not appear to call (informational).

### Security-focused static analysis (gosec)

| Finding | Location | Severity | Issue |
|---------|----------|----------|--------|
| Weak RNG | `internal/handlers/status.go:84` | HIGH | `math/rand` used instead of `crypto/rand` for weighted status |
| Weak crypto (MD5) | `internal/handlers/auth.go:178`, import at line 4 | MEDIUM | MD5 used in `generateOpaque`; blocklisted import |
| Slowloris risk | `internal/server/server.go:22–25` | MEDIUM | `ReadHeaderTimeout` not set on `http.Server` |
| Unhandled errors | `auth.go:170,177`, `methods.go:78` | LOW | `rand.Read`, `Encode` return values not checked |

**FIPS impact:** Use of MD5 and current RNG patterns would need to change for FIPS-compliant builds (e.g. SHA-256 for opaque, and review of RNG usage).

---

## 3. Code quality

### Linting (golangci-lint)

- **errcheck:** Unchecked return values for `w.Write` (3× in middleware tests), `rand.Read` (2× in auth), `json.Encoder.Encode` (methods).
- **staticcheck SA1019:** `rand.Seed` in `internal/handlers/status.go:13` is deprecated since Go 1.20; prefer `rand.New(rand.NewSource(seed))` for a local generator.

**Full golangci-lint output** (`golangci-lint run ./... --timeout=5m`):

```
internal/middleware/middleware_test.go:14:10: Error return value of `w.Write` is not checked (errcheck)
                w.Write([]byte("test response"))
internal/middleware/middleware_test.go:42:10: Error return value of `w.Write` is not checked (errcheck)
                w.Write([]byte("error"))
internal/middleware/middleware_test.go:69:10: Error return value of `w.Write` is not checked (errcheck)
                w.Write([]byte("part1"))
internal/handlers/auth.go:170:11: Error return value of `rand.Read` is not checked (errcheck)
        rand.Read(b)
internal/handlers/auth.go:177:11: Error return value of `rand.Read` is not checked (errcheck)
        rand.Read(b)
internal/handlers/methods.go:78:27: Error return value of `(*encoding/json.Encoder).Encode` is not checked (errcheck)
        json.NewEncoder(w).Encode(data)
internal/handlers/status.go:13:2: SA1019: rand.Seed has been deprecated since Go 1.20 and an alternative has been available since Go 1.0: As of Go 1.20 there is no reason to call Seed with a random value. Programs that call Seed with a known value to get a specific sequence of results should use New(NewSource(seed)) to obtain a local random generator. (staticcheck)
        rand.Seed(time.Now().UnixNano())
```

---

## 4. Build, packaging & release

### Build (Makefile)

- **Works:** `make build` (dynamic, CGO=1) and `make build-static` (static, CGO=0) both succeed. Outputs: `bin/httpbin`, `bin/httpbin-static` (~5.4M each). LDFLAGS inject version, buildTime, commit.
- **Issues:**
  - After `build-static`, the target runs `ls -lh $(OUTPUT_DIR)/$(BINARY_NAME)` — i.e. it lists the dynamic binary name, not `httpbin-static`. Cosmetic only.
  - Makefile uses a hardcoded Go path (`/usr/local/go/bin/go`); less portable than relying on `PATH`.

### Missing for assignment

- No **GitHub Actions** workflows (no test-on-PR, no release-on-tag).
- No **GoReleaser** config (no multi-arch binaries, no DEB/RPM, no container images, no GitHub release).
- No **DEB** or **RPM** packages.
- No **container image** build or publish.
- No **FIPS** build path or FIPS-specific artefacts.
- No **release runbook** or clear “how to cut a release” in docs.

---

## 5. Repo & config

- **go.mod:** Go 1.25.7; module path has typo: `tyk-devops-assignement` (should be `assignment` if corrected).
- **No SECURITY.md** — No documented way to report vulnerabilities.
- **No .github/workflows** or **.goreleaser.yaml** (as above).
- **Documentation:** README describes API and local build/run; no instructions yet for downloading from releases or running containers.

---

## 6. Tool run summary (evidence)

| Check | Command | Result |
|-------|---------|--------|
| govulncheck | `govulncheck ./...` | 1 reachable stdlib vuln (GO-2026-4601); 2 further in deps not in call path |
| gosec | `gosec ./...` | 7 findings (1 HIGH, 3 MEDIUM, 3 LOW) — see Section 2 |
| golangci-lint | `golangci-lint run ./...` | 7 findings (errcheck + SA1019) — see Section 3; timeout |
| go build | `make build` / `make build-static` | Both succeed; binaries in `bin/` |

