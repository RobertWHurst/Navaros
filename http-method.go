package navaros

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
