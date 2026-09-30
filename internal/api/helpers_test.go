package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func doRequest(t *testing.T,
	handler http.Handler,
	method, target string,
) (
	statusCode int,
	body []byte,
	headers http.Header,
) {
	t.Helper()

	request := httptest.NewRequest(method, target, nil)

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, request)

	response := w.Result()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error while read response body: %v", err)
	}

	return response.StatusCode, body, response.Header
}

// decodeJSON decodes body into a T, failing the test if it is not valid JSON.
func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()

	var v T

	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("unexpected error while decoding %q: %v", body, err)
	}

	return v
}

var _ api.Generator = generatorFunc(nil)

// generatorFunc is an api.Generator stub.
type generatorFunc func(fizzbuzz.Params) ([]string, error)

func (f generatorFunc) Generate(p fizzbuzz.Params) ([]string, error) { return f(p) }

var _ api.Statistics = (*statisticsMock)(nil)

type topResponse struct {
	p    fizzbuzz.Params
	hits int
	ok   bool
}

// statisticsMock records the params it receives and returns a canned Top.
// Its zero value records and reports nothing.
type statisticsMock struct {
	records     []fizzbuzz.Params
	topResponse topResponse
}

func (m *statisticsMock) Record(p fizzbuzz.Params) {
	m.records = append(m.records, p)
}

func (m *statisticsMock) Top() (p fizzbuzz.Params, hits int, ok bool) {
	return m.topResponse.p, m.topResponse.hits, m.topResponse.ok
}

func returnTopStatistics(p fizzbuzz.Params, hits int, ok bool) api.Statistics {
	return &statisticsMock{
		topResponse: topResponse{p: p, hits: hits, ok: ok},
	}
}

// expectRecords returns a statisticsMock builder that checks, when the test
// ends, that exactly want was recorded (nothing, if want is empty).
func expectRecords(want ...fizzbuzz.Params) func(*testing.T) api.Statistics {
	return func(t *testing.T) api.Statistics {
		t.Helper()

		st := &statisticsMock{}

		t.Cleanup(func() {
			if !slices.Equal(st.records, want) {
				t.Errorf("unexpected statistics records (got: %v, expected: %v)", st.records, want)
			}
		})

		return st
	}
}
