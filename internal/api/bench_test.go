package api_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/api"
	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
	"github.com/peczenyj/go-fizzbuzz/internal/stats"
)

// discardWriter is an http.ResponseWriter that drops the body, so the
// benchmark measures the handler rather than a growing recorder buffer.
type discardWriter struct{ header http.Header }

func (w *discardWriter) Header() http.Header         { return w.header }
func (w *discardWriter) Write(p []byte) (int, error) { return io.Discard.Write(p) }
func (w *discardWriter) WriteHeader(int)             {}

// BenchmarkFizzBuzzHandler measures a full /fizzbuzz request: routing,
// parsing, generation, statistics and JSON encoding.
func BenchmarkFizzBuzzHandler(b *testing.B) {
	handler := api.New(fizzbuzz.DefaultGenerator(), stats.NewCounter[fizzbuzz.Params]())

	worst := strings.Repeat("%3C", 64) // 64 × "<": JSON-escaped to "<"

	for _, bc := range []struct {
		name   string
		target string
	}{
		{"classic_100", "/fizzbuzz?int1=3&int2=5&limit=100&str1=fizz&str2=buzz"},
		{"max_limit_1024", "/fizzbuzz?int1=3&int2=5&limit=1024&str1=fizz&str2=buzz"},
		{"worst_case", "/fizzbuzz?int1=1&int2=1&limit=1024&str1=" + worst + "&str2=" + worst},
	} {
		b.Run(bc.name, func(b *testing.B) {
			request := httptest.NewRequest(http.MethodGet, bc.target, nil)
			w := &discardWriter{header: http.Header{}}

			b.ReportAllocs()

			for b.Loop() {
				handler.ServeHTTP(w, request)
			}
		})
	}
}
