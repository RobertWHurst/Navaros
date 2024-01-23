package navaros

import "strings"

// RequestParams represents the parameters extracted from the request path.
// Parameters are extracted from the path by matching the request path to
// the route pattern for the handler node. These may change each time Next
// is called on the context.
type RequestParams map[string]string

// Get returns the value of a given parameter key. If the key does not exist,
// an empty string is returned.
func (p RequestParams) Get(key string) string {
	for k, v := range p {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
