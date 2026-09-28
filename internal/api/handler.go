package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

// Handle is the fizbuzz handler.
func Handler(w http.ResponseWriter, _ *http.Request) {
	var params fizzbuzz.Params

	params.SetDefaults()

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
