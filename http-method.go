package navaros

import (
	"errors"
	"strings"
)

// HTTPMethod represents an HTTP method.
type HTTPMethod string

const (
	// All represents all HTTP methods.
	All HTTPMethod = "ALL"
	// Post represents the HTTP POST method.
	Post HTTPMethod = "POST"
	// Get represents the HTTP GET method.
	Get HTTPMethod = "GET"
	// Put represents the HTTP PUT method.
	Put HTTPMethod = "PUT"
	// Patch represents the HTTP PATCH method.
	Patch HTTPMethod = "PATCH"
	// Delete represents the HTTP DELETE method.
	Delete HTTPMethod = "DELETE"
	// Options represents the HTTP OPTIONS method.
	Options HTTPMethod = "OPTIONS"
	// Head represents the HTTP HEAD method.
	Head HTTPMethod = "HEAD"
)

// HTTPMethodFromString converts a string to an HTTPMethod.
// If the string is not a valid HTTP method, an error is returned.
func HTTPMethodFromString(method string) HTTPMethod {
	switch strings.ToUpper(method) {
	case "ALL", "*":
		return All
	case "POST":
		return Post
	case "GET":
		return Get
	case "PUT":
		return Put
	case "PATCH":
		return Patch
	case "DELETE":
		return Delete
	case "OPTIONS":
		return Options
	case "HEAD":
		return Head
	default:
		panic(errors.New("invalid http method `" + method + "`"))
	}
}
