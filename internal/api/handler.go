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
	contentTypeHeader = `Content-Type`
	mediaTypeJSON     = `application/json`
)

// Healthz reports liveness and readiness: it always returns 200 with an empty body.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// FizzBuzz serves GET /fizzbuzz: it parses the five query parameters and returns
// the sequence as a JSON array of strings, or a 400 ErrorBody naming the invalid field.
func FizzBuzz(w http.ResponseWriter, r *http.Request) {
	var params fizzbuzz.Params

	err := params.Parse(r.URL.Query())
	if err != nil {
		writeError(w, err)

		return
	}

	generator := fizzbuzz.DefaultGenerator()

	result, err := generator.Generate(params)
	if err != nil {
		writeError(w, err)

		return
	}

	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, statusCode int, result any) {
	w.Header().Set(contentTypeHeader, mediaTypeJSON)

	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)

	encoder.SetEscapeHTML(true)

	err := encoder.Encode(result)
	if err != nil {
		slog.Warn("unexpected error while perform json encode on result",
			slog.Any("error", err))
	}
}

// ErrorBody is the JSON body of every 4xx/5xx response; Field and Reason are
// set only for invalid parameters (400).
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
