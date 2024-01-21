package navaros

import (
	"errors"
	"strings"
)

type HTTPMethod string

const (
	All     HTTPMethod = "ALL"
	Post    HTTPMethod = "POST"
	Get     HTTPMethod = "GET"
	Put     HTTPMethod = "PUT"
	Patch   HTTPMethod = "PATCH"
	Delete  HTTPMethod = "DELETE"
	Options HTTPMethod = "OPTIONS"
	Head    HTTPMethod = "HEAD"
)

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
