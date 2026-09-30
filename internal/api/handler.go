package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

var (
	_ fizzbuzz.QueryValues = url.Values{}
	_ Generator            = (*fizzbuzz.Generator)(nil)
	_ Statistics           = (*stats.Counter[fizzbuzz.Params])(nil)
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

// Statistics records successful requests and reports the most frequent one.
type Statistics interface {
	Record(p fizzbuzz.Params)
	Top() (p fizzbuzz.Params, hits int, ok bool)
}

// API is the HTTP interface of the service. It routes requests itself, so
// tests exercise the same routing as production.
type API struct {
	mux        *http.ServeMux
	generator  Generator
	statistics Statistics
}

// New returns an API serving /healthz, /fizzbuzz and /statistics with the given arguments.
// It panics if gen or st is nil: that is a wiring bug, not a runtime condition.
func New(gen Generator, st Statistics) *API {
	if gen == nil {
		panic("api.New: nil Generator")
	}

	if st == nil {
		panic("api.New: nil Statistics")
	}

	a := &API{mux: http.NewServeMux(), generator: gen, statistics: st}

	a.mux.HandleFunc("GET /healthz", a.handleHealthz)
	a.mux.HandleFunc("GET /fizzbuzz", a.handleFizzBuzz)
	a.mux.HandleFunc("GET /statistics", a.handleStatistics)

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

	a.statistics.Record(params)

	writeJSON(w, http.StatusOK, result)
}

// ParamsBody is the JSON form of a fizzbuzz request.
type ParamsBody struct {
	Int1  int    `json:"int1"`
	Int2  int    `json:"int2"`
	Limit int    `json:"limit"`
	Str1  string `json:"str1"`
	Str2  string `json:"str2"`
}

// StatisticsBody is the JSON body of GET /statistics; Params is null until
// a request succeeded.
type StatisticsBody struct {
	Params *ParamsBody `json:"params"`
	Hits   int         `json:"hits"`
}

func (a *API) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		writeJSON(w, http.StatusBadRequest, &ErrorBody{Error: "unexpected parameter"})

		return
	}

	var body StatisticsBody
	if p, hits, ok := a.statistics.Top(); ok {
		body.Params = &ParamsBody{Int1: p.Int1, Int2: p.Int2, Limit: p.Limit, Str1: p.Str1, Str2: p.Str2}
		body.Hits = hits
	}

	writeJSON(w, http.StatusOK, &body)
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
