package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

var (
	_ fizzbuzz.QueryValues = url.Values{}
	_ Generator            = (*fizzbuzz.Generator)(nil)
	_ http.Handler         = (*API)(nil)
)

const (
	contentTypeHeader = `Content-Type`
	mediaTypeJSON     = `application/json`
)

// Generator produces a fizzbuzz sequence; *fizzbuzz.Generator implements it.
type Generator interface {
	Generate(p fizzbuzz.Params) ([]string, error)
}

// API is the HTTP interface of the service. It routes requests itself, so
// tests exercise the same routing as production.
type API struct {
	mux       *http.ServeMux
	generator Generator
}

// New returns an API serving /healthz and /fizzbuzz with gen.
// It panics if gen is nil: that is a wiring bug, not a runtime condition.
func New(gen Generator) *API {
	if gen == nil {
		panic("api.New: nil Generator")
	}
	a := &API{mux: http.NewServeMux(), generator: gen}
	a.mux.HandleFunc("GET /healthz", a.handleHealthz)
	a.mux.HandleFunc("GET /fizzbuzz", a.handleFizzBuzz)
	return a
}

// ServeHTTP implements http.Handler.
func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) { a.mux.ServeHTTP(w, r) }

func (a *API) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (a *API) handleFizzBuzz(w http.ResponseWriter, r *http.Request) {
	var params fizzbuzz.Params

	err := params.Parse(r.URL.Query())
	if err != nil {
		writeError(w, err)

		return
	}

	slog.Debug("will generate fizzbuzz sequence with parameters", slog.Any("params", params))

	result, err := a.generator.Generate(params)
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
