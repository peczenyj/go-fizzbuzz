package api

import (
	"encoding/json"
	"errors"
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

// FizzBuzz is the http fizzbuzz handler.
func FizzBuzz(w http.ResponseWriter, r *http.Request) {
	params, err := parseParams(r.URL.Query())
	if err != nil {
		writeError(w, err)

		return
	}

	result, err := fizzbuzz.Generate(params)
	if err != nil {
		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, result)
}

// ErrNotAnInteger guard error.
var ErrNotAnInteger = errors.New(`must be an integer`)

func parseParams(query url.Values) (params fizzbuzz.Params, err error) {
	params = fizzbuzz.DefaultParams()

	if query.Has("int1") {
		params.Int1, err = strconv.Atoi(query.Get("int1"))
		if err != nil {
			return params, &fizzbuzz.ParamError{Field: fizzbuzz.FieldInt1, Err: ErrNotAnInteger}
		}
	}

	if query.Has("int2") {
		params.Int2, err = strconv.Atoi(query.Get("int2"))
		if err != nil {
			return params, &fizzbuzz.ParamError{Field: fizzbuzz.FieldInt2, Err: ErrNotAnInteger}
		}
	}

	if query.Has("limit") {
		params.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			return params, &fizzbuzz.ParamError{Field: fizzbuzz.FieldLimit, Err: ErrNotAnInteger}
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

func writeJSON(w http.ResponseWriter, statusCode int, result any) {
	w.Header().Set(ContentTypeHeaderName, ContentTypeApplicationJSON)

	w.WriteHeader(statusCode)

	err := json.NewEncoder(w).Encode(result)
	if err != nil {
		slog.Warn("unexpected error while perform json encode on result",
			slog.Any("error", err))
	}
}

// ErrorBody error body envelope.
type ErrorBody struct {
	Error  string `json:"error"`
	Field  string `json:"field,omitempty"`
	Reason string `json:"reason,omitempty"`
}

func writeError(w http.ResponseWriter, err error) {
	if pe, ok := errors.AsType[*fizzbuzz.ParamError](err); ok {
		writeJSON(w, http.StatusBadRequest, &ErrorBody{
			Error:  `invalid parameter`,
			Field:  string(pe.Field),
			Reason: pe.Err.Error(),
		})

		return
	}

	slog.Error("unexpected error", slog.Any("error", err))

	writeJSON(w, http.StatusInternalServerError, &ErrorBody{Error: `internal error`})
}
