# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/peczenyj/go-fizzbuzz/releases/tag/v0.1.0
