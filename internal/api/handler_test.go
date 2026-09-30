package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"slices"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

// TestRequests checks routing only: method and path → status and headers.
// Response content is covered by the per-endpoint tests.
func TestRequests(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label           string
		method          string
		target          string
		statusCode      int
		responseBody    []byte // compared only when not nil
		expectedHeaders map[string]string
	}{
		{
			label:        "should return 200 OK in case of GET on /healthz",
			method:       http.MethodGet,
			target:       "/healthz",
			statusCode:   http.StatusOK,
			responseBody: []byte{},
		},
		{
			label:        "should return 200 OK in case of HEAD on /healthz",
			method:       http.MethodHead,
			target:       "/healthz",
			statusCode:   http.StatusOK,
			responseBody: []byte{},
		},
		{
			label:        "should return 405 Method Not Allowed in case of POST on /healthz",
			method:       http.MethodPost,
			target:       "/healthz",
			statusCode:   http.StatusMethodNotAllowed,
			responseBody: []byte("Method Not Allowed\n"),
			expectedHeaders: map[string]string{
				"Allow": "GET, HEAD",
			},
		},
		{
			label:        "should return 404 Not Found in case of GET on /healthz/",
			method:       http.MethodGet,
			target:       "/healthz/",
			statusCode:   http.StatusNotFound,
			responseBody: []byte("404 page not found\n"),
		},
		{
			label:        "should return 404 Not Found in case of GET on /not-exist",
			method:       http.MethodGet,
			target:       "/not-exist",
			statusCode:   http.StatusNotFound,
			responseBody: []byte("404 page not found\n"),
		},
		{
			label:      "should return 200 OK in case of GET on /fizzbuzz with valid parameters",
			method:     http.MethodGet,
			target:     "/fizzbuzz?int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			statusCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			label:        "should return 405 Method Not Allowed in case of POST on /fizzbuzz",
			method:       http.MethodPost,
			target:       "/fizzbuzz",
			statusCode:   http.StatusMethodNotAllowed,
			responseBody: []byte("Method Not Allowed\n"),
			expectedHeaders: map[string]string{
				"Allow": "GET, HEAD",
			},
		},
		{
			label:      "should return 200 OK in case of GET on /statistics",
			method:     http.MethodGet,
			target:     "/statistics",
			statusCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			// Body not compared: httptest.ResponseRecorder keeps it on HEAD,
			// the real server drops it.
			label:      "should return 200 OK in case of HEAD on /statistics",
			method:     http.MethodHead,
			target:     "/statistics",
			statusCode: http.StatusOK,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			label:      "should return 400 Bad Request in case of GET on /statistics with any parameter",
			method:     http.MethodGet,
			target:     "/statistics?foo=bar",
			statusCode: http.StatusBadRequest,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
		},
		{
			label:        "should return 405 Method Not Allowed in case of POST on /statistics",
			method:       http.MethodPost,
			target:       "/statistics",
			statusCode:   http.StatusMethodNotAllowed,
			responseBody: []byte("Method Not Allowed\n"),
			expectedHeaders: map[string]string{
				"Allow": "GET, HEAD",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			handler := api.New(fizzbuzz.DefaultGenerator(), &statisticsMock{})

			statusCode, responseBody, responseHeaders := doRequest(t, handler, tc.method, tc.target)

			if statusCode != tc.statusCode {
				t.Fatalf("unexpected http status code for %s %s (got: %v, expected: %v)", tc.method, tc.target, statusCode, tc.statusCode)
			}

			if tc.responseBody != nil && !bytes.Equal(responseBody, tc.responseBody) {
				t.Fatalf("unexpected http body (got: %q, expected: %q)", responseBody, tc.responseBody)
			}

			for key, value := range tc.expectedHeaders {
				if responseHeaders.Get(key) != value {
					t.Fatalf("unexpected response header %s (got: %v, expected: %v)", key, responseHeaders.Get(key), value)
				}
			}
		})
	}
}

