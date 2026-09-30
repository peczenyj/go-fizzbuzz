ARG GO_VERSION=1.27
FROM golang:${GO_VERSION} AS build

WORKDIR /src

# Dependencies first: this layer is cached until go.mod/go.sum change.
COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

ARG VERSION=dev
ARG REVISION=unknown

# Static, reproducible binary: no cgo, no local paths, no debug symbols.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X main.version=${VERSION} -X main.revision=${REVISION}" \
      -o /out/fizzbuzz \
      ./cmd/server

# ---- runtime stage ----------------------------------------------------------
# distroless/static: no shell, no package manager, runs as uid 65532.
FROM gcr.io/distroless/static-debian12:nonroot

ARG VERSION=dev
ARG REVISION=unknown
ARG CREATED
LABEL org.opencontainers.image.title="go-fizzbuzz" \
      org.opencontainers.image.description="Basic fizz-buzz REST API" \
      org.opencontainers.image.source="https://github.com/peczenyj/go-fizzbuzz" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${REVISION}" \
      org.opencontainers.image.created="${CREATED}"

COPY --from=build /out/fizzbuzz /fizzbuzz

USER nonroot:nonroot
EXPOSE 8080

ENTRYPOINT ["/fizzbuzz"]
