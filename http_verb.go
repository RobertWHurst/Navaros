package navaros

type HttpVerb string

const (
	All     HttpVerb = "ALL"
	Post    HttpVerb = "POST"
	Get     HttpVerb = "GET"
	Put     HttpVerb = "PUT"
	Patch   HttpVerb = "PATCH"
	Delete  HttpVerb = "DELETE"
	Options HttpVerb = "OPTIONS"
	Head    HttpVerb = "HEAD"
)