func TestFizzBuzzHandler(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label  string
		target string

		buildGenerator  func(t *testing.T) api.Generator
		buildStatistics func(t *testing.T) api.Statistics

		errStatusCode int
		errBody       *api.ErrorBody // set for error cases
		expected      []string       // compared only when not nil

		verifyBody func(*testing.T, []byte) // extra checks on the raw body
	}{
		{
			label:  "should return first 100 elements of default fizzbuzz sequence with no arguments",
			target: "/fizzbuzz?int1=3&int2=5&limit=100&str1=fizz&str2=buzz",
			expected: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
				"16", "17", "fizz", "19", "buzz", "fizz", "22", "23", "fizz", "buzz", "26", "fizz", "28", "29", "fizzbuzz",
				"31", "32", "fizz", "34", "buzz", "fizz", "37", "38", "fizz", "buzz", "41", "fizz", "43", "44", "fizzbuzz",
				"46", "47", "fizz", "49", "buzz", "fizz", "52", "53", "fizz", "buzz", "56", "fizz", "58", "59", "fizzbuzz",
				"61", "62", "fizz", "64", "buzz", "fizz", "67", "68", "fizz", "buzz", "71", "fizz", "73", "74", "fizzbuzz",
				"76", "77", "fizz", "79", "buzz", "fizz", "82", "83", "fizz", "buzz", "86", "fizz", "88", "89", "fizzbuzz",
				"91", "92", "fizz", "94", "buzz", "fizz", "97", "98", "fizz", "buzz",
			},
		},
		{
			label:  "should return first 15 elements of default fizzbuzz sequence",
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
			expected: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
			},
		},
		{
			label:  "should return first 15 elements of alternate fizzbuzz sequence",
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=abc&str2=x,z",
			expected: []string{
				"1", "2", "abc", "4", "x,z", "abc", "7", "8", "abc", "x,z", "11", "abc", "13", "14", "abcx,z",
			},
		},
		{
			label:  "should return first 15 elements of unicode fizzbuzz sequence",
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=%C3%A9&str2=%F0%9F%8D%95",
			expected: []string{
				"1", "2", "é", "4", "🍕", "é", "7", "8", "é", "🍕", "11", "é", "13", "14", "é🍕",
			},
		},
		{
			label:  "should return first 15 elements of special html fizzbuzz sequence < and &",
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=%3C&str2=%26",
			expected: []string{
				"1", "2", "<", "4", "&", "<", "7", "8", "<", "&", "11", "<", "13", "14", "<&",
			},
		},
		{
			label:  "should return one element if limit is 1 and record statistics",
			target: "/fizzbuzz?int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			expected: []string{
				"1",
			},
			buildStatistics: expectRecords(fizzbuzz.Params{Int1: 3, Int2: 5, Limit: 1, Str1: "fizz", Str2: "buzz"}),
		},
		{
			label:           "should not record statistics if the request is invalid",
			target:          "/fizzbuzz?int1=0&int2=5&limit=15&str1=fizz&str2=buzz",
			errStatusCode:   http.StatusBadRequest,
			errBody:         &api.ErrorBody{Error: "invalid parameter", Field: "int1", Reason: "must be bigger than zero"},
			buildStatistics: expectRecords(),
		},
		{
			label:  "should return success if limit is equal to max value",
			target: "/fizzbuzz?int1=3&int2=5&limit=1024&str1=fizz&str2=buzz",

			verifyBody: func(t *testing.T, body []byte) {
				t.Helper()

				if got := len(decodeJSON[[]string](t, body)); got != 1024 {
					t.Fatalf("unexpected result length size (got: %v, expected: %v)", got, 1024)
				}
			},
		},
		{
			label:         "should return error with no parameters",
			target:        "/fizzbuzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `required`},
		},
		{
			label:         "should return error if int1 is missing",
			target:        "/fizzbuzz?int2=5&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `required`},
		},
		{
			label:         "should return error if int2 is missing",
			target:        "/fizzbuzz?int1=3&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int2`, Reason: `required`},
		},
		{
			label:         "should return error if limit is missing",
			target:        "/fizzbuzz?int1=3&int2=5&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `required`},
		},
		{
			label:         "should return error if str1 is missing",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str1`, Reason: `required`},
		},
		{
			label:         "should return error if str2 is missing",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str2`, Reason: `required`},
		},
		{
			label:         "should return error if int1 is not a number",
			target:        "/fizzbuzz?int1=lol&int2=5&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `must be an integer`},
		},
		{
			label:         "should return error if int2 is not a number",
			target:        "/fizzbuzz?int1=3&int2=hehe&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int2`, Reason: `must be an integer`},
		},
		{
			label:         "should return error if limit is not a number",
			target:        "/fizzbuzz?int1=3&int2=5&limit=ops&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must be an integer`},
		},
		{
			label:         "should return error if limit is not allowed",
			target:        "/fizzbuzz?int1=3&int2=5&limit=4096&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must not exceed max value 1024`},
		},
		{
			label:         "should return error if int1 is zero",
			target:        "/fizzbuzz?int1=0&int2=5&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `must be bigger than zero`},
		},
		{
			label:         "should return error if int2 is zero",
			target:        "/fizzbuzz?int1=3&int2=0&limit=15&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `int2`, Reason: `must be bigger than zero`},
		},
		{
			label:         "should return error if limit is zero",
			target:        "/fizzbuzz?int1=3&int2=5&limit=0&str1=fizz&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must be bigger than zero`},
		},
		{
			label:         "should return error if str1 is empty",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str1=&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str1`, Reason: `must not be empty string`},
		},
		{
			label:         "should return error if str2 is empty",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str2`, Reason: `must not be empty string`},
		},
		{
			label:         "should return error if str1 is too big",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str1=12345678901234567890123456789012345678901234567890123456789012345&str2=buzz",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str1`, Reason: `must not exceed max string length 64`},
		},
		{
			label:         "should return error if str2 is too big",
			target:        "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=12345678901234567890123456789012345678901234567890123456789012345",
			errStatusCode: http.StatusBadRequest,
			errBody:       &api.ErrorBody{Error: `invalid parameter`, Field: `str2`, Reason: `must not exceed max string length 64`},
		},
		{
			label:  "should return 500 internal server error in case of unexpected error",
			target: "/fizzbuzz?int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			buildGenerator: func(t *testing.T) api.Generator {
				t.Helper()

				var calls int

				t.Cleanup(func() {
					t.Helper()

					if calls != 1 {
						t.Fatalf("unexpected number of calls on fizzbuzz generator (got: %v, expected: %v)", calls, 1)
					}
				})

				return generatorFunc(func(fizzbuzz.Params) ([]string, error) {
					calls++

					return nil, errors.New("ops")
				})
			},
			errStatusCode: http.StatusInternalServerError,
			errBody:       &api.ErrorBody{Error: `internal error`},
			verifyBody: func(t *testing.T, responseBody []byte) {
				t.Helper()

				if bytes.Contains(responseBody, []byte("ops")) {
					t.Fatalf("response body must not leak original inner error")
				}
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			var generator api.Generator = fizzbuzz.DefaultGenerator()

			if tc.buildGenerator != nil {
				generator = tc.buildGenerator(t)
			}

			var statistics api.Statistics = &statisticsMock{}

			if tc.buildStatistics != nil {
				statistics = tc.buildStatistics(t)
			}

			handler := api.New(generator, statistics)

			statusCode, responseBody, responseHeaders := doRequest(t, handler, http.MethodGet, tc.target)

			expectedStatusCode := http.StatusOK
			if tc.errBody != nil {
				expectedStatusCode = tc.errStatusCode
			}

			if statusCode != expectedStatusCode {
				t.Fatalf("unexpected http status code from endpoint /fizzbuzz (got: %v, expected: %v)", statusCode, expectedStatusCode)
			}

			if contentType := responseHeaders.Get(`Content-Type`); contentType != `application/json` {
				t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, `application/json`)
			}

			if tc.verifyBody != nil {
				tc.verifyBody(t, responseBody)
			}

			if tc.errBody != nil {
				if got := decodeJSON[api.ErrorBody](t, responseBody); got != *tc.errBody {
					t.Fatalf("unexpected result (got: %+v, expect: %+v)", got, tc.errBody)
				}

				return
			}

			if tc.expected != nil {
				if got := decodeJSON[[]string](t, responseBody); !slices.Equal(got, tc.expected) {
					t.Fatalf("unexpected result (got: %v, expect: %v)", got, tc.expected)
				}
			}
		})
	}
}

