# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

Official Go client library for the CloudConnexa API (OpenVPN). Module path: `github.com/openvpn/cloudconnexa-go-client/v2`. Requires Go 1.24+.

## Common commands

```bash
make test     # Unit tests: go test -v -race ./cloudconnexa/...
make e2e      # E2E tests: go test -v -race ./e2e/...   (requires live API credentials)
make lint     # golangci-lint run
make build    # go build -v ./...
make deps     # go mod download
```

Run a single unit test:
```bash
go test -v -race -run TestNewClient ./cloudconnexa/
```

E2E tests require these environment variables:
- `CLOUDCONNEXA_BASE_URL` (e.g., `https://your-org.api.openvpn.com`)
- `CLOUDCONNEXA_CLIENT_ID`
- `CLOUDCONNEXA_CLIENT_SECRET`

## Architecture

### Client construction and the service pattern

`cloudconnexa/cloudconnexa.go` is the heart of the library. `NewClient` performs an OAuth2 client-credentials handshake at `/api/v1/oauth/token`, then builds a `Client` struct exposing one field per resource (`Networks`, `Users`, `DNSRecords`, `Sessions`, etc.).

All service types follow the same pattern:
```go
type FooService service          // service is { client *Client }
c.Foo = (*FooService)(&c.common) // services share the same backing struct
```
This means every service has access to the shared `*Client` (token, rate limiters, HTTP client) without duplication. When adding a new resource service, mirror the existing services: define `type XService service`, attach methods on `*XService`, and wire it up in `NewClient` via `c.X = (*XService)(&c.common)`.

### Request flow

Every request goes through `Client.DoRequest`:
1. Picks `ReadRateLimiter` (GET) or `UpdateRateLimiter` (other verbs) and waits.
2. Adds `Authorization: Bearer <token>`, `User-Agent`, and `Content-Type` if unset (`setCommonHeaders`).
3. Reads the body through an `io.LimitReader` capped at `DefaultMaxResponseSize` (10 MB) — bodies that exceed this fail with `ErrResponseTooLarge` (CWE-400 mitigation). The OAuth token response is bounded separately at 1 MB.
4. Returns `*ErrClientResponse` (with `StatusCode()` / `Body()` accessors) for non-2xx responses.
5. Calls `AssignLimits` to dynamically adjust the rate limiter from `X-RateLimit-Replenish-Rate`, `X-RateLimit-Replenish-Time`, and `X-RateLimit-Remaining` response headers.

When writing service methods, always go through `c.client.DoRequest` — it's the single chokepoint for auth, rate limiting, body-size enforcement, and error wrapping. Don't call `c.client.client.Do` directly.

### URL construction

- Use `c.client.GetV1Url()` to get `<base>/api/v1`.
- Use `buildURL(base, segments...)` to concatenate path segments — it `url.PathEscape`s each segment, which matters for IDs that may contain reserved characters.
- Inline `fmt.Sprintf` is used for query-string-bearing endpoints (e.g., `?page=&size=`); follow the existing convention in the file you're editing.

### Base URL validation

`validateBaseURL` enforces:
- HTTPS only by default.
- HTTP allowed only for loopback hosts (`localhost`, 127.0.0.0/8, `::1`) and only when `ClientOptions.AllowInsecureHTTP` is true. This is the path mock-server tests use.
- Rejects URLs with embedded credentials, missing scheme/host, or non-http(s) schemes.
- Strips path/query/fragment and returns `scheme://host` only.

If you're writing a unit test against `httptest.NewServer`, construct the client via `NewClientWithOptions(url, id, secret, &ClientOptions{AllowInsecureHTTP: true})` — `NewClient` will reject the loopback HTTP URL.

### Pagination conventions

Two pagination styles coexist:
- **Page-based (legacy)**: `GetByPage(page, size)` returns a `*PageResponse` struct with `Content`, `TotalPages`, etc.; `List()` loops through all pages using `defaultPageSize = 100`. Used by Networks, Users, DNS Records, Connectors, etc.
- **Cursor-based**: Sessions API uses `SessionsListOptions` with `Cursor` and a `NextCursor` in the response.

When adding a new collection endpoint, match whichever style the upstream API uses and follow the existing service's structure for consistency.

### Errors

Sentinel errors live in `cloudconnexa/errors.go` (`ErrCredentialsRequired`, `ErrEmptyID`, `ErrResponseTooLarge`, `ErrInvalidBaseURL`, `ErrHTTPSRequired`). Resource-specific not-found errors (e.g., `ErrUserNotFound`, `ErrDNSRecordNotFound`) are defined in their respective service files. API-level errors are returned as `*ErrClientResponse` — callers extract status with `errors.As(err, &apiErr); apiErr.StatusCode()`.

`validateID(id)` is the standard guard at the top of any method that takes an ID parameter — call it before constructing the URL.

## Testing notes

- Unit tests live alongside the code (`*_test.go` in `cloudconnexa/`) and use `httptest.NewServer` mock servers; `_test.go` files are excluded from `gosec` lint (see `.golangci.yml`).
- E2E tests in `e2e/client_test.go` hit a live API and include retry/backoff for 429s. `TestCreateNetwork` searches for a non-overlapping RFC1918 /24 subnet to avoid collisions in CI matrix runs — preserve that logic if modifying network-creation tests.
- Always run `-race` (the Makefile already does); the client is intended to be safe for concurrent use.

## API version

Targets CloudConnexa API v1.2.0. Field structs (e.g., `User`, `Session`) are documented as matching specific API schema versions — keep those doc comments accurate when DTOs change.
