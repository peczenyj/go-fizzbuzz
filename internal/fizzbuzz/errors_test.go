package fizzbuzz_test

import (
	"errors"
	"testing"

	"github.com/peczenyj/go-fizzbuzz/internal/fizzbuzz"
)

func TestParamError(t *testing.T) {
	t.Parallel()

	pe := &fizzbuzz.ParamError{Field: fizzbuzz.FieldInt1, Err: fizzbuzz.ErrMustBeBiggerThanZero}

	expected := `invalid field "int1": must be bigger than zero`

	if pe.Error() != expected {
		t.Fatalf("unexpected error message (got: %v, expected: %v)", pe.Error(), expected)
	}

	if !errors.Is(pe, fizzbuzz.ErrMustBeBiggerThanZero) {
		t.Fatalf("unexpected unwrap error (expected: %v)", fizzbuzz.ErrMustBeBiggerThanZero)
	}
}
