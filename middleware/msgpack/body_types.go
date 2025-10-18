package msgpack

type M map[string]any

type Error string

type FieldError struct {
	Field string
	Error string
}
