package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

var _ fizzbuzz.QueryValues = url.Values{}

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
	params, err := fizzbuzz.ParseParams(r.URL.Query())
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
