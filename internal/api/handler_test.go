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
			label:  "should return first 100 elements of default fizzbuzz sequence with no arguments",
			target: "/fizzbuzz",
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
			target: "/fizzbuzz?int1=3&int2=5&limit=15&str1=abc&str2=xyz",
			expected: []string{
				"1", "2", "abc", "4", "xyz", "abc", "7", "8", "abc", "xyz", "11", "abc", "13", "14", "abcxyz",
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
			target:  "/fizzbuzz?int1=3&int2=hehe&limit=15&str1=fizz&str2=buzz",
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
		{
			label:   "should return error if str1 is empty",
			target:  "/fizzbuzz?int1=3&int2=5&limit=15&str1=&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `str1`, Reason: `must not be empty string`},
		},
		{
			label:   "should return error if str2 is empty",
			target:  "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `str2`, Reason: `must not be empty string`},
		},
		{
			label:   "should return error if str1 is too big",
			target:  "/fizzbuzz?int1=3&int2=5&limit=15&str1=12345678901234567890123456789012345678901234567890123456789012345&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `str1`, Reason: `must not exceed max string length 64`},
		},
		{
			label:   "should return error if str2 is too big",
			target:  "/fizzbuzz?int1=3&int2=5&limit=15&str1=fizz&str2=12345678901234567890123456789012345678901234567890123456789012345",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `str2`, Reason: `must not exceed max string length 64`},
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

				if contentType := response.Header.Get(api.ContentTypeHeaderName); contentType != `application/json` {
					t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, `application/json`)
				}

				var got api.ErrorBody

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

			if contentType := response.Header.Get(api.ContentTypeHeaderName); contentType != `application/json` {
				t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, `application/json`)
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
