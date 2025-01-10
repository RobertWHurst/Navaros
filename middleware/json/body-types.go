package json

// M is shorthand for a map[string]any. It is provided as a convenience for
// defining JSON objects in a more concise manner.
type M map[string]any

// E is shorthand for Error. If you want to return a JSON wrapped error, you can
// use this type. The JSON response will be {"error": "your error message"}.
// If the status is not set, it will default to 400.
type E string
