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

	statusCode, responseBody, _ := doRequest(t, api.Healthz, "/healthz")

	if statusCode != http.StatusOK {
		t.Fatalf("unexpected http status code from endpoint /healthz (got: %v, expected: %v)", statusCode, http.StatusOK)
	}

	if len(responseBody) != 0 {
		t.Fatalf("unexpected http body: %v", string(responseBody))
	}
}

func TestFizzBuzzHandler(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label    string
		target   string
		errBody  *api.ErrorBody
		expected []string

		skipExpectedCheck bool

		extraResultCheck func(*testing.T, []string)
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
			label:  "should return one element if limit is 1",
			target: "/fizzbuzz?int1=3&int2=5&limit=1&str1=fizz&str2=buzz",
			expected: []string{
				"1",
			},
		},
		{
			label:  "should return success if limit is equal to max value",
			target: "/fizzbuzz?int1=3&int2=5&limit=1024&str1=fizz&str2=buzz",

			skipExpectedCheck: true,
			extraResultCheck: func(t *testing.T, s []string) {
				t.Helper()

				if got := len(s); got != 1024 {
					t.Fatalf("unexpected result length size (got: %v, expected: %v)", got, 1024)
				}
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
			label:   "should return error if int1 is zero",
			target:  "/fizzbuzz?int1=0&int2=5&limit=15&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `int1`, Reason: `must be bigger than zero`},
		},
		{
			label:   "should return error if int2 is zero",
			target:  "/fizzbuzz?int1=3&int2=0&limit=15&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `int2`, Reason: `must be bigger than zero`},
		},
		{
			label:   "should return error if limit is zero",
			target:  "/fizzbuzz?int1=3&int2=5&limit=0&str1=fizz&str2=buzz",
			errBody: &api.ErrorBody{Error: `invalid parameter`, Field: `limit`, Reason: `must be bigger than zero`},
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

			statusCode, responseBody, responseHeaders := doRequest(t, api.FizzBuzz, tc.target)

			if tc.errBody != nil {
				if statusCode != http.StatusBadRequest {
					t.Fatalf("unexpected http status code from endpoint /fizzbuzz (got: %v, expected: %v)", statusCode, http.StatusBadRequest)
				}

				if contentType := responseHeaders.Get(api.ContentTypeHeaderName); contentType != `application/json` {
					t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, `application/json`)
				}

				var got api.ErrorBody

				err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&got)
				if err != nil {
					t.Fatalf("unexpected error while decode http response body: %v", err)
				}

				if got != *tc.errBody {
					t.Fatalf("unexpected result (got: %+v, expect: %+v)", got, tc.errBody)
				}

				return
			}

			if statusCode != http.StatusOK {
				t.Fatalf("unexpected http status code from endpoint /fizzbuzz (got: %v, expected: %v)", statusCode, http.StatusOK)
			}

			if contentType := responseHeaders.Get(api.ContentTypeHeaderName); contentType != `application/json` {
				t.Fatalf("unexpected content type (got: %v, expected %v)", contentType, `application/json`)
			}

			if len(responseBody) == 0 {
				t.Fatalf("unexpected empty http body")
			}

			var got []string

			err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&got)
			if err != nil {
				t.Fatalf("unexpected error while decode http response body: %v", err)
			}

			if tc.extraResultCheck != nil {
				tc.extraResultCheck(t, got)
			}

			if tc.skipExpectedCheck {
				return
			}

			if !slices.Equal(got, tc.expected) {
				t.Fatalf("unexpected result (got: %v, expect: %v)", got, tc.expected)
			}
		})
	}
}

func doRequest(t *testing.T,
	handleFunc func(w http.ResponseWriter, r *http.Request),
	target string,
) (
	statusCode int,
	body []byte,
	headers http.Header,
) {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, target, nil)

	w := httptest.NewRecorder()

	handleFunc(w, request)

	response := w.Result()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error while read response body: %v", err)
	}

	return response.StatusCode, body, response.Header
}
