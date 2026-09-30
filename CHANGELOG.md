# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

## [0.3.1] - 2026-09-30

### Added

- OpenAPI 3.1 description of the API in `api/openapi.yaml`, linted in CI
  with Redocly CLI (`make openapi`).

### Fixed

- Invalid `FIZZBUZZ_*` environment variables are all reported at startup,
  together with an invalid flag, unexpected arguments or invalid flag values,
  instead of one error per run. The `-max-limit` and `-max-str-length` bounds
  are still checked afterwards, by `fizzbuzz.NewGenerator`, one at a time.
- A `FIZZBUZZ_*` variable is checked on its own, like its flag: an invalid
  address, log format or shutdown timeout is rejected even when a flag
  replaces it, and no longer shows up as the flag's default in `-h`.
- `FIZZBUZZ_MAX_LIMIT` and `FIZZBUZZ_MAX_STR_LENGTH` accept the same integer
  syntax as their flags (`0x400`, `1_000`, and `010` as octal), instead of
  base 10 only.
- An invalid port in `-addr` or `FIZZBUZZ_ADDR`, such as `:99999`, is rejected
  at startup with exit code 2, instead of failing to listen with exit code 1.

## [0.3.0] - 2026-09-30

### Added

- Configuration with flags and `FIZZBUZZ_*` environment variables: listen
  address, log level and format (text or JSON), `limit` and string maximums,
  shutdown timeout. Invalid values stop the server at startup with exit code 2.
- `-version` prints the version and revision, and exits.

## [0.2.0] - 2026-09-30

### Added

- `GET /statistics`: the most frequent successful `/fizzbuzz` request and its
  number of hits. Only successful `GET` and `HEAD` requests count; parameters are
  compared after decoding; ties go to the first request to reach the count.
- `internal/stats`: generic concurrency-safe counter with O(1) top.
- Go benchmarks (`make bench`) for generation, the handler and the counter, with results in the README.
- `scripts/loadtest.sh` and load-test results for v0.1.0 in the README.

## [0.1.0] - 2026-09-30

### Added

- `GET /fizzbuzz`: configurable fizz-buzz as a JSON array of strings; all five
  parameters required, bounded input (`limit ≤ 1024`, strings ≤ 64 bytes).
- JSON `400` errors naming the invalid field and reason; `500` without leaking details.
- `GET /healthz` for liveness and readiness probes.
- HTTP server timeouts, 16 KiB header limit, graceful shutdown on SIGINT/SIGTERM.
- Table-driven tests, fuzz targets for parsing and generation.
- Makefile, golangci-lint v2, govulncheck, GitHub Actions CI with Docker smoke test.
- Distroless Docker image, published to GHCR on version tags.

[Unreleased]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.3.1...HEAD
[0.3.1]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/peczenyj/go-fizzbuzz/releases/tag/v0.1.0
