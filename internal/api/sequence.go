package api

import (
	"bufio"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"unicode/utf8"
)

// sequenceBufferSize is the size of the buffer a sequence is written through.
const sequenceBufferSize = 32 << 10

// maxEscaped is how many escaped strings writeSequence remembers per call:
// a fizzbuzz sequence has at most three labels (str1, str2 and str1str2).
const maxEscaped = 3

var sequenceWriters = sync.Pool{
	New: func() any { return bufio.NewWriterSize(nil, sequenceBufferSize) },
}

// writeSequence writes seq as the JSON body with statusCode. The body is byte
// for byte what writeJSON writes for a []string, including HTML escaping and
// the trailing newline, but it is produced differently:
//   - a string with no byte to escape, such as any number, is copied between
//     quotes as is;
//   - any other string is escaped by encoding/json once, and the result is
//     reused: a sequence repeats the same few labels, which encoding/json
//     would escape again for every element;
//   - the body goes out through a pooled 32 KB buffer, instead of being built
//     whole in memory first.
//
// The returned error is already logged; callers only need it to know if the
// body was sent.
func writeSequence(w http.ResponseWriter, statusCode int, seq []string) error {
	w.Header().Set(contentTypeHeader, mediaTypeJSON)

	w.WriteHeader(statusCode)

	bw, _ := sequenceWriters.Get().(*bufio.Writer)
	bw.Reset(w)

	defer func() {
		bw.Reset(nil) // don't keep the ResponseWriter alive in the pool
		sequenceWriters.Put(bw)
	}()

	err := encodeSequence(bw, seq)
	if err != nil {
		slog.Warn("unexpected error while writing the fizzbuzz sequence",
			slog.Any("error", err))
	}

	return err
}

// encodeSequence writes seq to bw as JSON and flushes bw. bufio.Writer errors
// are sticky, so the writes in between are not checked: Flush reports the
// first one.
func encodeSequence(bw *bufio.Writer, seq []string) error {
	if seq == nil {
		bw.WriteString("null\n")

		return bw.Flush()
	}

	var (
		escaped [maxEscaped]escapedString
		known   int
	)

	bw.WriteByte('[')

	for i, s := range seq {
		if i > 0 {
			bw.WriteByte(',')
		}

		if !needsEscape(s) {
			bw.WriteByte('"')
			bw.WriteString(s)
			bw.WriteByte('"')

			continue
		}

		b := lookup(escaped[:known], s)
		if b == nil {
			var err error

			if b, err = json.Marshal(s); err != nil {
				return err
			}

			if known < maxEscaped {
				escaped[known] = escapedString{s: s, json: b}
				known++
			}
		}

		bw.Write(b)
	}

	bw.WriteString("]\n")

	return bw.Flush()
}

// escapedString is a string and its JSON form.
type escapedString struct {
	s    string
	json []byte
}

// lookup returns the JSON form of s from escaped, or nil if s is not there.
func lookup(escaped []escapedString, s string) []byte {
	for _, e := range escaped {
		if e.s == s {
			return e.json
		}
	}

	return nil
}

// needsEscape reports whether encoding/json, with HTML escaping, could write s
// other than as is between quotes. It is conservative: any non-ASCII byte
// counts, even though encoding/json writes most valid UTF-8 as is.
func needsEscape(s string) bool {
	for i := range len(s) {
		switch c := s[i]; {
		case c < 0x20, c >= utf8.RuneSelf, c == '"', c == '\\', c == '<', c == '>', c == '&':
			return true
		}
	}

	return false
}