func TestAPIConstructor_panic_on_nil_generator(t *testing.T) {
	t.Parallel()

	defer func() {
		expected := `api.New: nil Generator`

		if got := recover(); got != expected {
			t.Fatalf("recover from panic: (got: %v, expected: %v)", got, expected)
		}
	}()

	_ = api.New(nil, &statisticsMock{})

	t.Fatalf("a panic is expected")
}

func TestAPIConstructor_panic_on_nil_statistics(t *testing.T) {
	t.Parallel()

	defer func() {
		expected := `api.New: nil Statistics`

		if got := recover(); got != expected {
			t.Fatalf("recover from panic: (got: %v, expected: %v)", got, expected)
		}
	}()

	_ = api.New(fizzbuzz.DefaultGenerator(), nil)

	t.Fatalf("a panic is expected")
}

func TestFizzBuzzHandler_passes_params_and_result_through(t *testing.T) {
	t.Parallel()

	var (
		calls int
		got   fizzbuzz.Params
	)

	// Canned output: deliberately not a fizzbuzz sequence, so the test proves
	// the handler writes the generator's result verbatim.
	canned := []string{"a", "b", "c"}

	gen := generatorFunc(func(p fizzbuzz.Params) ([]string, error) {
		calls++
		got = p

		return canned, nil
	})
	handler := api.New(gen, &statisticsMock{})

	// URL-encoded values: a comma, a space and a multi-byte rune.
	target := "/fizzbuzz?int1=7&int2=11&limit=42&str1=x%2Cy&str2=%C3%A9%20%F0%9F%8D%95"

	statusCode, responseBody, _ := doRequest(t, handler, http.MethodGet, target)

	if statusCode != http.StatusOK {
		t.Fatalf("unexpected status code (got: %v, expected: %v)", statusCode, http.StatusOK)
	}

	if calls != 1 {
		t.Fatalf("generator called %d times, expected 1", calls)
	}

	expected := fizzbuzz.Params{Int1: 7, Int2: 11, Limit: 42, Str1: "x,y", Str2: "é 🍕"}
	if got != expected {
		t.Fatalf("unexpected params passed to (got: %+v, expected: %+v)", got, expected)
	}

	if result := decodeJSON[[]string](t, responseBody); !slices.Equal(result, canned) {
		t.Fatalf("unexpected result (got: %v, expected: %v)", result, canned)
	}
}
