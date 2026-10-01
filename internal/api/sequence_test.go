package api

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

// checkSameAsJSON checks that writeSequence writes exactly what writeJSON
// writes for seq: status, Content-Type and body.
func checkSameAsJSON(t *testing.T, seq []string) {
	t.Helper()

	want := httptest.NewRecorder()
	if err := writeJSON(want, http.StatusOK, seq); err != nil {
		t.Fatalf("unexpected error from writeJSON: %v", err)
	}

	got := httptest.NewRecorder()
	if err := writeSequence(got, http.StatusOK, seq); err != nil {
		t.Fatalf("unexpected error from writeSequence: %v", err)
	}

	if got.Code != want.Code {
		t.Fatalf("unexpected status (got: %d, expect: %d)", got.Code, want.Code)
	}

	if g, w := got.Header().Get(contentTypeHeader), want.Header().Get(contentTypeHeader); g != w {
		t.Fatalf("unexpected Content-Type (got: %q, expect: %q)", g, w)
	}

	if !bytes.Equal(got.Body.Bytes(), want.Body.Bytes()) {
		t.Fatalf("unexpected body for %q\n got: %q\nwant: %q", seq, got.Body.Bytes(), want.Body.Bytes())
	}
}

func TestWriteSequence(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		label string
		seq   []string
	}{
		{label: "nil", seq: nil},
		{label: "empty", seq: []string{}},
		{label: "numbers", seq: []string{"1", "2", "100", "1024"}},
		{label: "classic", seq: []string{"1", "2", "fizz", "4", "buzz", "fizz", "7", "8", "fizz", "buzz", "11", "fizz", "13", "14", "fizzbuzz"}},
		{label: "empty string", seq: []string{"", "1", ""}},
		{label: "HTML", seq: []string{"<b>", "1", "&amp;", "<b>", "&amp;", "<b>&amp;"}},
		{label: "quote and backslash", seq: []string{`"`, `\`, `a"b\c`, `"`}},
		{label: "control bytes", seq: []string{"\x00", "\t\n\r", "\x1f", "\x7f", "\b\f"}},
		{label: "UTF-8", seq: []string{"é", "日本", "🎉", "é"}},
		{label: "invalid UTF-8", seq: []string{"\xff", "a\xc3", "\xc3\x28", "\xff"}},
		{label: "line and paragraph separators", seq: []string{" ", " ", "a b"}},
		{label: "more distinct escaped strings than remembered", seq: []string{"<1", "<2", "<3", "<4", "<5", "<1", "<5"}},
		{label: "label that looks like a number", seq: []string{"1", "2", "3", "3"}},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()

			checkSameAsJSON(t, tc.seq)
		})
	}
}

// TestWriteSequenceLarge writes a body much larger than the buffer, so it
// goes out in several writes.
func TestWriteSequenceLarge(t *testing.T) {
	t.Parallel()

	label := strings.Repeat("<", 255)
	seq := make([]string, 0, 10_000)

	for i := range cap(seq) {
		if i%2 == 0 {
			seq = append(seq, label)
		} else {
			seq = append(seq, strconv.Itoa(i))
		}
	}

	checkSameAsJSON(t, seq)
}

// failingWriter is an http.ResponseWriter whose body writes always fail.
type failingWriter struct{ header http.Header }

var errWrite = errors.New("write failed")

func (w *failingWriter) Header() http.Header       { return w.header }
func (w *failingWriter) Write([]byte) (int, error) { return 0, errWrite }
func (w *failingWriter) WriteHeader(int)           {}

func TestWriteSequenceWriteError(t *testing.T) {
	t.Parallel()

	for _, seq := range [][]string{nil, {"1", "<2>"}} {
		err := writeSequence(&failingWriter{header: http.Header{}}, http.StatusOK, seq)
		if !errors.Is(err, errWrite) {
			t.Fatalf("unexpected error for %q (got: %v, expect: %v)", seq, err, errWrite)
		}
	}
}

// FuzzWriteSequence checks that writeSequence and writeJSON write the same
// body, both for generated sequences and for sequences mixing any strings.
func FuzzWriteSequence(f *testing.F) {
	generator := fizzbuzz.DefaultGenerator()

	f.Add(3, 5, 15, "fizz", "buzz", "x")
	f.Add(1, 1, 3, "<", "&", " ")
	f.Add(2, 3, 10, "\xff", "\xc3", "\x00")
	f.Add(7, 7, 20, `"\`, "é", "")

	f.Fuzz(func(t *testing.T, int1, int2, limit int, str1, str2, other string) {
		checkSameAsJSON(t, []string{str1, str2, str1 + str2, other, "100", str1, other, str2, str1 + str2})

		params := fizzbuzz.Params{Int1: int1, Int2: int2, Limit: limit, Str1: str1, Str2: str2}

		if seq, err := generator.Generate(params); err == nil {
			checkSameAsJSON(t, seq)
		}
	})
}
