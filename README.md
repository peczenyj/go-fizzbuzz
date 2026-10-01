# go-fizzbuzz

[![Latest release](https://img.shields.io/github/release/peczenyj/go-fizzbuzz.svg)](https://github.com/peczenyj/go-fizzbuzz/releases/latest)
[![Container package](https://img.shields.io/badge/container-ghcr.io%2Fgo--fizzbuzz-blue)](https://github.com/peczenyj/go-fizzbuzz/pkgs/container/go-fizzbuzz)
[![CI](https://github.com/peczenyj/go-fizzbuzz/actions/workflows/ci.yml/badge.svg)](https://github.com/peczenyj/go-fizzbuzz/actions/workflows/ci.yml)
[![CodeQL](https://github.com/peczenyj/go-fizzbuzz/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/peczenyj/go-fizzbuzz/actions/workflows/github-code-scanning/codeql)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.26-%23007d9c)
[![codecov](https://codecov.io/gh/peczenyj/go-fizzbuzz/graph/badge.svg)](https://codecov.io/gh/peczenyj/go-fizzbuzz)

A configurable fizz-buzz REST API written in Go, using only the standard library.

```console
$ curl -s "localhost:8080/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz"
["1","2","fizz","4","buzz","fizz","7","8","fizz","buzz","11","fizz","13","14","fizzbuzz"]
```

- [Objective](#objective)
- [Quick start](#quick-start)
- [Configuration](#configuration)
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
- Returns a list of strings with numbers from 1 to limit, where: all multiples of int1 are replaced by str1, all multiples of int2 are replaced by str2, all multiples of int1 and int2 are replaced by [...]
 
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
... level=INFO msg="application start" version=dev ...
... level=INFO msg="starting server, will listen and serve" addr=:8080
```

With Docker:

```console
$ docker build -t go-fizzbuzz .
$ docker run --rm -p 8080:8080 go-fizzbuzz
```

`make docker` builds the same image, tagged with `git describe` and carrying version labels.

The server listens on `:8080` by default and stops gracefully on `SIGINT` or `SIGTERM`.

## Configuration

Every setting is a command-line flag whose default comes from an environment variable, so the precedence is **flag > environment variable > built-in default**. Flags suit local use; environment va[...]

| Flag | Environment variable | Default | Accepted values |
|---|---|---|---|
| `-addr` | `FIZZBUZZ_ADDR` | `:8080` | `host:port` |
| `-log-level` | `FIZZBUZZ_LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `-log-format` | `FIZZBUZZ_LOG_FORMAT` | `text` | `text`, `json` |
| `-max-limit` | `FIZZBUZZ_MAX_LIMIT` | `1024` | 1 to 100,000 |
| `-max-str-length` | `FIZZBUZZ_MAX_STR_LENGTH` | `64` | 1 to 255 bytes |
| `-shutdown-timeout` | `FIZZBUZZ_SHUTDOWN_TIMEOUT` | `10s` | a positive Go duration |
| `-version` | | | prints the version and exits |

```console
$ go run ./cmd/server -log-level=debug -max-limit=100
$ docker run --rm -p 8080:8080 -e FIZZBUZZ_LOG_FORMAT=json ghcr.io/peczenyj/go-fizzbuzz:latest
$ docker run --rm ghcr.io/peczenyj/go-fizzbuzz:latest -version
go-fizzbuzz v0.3.2 (revision …)
```

- **Invalid values stop the server at startup** with exit code 2 and a message naming the flag or variable, instead of falling back to a default. Each environment variable is checked like its flag, ev[...]
- **The limits are validated by `fizzbuzz.NewGenerator`**, which owns the hard ceilings, so the configuration doesn't duplicate the domain rules.
- `-h` lists every flag with its environment variable and default:

```console
$ go run ./cmd/server -h
Usage of fizzbuzz:
  -addr string
        listen address (env FIZZBUZZ_ADDR) (default ":8080")
  -log-format string
        log format: text or json (env FIZZBUZZ_LOG_FORMAT) (default "text")
  -log-level value
        minimum log level: debug, info, warn or error (env FIZZBUZZ_LOG_LEVEL) (default INFO)
  -max-limit int
        largest accepted limit, up to 100000 (env FIZZBUZZ_MAX_LIMIT) (default 1024)
  -max-str-length int
        largest accepted str1/str2 in bytes, up to 255 (env FIZZBUZZ_MAX_STR_LENGTH) (default 64)
  -shutdown-timeout duration
        grace period for in-flight requests on shutdown (env FIZZBUZZ_SHUTDOWN_TIMEOUT) (default 10s)
  -version
        print the version and exit
```

## API

The API is described in [`api/openapi.yaml`](api/openapi.yaml) (OpenAPI 3.1), linted in CI. Paste it into [Swagger Editor](https://editor.swagger.io/) or any OpenAPI viewer to browse it.

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

Returns the most frequent successful `/fizzbuzz` request and its number of hits.

```console
$ curl -s localhost:8080/statistics
{"params":{"int1":3,"int2":5,"limit":15,"str1":"fizz","str2":"buzz"},"hits":2}
```

Before any successful request, it returns `{"params":null,"hits":0}`: same status, same shape, so clients always parse one format.

What counts as a hit:

- **Only successful requests.** A request rejected with `400`, failing with `500`, or whose response couldn't be delivered is not counted.
- **`GET` and `HEAD`.** A `HEAD /fizzbuzz` does the same work as a `GET`, without the body, so it counts as a hit too.
- **The same request, however it is written.** Parameters are compared after decoding, so `int1=3&int2=5&…` and `int2=5&int1=3&…` are the same request, and so are `str1=fizz` and `str1=%66izz`.
- **Ties:** when two requests have the same number of hits, the first one to reach that count is reported.

The endpoint accepts no parameter: any query string returns `400` with `{"error":"unexpected parameter"}`. `HEAD` is accepted, and other methods return `405`.

## Design decisions

To keep the server production-ready and easy to maintain, I made the following choices:

- **Standard library only.** `net/http` routing (method patterns, Go 1.22+), `encoding/json` and `log/slog` cover everything the brief needs. There are no runtime dependencies to audit or upgrade[...]
- **Configuration with the standard `flag` package.** Environment variables provide the flag defaults, which gives the flag > env > default precedence without a configuration library. `config.Parse` t[...]
- **No premature optimization.** Correctness and bounded resources come first.
- **JSON array output.** Quoting separates values without ambiguity: `str1=a,b` can't be read as two elements. HTML-sensitive characters are escaped.
- **Strict input.** Every parameter is required. Values that are missing, not integers, or out of range are rejected with a `400` that names the field. There are no silent defaults.
- **Bounded input.** By default, `limit ≤ 1024` and `str1`/`str2` ≤ 64 bytes, to keep response size under control and avoid a denial-of-service risk.
  - The limits apply to input bytes. JSON escaping can expand one byte up to 6× (`<` → `\u003c`; control characters → `\u0001`), so the worst-case response is about 790 KB.
  - I measured it at 789,506 bytes, with `int1=1&int2=1&limit=1024` and 64 × `%01` in both strings.
  - The limits belong to `fizzbuzz.Generator` and can't exceed hard ceilings: 100,000 elements and 255 bytes.
- **The domain is separate from HTTP.**
  - `internal/fizzbuzz` holds the rules: parsing (`Params.Parse`), validation, and generation. It depends on a minimal `QueryValues` interface (`url.Values` satisfies it), not on `net/http`.
  - `internal/api` translates between HTTP and the domain. It depends on two small interfaces, `Generator` and `Statistics`, defined where they are used, so tests can replace them with fakes.
- **Statistics as a generic counter.** `internal/stats` provides `Counter[K comparable]`, a mutex-protected map that knows nothing about fizzbuzz. It is used with `fizzbuzz.Params` as the key, which i[...]
  - Counts only ever increase, so the top can change only on the key just incremented. `Record` and `Top` are O(1): there is no scan and no sort.
  - Only successful requests are recorded, after the response is written, so invalid requests can't make themselves the most frequent, and a response that never reached the client doesn't count.
- **Typed errors.** `ParamError{Field, Err}` wraps sentinel errors. The API maps it to a `400` with `errors.As`, and tests match reasons with `errors.Is`.
- **A hardened HTTP server.** It sets read, header, write and idle timeouts, limits headers to 16 KiB, and shuts down gracefully with a timeout.
- **Minimal container.** A multi-stage build produces a static binary on distroless `nonroot`, with OCI labels for version and revision.
  - Why a container: it's the unit Kubernetes and most platforms deploy, so the same image tested by CI's smoke test runs locally and in production, with `/healthz` ready for the probes.

## Project layout

```text
cmd/server/          entry point: HTTP server, signals, graceful shutdown
internal/config/     flags and FIZZBUZZ_* environment variables
internal/fizzbuzz/   domain: Params, parsing, validation, Generator, errors
internal/api/        HTTP layer: routes, handlers, JSON responses
internal/stats/      generic concurrency-safe counter with O(1) top
api/openapi.yaml     OpenAPI 3.1 description of the API
scripts/loadtest.sh  load test with hey (see Performance)
.github/workflows/   CI: lint, test, vulncheck, Docker smoke test, publish on tags
```

## Development

```console
$ make help
  help     Show available targets
  build    Build the server binary into bin/
  run      Run the server locally (flags via ARGS, e.g. ARGS=-log-level=debug)
  test     Run all tests with the race detector
  cover    Run tests with a coverage summary
  fuzz     Fuzz each target for FUZZTIME (default 30s)
  bench    Run benchmarks with allocation stats
  lint     go vet + golangci-lint
  openapi  Lint api/openapi.yaml (needs npx)
  fmt      Format the code
  tidy     go mod tidy + verify
  docker   Build the Docker image
  clean    Remove build artefacts
```

- **Tests** are table-driven and use only the standard `testing` package. They cover parsing, validation, generation, routing (`404`, `405`, `HEAD`), error bodies, and a `500` path that must not [...]
- **Fuzzing:**
  - `FuzzGenerate` checks that the output has `limit` elements, each one correct, or that the error is a `ParamError`.
  - `FuzzParseParams` checks that any query string yields only the documented parse errors.
  - Run `make fuzz FUZZTIME=10s`.
- **Benchmarks:** `make bench` runs benchmarks for generation, the full handler and the statistics counter, with allocation stats. Results are under [Performance](#performance).
- **Linting:** `golangci-lint` v2, configured in `.golangci.yml`. It also requires doc comments on exported identifiers.
- **CI (GitHub Actions)** runs on pushes to `main` and `devel`, on version tags, and on pull requests:
  - lint
  - tests on the two supported Go releases, with coverage
  - `govulncheck`
  - each benchmark once, so they keep compiling and running (numbers from shared runners are too noisy to compare)
  - a Docker build, then a `curl` smoke test against the running container
- **Releases:** tags `v*` publish the image to GHCR. Dependabot keeps the GitHub Actions up to date.

## Performance

These are indicative numbers from a laptop, not a benchmark. I ran `scripts/loadtest.sh` ([hey](https://github.com/rakyll/hey), 50 connections, 30 s per scenario) against the released image:

```console
$ docker run --rm --cpus=2 -p 8080:8080 ghcr.io/peczenyj/go-fizzbuzz:v0.2.0
$ scripts/loadtest.sh
```

Setup: v0.2.0, server limited to **2 CPUs**; `hey` ran on the same machine (Intel i7-1185G7, 8 CPUs) using the remaining cores.

| Scenario | Req/s | p50 | p95 | p99 | Status | Req/s in v0.1.0 |
|---|---:|---:|---:|---:|---|---:|
| healthz | 60060 | 0.6 ms | 1.7 ms | 2.5 ms | 200 | 62695 |
| classic 1..100 | 38964 | 1.2 ms | 2.7 ms | 3.4 ms | 200 | 35329 |
| max limit 1024 | 15542 | 2.9 ms | 7.3 ms | 9.0 ms | 200 | 14516 |
| worst case (~790 KB) | 1495 | 32.7 ms | 51.8 ms | 73.4 ms | 200 | 1504 |
| invalid (400) | 41352 | 1.1 ms | 2.5 ms | 3.3 ms | 400 | 41268 |
| statistics | 45851 | 1.0 ms | 2.2 ms | 2.9 ms | 200 | — |

- **Normal requests:** sub-millisecond medians, and p99 under 10 ms even at `limit=1024`.
- **Invalid requests** are cheaper than valid ones, because they are rejected before any generation.
- **Worst case:** about 1.2 GB/s of JSON from 2 CPUs, with p99 at 73 ms. Response size, not request count, drives the cost. That is why the input limits matter: without them, a few such requests could[...]
- **Recording statistics has no measurable cost.** Every scenario is within run-to-run variation of v0.1.0, which had no statistics (both directions: `healthz` is 4% slower, `classic` 10% faster). `/s[...]
- `hey` computes percentiles over at most 1,000,000 responses. Where a run went past that, the percentiles cover its first million responses.

### Benchmarks

`make bench` isolates the code from the HTTP stack and the network. These are medians of 5 runs (`-count=5`) on the same laptop, Go 1.27, so again indicative only.

| Benchmark | Time/op | Memory/op | Allocs/op |
|---|---:|---:|---:|
| `Generate` classic 1..100 | 0.70 µs | 1.8 KB | 2 |
| `Generate` limit 1024 | 11.8 µs | 20 KB | 496 |
| `Generate` limit 1024, 64-byte labels | 6.5 µs | 18.6 KB | 2 |
| handler classic 1..100 | 5.0 µs | 2.5 KB | 14 |
| handler limit 1024 | 42.5 µs | 20.8 KB | 508 |
| handler worst case (~790 KB) | 815 µs | 20–32 KB | 16 |
| `Counter.Record`, 1 goroutine | 30 ns | 0 | 0 |
| `Counter.Record`, 8 goroutines | 88 ns | 0 | 0 |

- **JSON encoding dominates.** For the classic request, generation is about 15% of the handler time; the rest is routing, parsing and, mostly, encoding. In the worst case, generation is under 1%: esca[...]
- **Allocations come from numbers, not labels.** `strconv.Itoa` returns cached strings below 100, so allocations appear only for printed numbers from 100 up: 494 of them at `limit=1024`, plus the slic[...]
- **The lock is not a bottleneck.** `Counter.Record` never allocates. Under 8-way contention, it still handles about 11 million records per second (1 / 88 ns), nearly 300× the measured `/fizzbuz[...]

## Limitations and next steps

- **Statistics are in memory and per instance.** They are lost on restart and wrong behind several replicas, since each instance counts only its own traffic. A shared store would fix both, for example[...]
- **Statistics memory is unbounded.** Every distinct successful request adds a key: up to a few hundred bytes with the current input limits, and never removed. A client sending many distinct requests [...]
- **One lock for all requests.** Every successful request takes the counter's mutex. The benchmark shows about 11 million records per second under 8-way contention, far above the request rate, so this[...]
- **Statistics can't be disabled.** Next: a `-statistics=false` flag that removes the `/statistics` route and stops recording, rather than serving an empty result.
- **No configuration file.** Flags and environment variables cover the current settings. A file (YAML or TOML) would only be worth it with many more settings, or with settings that change without a re[...]
- **Observability.** Structured logs with `slog`, but no metrics yet. Next: Prometheus metrics for request count, latency and response size per status, served on a separate port.
- **Streaming.** The response is built in memory. That is fine with the current limits (under 1 MB). Much larger limits would call for streaming the JSON array instead.
- **The OpenAPI description is written by hand.** CI checks that it is valid, but no test checks the server's responses against it, so the two can drift apart. A contract test would need a third-party[...]
- **Rate limiting.** I left it out of the service on purpose; I'd expect it at the ingress or API gateway.

## How I worked

I wrote the core myself (API, domain, statistics, tests) and used Claude Code for reviews; the benchmarks and the configuration package were implemented by Claude Code under my direction and reviewed [...]

I started with a simple approach: the default HTTP server, a function generating the fizzbuzz sequence, and implicit defaults when a query parameter was missing. The first problem I identified was the[...]

There is room for improvement, but I decided to wait for feedback instead of implementing everything. There are also several third-party libraries that could reduce the amount of code I wrote, like [v[...]
