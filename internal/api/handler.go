package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

const (
	ContentTypeHeaderName      = `Content-Type`
	ContentTypeApplicationJSON = `application/json`
)

// Healthz k8s api health endpoint.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// FizzBuzz is the http fizbuzz handler.
func FizzBuzz(w http.ResponseWriter, r *http.Request) {
	params, err := parseParams(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: unable to parse query string: %v", err)

		return
	}

	result, err := fizzbuzz.Generate(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %v", err)

		return
	}

	w.Header().Set(ContentTypeHeaderName, ContentTypeApplicationJSON)

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		slog.Warn("unexpected error while perform json encode", slog.Any("error", err))
	}
}

func parseParams(query url.Values) (params fizzbuzz.Params, err error) {
	params = fizzbuzz.DefaultParams()

	if query.Has("int1") {
		params.Int1, err = strconv.Atoi(query.Get("int1"))
		if err != nil {
			return params, fmt.Errorf("unable to convert param 'int1' to integer: %w", err)
		}
	}

	if query.Has("int2") {
		params.Int2, err = strconv.Atoi(query.Get("int2"))
		if err != nil {
			return params, fmt.Errorf("unable to convert param 'int2' to integer: %w", err)
		}
	}

	if query.Has("limit") {
		params.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			return params, fmt.Errorf("unable to convert param 'limit' to integer: %w", err)
		}
	}

	if query.Has("str1") {
		params.Str1 = query.Get("str1")
	}

	if query.Has("str2") {
		params.Str2 = query.Get("str2")
	}

	return params, nil
}
