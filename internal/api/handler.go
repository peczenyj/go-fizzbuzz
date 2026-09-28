package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

// Handle is the fizbuzz handler.
func Handler(w http.ResponseWriter, r *http.Request) {
	var (
		params fizzbuzz.Params
		err    error
	)

	params.SetDefaults()

	query := r.URL.Query()

	if query.Has("int1") {
		params.Int1, err = strconv.Atoi(query.Get("int1"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error: unable to parse param 'int1': %v", err)

			return
		}
	}
	if query.Has("int2") {
		params.Int2, err = strconv.Atoi(query.Get("int2"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error: unable to parse param 'int2': %v", err)

			return
		}
	}
	if query.Has("limit") {
		params.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "error: unable to parse param 'limit': %v", err)

			return
		}
	}

	if query.Has("str1") {
		params.Str1 = query.Get("int1")
	}

	if query.Has("str2") {
		params.Str2 = query.Get("int2")
	}

	result, err := fizzbuzz.Generate(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error: %v", err)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		slog.Warn("unexpected error while perform json encode", slog.Any("error", err))
	}
}

// Healthz k8s api health endpoint.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
