# go-fizzbuzz

A configurable fizz-buzz REST API written in Go, using only the standard library.

```console
$ curl -s "localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]
```

- [Objective](#objective)
- [Quick start](#quick-start)
- [API](#api)
- [Design decisions](#design-decisions)
- [Project layout](#project-layout)
- [Development](#development)
- [Performance](#performance)
- [Limitations and next steps](#limitations-and-next-steps)
- [How I worked](#how-i-worked)

## Objective

Implement an HTTP endpoint that returns a fizz-buzz sequence built from query parameters.

The original brief, verbatim:

```text
The original fizz-buzz consists in writing all numbers from 1 to 100, and just replacing all multiples of 3 by "fizz", all multiples of 5 by "buzz", and all multiples of 15 by "fizzbuzz".
The output would look like this: "1,2,fizz,4,buzz,fizz,7,8,fizz,buzz,11,fizz,13,14,fizzbuzz,16,...".

Your goal is to implement a web server that will expose a REST API endpoint that:
- Accepts five parameters: three integers int1, int2 and limit, and two strings str1 and str2.
- Returns a list of strings with numbers from 1 to limit, where: all multiples of int1 are replaced by str1, all multiples of int2 are replaced by str2, all multiples of int1 and int2 are replaced by str1str2.
 
The server needs to be:
- Ready for production
- Easy to maintain by other developers

Bonus: add a statistics endpoint allowing users to know what the most frequent request has been.

This endpoint should:
- Accept no parameter
- Return the parameters corresponding to the most used request, as well as the number of hits for this request
```

## Quick start

Requirements: Go 1.26 or later, or Docker.

```console
$ make run
go run ./cmd/server
... INFO starting server, will listen and serve addr=:8080
```

With Docker:

```console
$ docker build -t go-fizzbuzz .
$ docker run --rm -p 8080:8080 go-fizzbuzz
```

`make docker` builds the same image, tagged with `git describe` and carrying version labels.

The server listens on `:8080` and stops gracefully on `SIGINT` or `SIGTERM`.

## API

### `GET /fizzbuzz`

Returns the sequence as a JSON array of strings.

| Parameter | Type    | Rule                               |
|-----------|---------|------------------------------------|
| `int1`    | integer | required, `> 0`                    |
| `int2`    | integer | required, `> 0`                    |
| `limit`   | integer | required, `> 0`, `≤ 1024`          |
| `str1`    | string  | required, not empty, `≤ 64` bytes  |
| `str2`    | string  | required, not empty, `≤ 64` bytes  |

Numbers from 1 to `limit` are printed as-is, except:

- multiples of both `int1` and `int2` → `str1str2`
- multiples of `int1` only → `str1`
- multiples of `int2` only → `str2`

`HEAD` is also accepted. Other methods return `405 Method Not Allowed` with an `Allow: GET, HEAD` header.

#### Errors

An invalid request returns `400 Bad Request` with a JSON body. The body names the parameter and the reason, and reports the first error found:

```console
$ curl -s "localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz"
{"error":"invalid parameter","field":"str2","reason":"required"}

$ curl -s "localhost:8080/fizzbuzz?int1=x&int2=5&limit=15&str1=a&str2=b"
{"error":"invalid parameter","field":"int1","reason":"must be an integer"}

$ curl -s "localhost:8080/fizzbuzz?int1=0&int2=5&limit=15&str1=a&str2=b"
{"error":"invalid parameter","field":"int1","reason":"must be bigger than zero"}

$ curl -s "localhost:8080/fizzbuzz?int1=3&int2=5&limit=5000&str1=a&str2=b"
{"error":"invalid parameter","field":"limit","reason":"must not exceed max value 1024"}
```

Any other failure returns `500` with `{"error":"internal error"}`. Details are logged, never sent to the client.

### `GET /healthz`

Returns `200 OK` with an empty body. Kubernetes can use it for both liveness and readiness probes: the service has no dependencies, so "alive" and "ready" mean the same thing.

### `GET /statistics`

Not implemented yet; planned for the next release. See [Limitations and next steps](#limitations-and-next-steps).

## Design decisions

To keep the server production-ready and easy to maintain, I made the following choices:

- **Standard library only.** `net/http` routing (method patterns, Go 1.22+), `encoding/json` and `log/slog` cover everything the brief needs. There are no runtime dependencies to audit or upgrade.
- **No premature optimization.** Correctness and bounded resources come first.
- **JSON array output.** Quoting separates values without ambiguity: `str1=a,b` can't be read as two elements. HTML-sensitive characters are escaped.
- **Strict input.** Every parameter is required. Values that are missing, not integers, or out of range are rejected with a `400` that names the field. There are no silent defaults.
- **Bounded input.** By default, `limit ≤ 1024` and `str1`/`str2` ≤ 64 bytes, to keep response size under control and avoid a denial-of-service risk.
  - The limits apply to input bytes. JSON escaping can expand one byte up to 6× (`<` → `<`; control characters → `\u0001`), so the worst-case response is about 790 KB.
  - I measured it at 789,506 bytes, with `int1=1&int2=1&limit=1024` and 64 × `%01` in both strings.
  - The limits belong to `fizzbuzz.Generator` and can't exceed hard ceilings: 100,000 elements and 255 bytes.
- **The domain is separate from HTTP.**
  - `internal/fizzbuzz` holds the rules: parsing (`Params.Parse`), validation, and generation. It depends on a minimal `QueryValues` interface (`url.Values` satisfies it), not on `net/http`.
  - `internal/api` translates between HTTP and the domain. It depends on a small `Generator` interface, so tests can replace it with a fake.
- **Typed errors.** `ParamError{Field, Err}` wraps sentinel errors. The API maps it to a `400` with `errors.As`, and tests match reasons with `errors.Is`.
- **A hardened HTTP server.** It sets read, header, write and idle timeouts, limits headers to 16 KiB, and shuts down gracefully with a timeout.
- **Minimal container.** A multi-stage build produces a static binary on distroless `nonroot`, with OCI labels for version and revision.

## Project layout

```text
cmd/server/          entry point: HTTP server, signals, graceful shutdown
internal/fizzbuzz/   domain: Params, parsing, validation, Generator, errors
internal/api/        HTTP layer: routes, handlers, JSON responses
scripts/loadtest.sh  load test with hey (see Performance)
.github/workflows/   CI: lint, test, vulncheck, Docker smoke test, publish on tags
```

## Development

```console
$ make help
  help     Show available targets
  build    Build the server binary into bin/
  run      Run the server locally
  test     Run all tests with the race detector
  cover    Run tests with a coverage summary
  fuzz     Fuzz each target for FUZZTIME (default 30s)
  lint     go vet + golangci-lint
  fmt      Format the code
  tidy     go mod tidy + verify
  docker   Build the Docker image
  clean    Remove build artefacts
```

- **Tests** are table-driven and use only the standard `testing` package. They cover parsing, validation, generation, routing (`404`, `405`, `HEAD`), error bodies, and a `500` path that must not leak the internal error.
- **Fuzzing:**
  - `FuzzGenerate` checks that the output has `limit` elements, each one correct, or that the error is a `ParamError`.
  - `FuzzParseParams` checks that any query string yields only the documented parse errors.
  - Run `make fuzz FUZZTIME=10s`.
- **Linting:** `golangci-lint` v2, configured in `.golangci.yml`. It also requires doc comments on exported identifiers.
- **CI (GitHub Actions)** runs on every push and pull request:
  - lint
  - tests on the two supported Go releases, with coverage
  - `govulncheck`
  - a Docker build, then a `curl` smoke test against the running container
- **Releases:** tags `v*` publish the image to GHCR. Dependabot keeps the GitHub Actions up to date.

## Performance

These are indicative numbers from a laptop, not a benchmark. I ran `scripts/loadtest.sh` ([hey](https://github.com/rakyll/hey), 50 connections, 30 s per scenario) against the released image:

```console
$ docker run --rm --cpus=2 -p 8080:8080 ghcr.io/peczenyj/go-fizzbuzz:v0.1.0
$ scripts/loadtest.sh
```

Setup: v0.1.0, server limited to **2 CPUs**; `hey` ran on the same machine (Intel i7-1185G7, 8 CPUs) using the remaining cores.

| Scenario | Req/s | p50 | p95 | p99 | Status |
|---|---:|---:|---:|---:|---|
| healthz | 62695 | 0.7 ms | 1.5 ms | 2.0 ms | 200 |
| classic 1..100 | 35329 | 1.3 ms | 3.0 ms | 4.0 ms | 200 |
| max limit 1024 | 14516 | 3.1 ms | 7.7 ms | 9.9 ms | 200 |
| worst case (~790 KB) | 1504 | 32.5 ms | 51.9 ms | 74.9 ms | 200 |
| invalid (400) | 41268 | 1.1 ms | 2.5 ms | 3.3 ms | 400 |

- **Normal requests:** sub-millisecond medians, and p99 under 10 ms even at `limit=1024`.
- **Invalid requests** are cheaper than valid ones, because they are rejected before any generation.
- **Worst case:** about 1.2 GB/s of JSON from 2 CPUs, with p99 at 75 ms. Response size, not request count, drives the cost. That is why the input limits matter: without them, a few such requests could saturate the service.
- `hey` computes percentiles over at most 1,000,000 responses. Where a run went past that, the percentiles cover its first million responses.

## Limitations and next steps

- **`/statistics`** (the bonus) is next. The plan is a generic, concurrency-safe counter that tracks the most frequent key in O(1), and records only successful requests.
  - The counts will be in memory, so they are lost on restart and wrong behind several replicas, since each instance counts only its own traffic. A shared store would fix both, for example Redis `ZINCRBY` / `ZREVRANGE … WITHSCORES`.
  - Every distinct request adds a key. To bound memory, cap the number of keys or use a heavy-hitters algorithm (Space-Saving, Misra–Gries).
- **Configuration.** The listen address and limits are constants today. Next: environment variables for the address, log level and format, and the `limit`/string maximums, validated at startup.
- **Observability.** Structured logs with `slog`, but no metrics yet. Next: Prometheus metrics for request count, latency and response size per status, served on a separate port.
- **Streaming.** The response is built in memory. That is fine with the current limits (under 1 MB). Much larger limits would call for streaming the JSON array instead.
- **Rate limiting.** I left it out of the service on purpose; I'd expect it at the ingress or API gateway.

## How I worked

I wrote the code myself, and used an AI assistant for code reviews.
