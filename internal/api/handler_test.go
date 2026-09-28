package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
)

func TestHealthz(t *testing.T) {
	t.Parallel()

	request := httptest.NewRequest("GET", "/healthz", nil)

	w := httptest.NewRecorder()

	api.Healthz(w, request)

	response := w.Result()
	body, _ := io.ReadAll(response.Body)

	if response.StatusCode != http.StatusOK {
		t.Fatalf("unexpected http status code from endpoint /healthz (got: %v, expected: %v)", response.StatusCode, http.StatusOK)
	}

	if len(body) != 0 {
		t.Fatalf("unexpected http body: %v", string(body))
	}
}

func TestFizzBuzzHandler(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		target   string
		errBody  *api.ErrorBody
		expected []string
	}{
		{
			label:  "should return first 15 elements of default fizzbuzz sequence",
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=buzz",
			expected: []string{
				"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz",
			},
		},
		{
			label:  "should return one element if limit is 1",
			target: "/fizzbuzz?int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			expected: []string{
				"1",
			},
		},
		{
			label:   "should return error if int1 is not a number",
			target:  "/fizzbuzz?int1=lol&int2=5&limit=15&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `must be an integer`},
		},
		{
			label:   "should return error if int2 is not a number",
			target:  "/fizzbuzz?int1=3&int2=hehe&limit=ops&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `int2`, Reason: `must be an integer`},
		},
		{
			label:   "should return error if limit is not a number",
			target:  "/fizzbuzz?int1=3&int2=5&limit=ops&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must be an integer`},
		},
		{
			label:   "should return error if limit is not allowed",
			target:  "/fizzbuzz?int1=3&int2=5&limit=4096&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must not exceed max value 1024`},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest("GET", tc.target, nil)

			w := httptest.NewRecorder()

			api.FizzBuzz(w, request)

			response := w.Result()
			body, _ := io.ReadAll(response.Body)

			if tc.errBody != nil {
				if response.StatusCode != http.StatusBadRequest {
					t.Fatalf("unexpected http status code from endpoint /fizzbuzz (got: %v, expected: %v)", response.StatusCode, http.StatusBadRequest)
				}

				if contentType := response.Header.Get(api.ContentTypeHeaderName); contentType != api.ContentTypeApplicationJSON {
					t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, api.ContentTypeApplicationJSON)
				}

				var got api.ErrorBody

				t.Logf("body: %s", string(body))

				err := json.NewDecoder(bytes.NewReader(body)).Decode(&got)
				if err != nil {
					t.Fatalf("unexpected error while decode http response body: %v", err)
				}

				if got != *tc.errBody {
					t.Fatalf("unexpected result (got: %+v, expect: %+v)", got, tc.errBody)
				}

				return
			}

			if response.StatusCode != http.StatusOK {
				t.Fatalf("unexpected http status code from endpoint /fizzbuzz (got: %v, expected: %v)", response.StatusCode, http.StatusOK)
			}

			if contentType := response.Header.Get(api.ContentTypeHeaderName); contentType != api.ContentTypeApplicationJSON {
				t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, api.ContentTypeApplicationJSON)
			}

			if len(body) == 0 {
				t.Fatalf("unexpected empty http body")
			}

			var got []string

			err := json.NewDecoder(bytes.NewReader(body)).Decode(&got)
			if err != nil {
				t.Fatalf("unexpected error while decode http response body: %v", err)
			}

			if !slices.Equal(got, tc.expected) {
				t.Fatalf("unexpected result (got: %v, expect: %v)", got, tc.expected)
			}
		})
	}
}
