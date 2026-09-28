package fizzbuzz

// Field is a tiny type to help desambiguate a field from a string.
type Field string

const (
	FieldInt1  Field = `int1`
	FieldInt2  Field = `int2`
	FieldLimit Field = `limit`
	FieldStr1  Field = `str1`
	FieldStr2  Field = `str2`
)
