# Changelog

All notable changes to this project are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

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

[Unreleased]: https://github.com/peczenyj/go-fizzbuzz/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/peczenyj/go-fizzbuzz/releases/tag/v0.1.0
