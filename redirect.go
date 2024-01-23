package navaros

// Redirect represents an HTTP redirect. If you want to redirect a request
// in a handler, you can initialize a Redirect with a target relative path
// or absolute URL, and set it as the body of the context. This will
// cause Navaros to send the client a Location header with the redirect
// url, and a 302 status code. The status code can be changed by setting
// the Status field of the context.
type Redirect struct {
	To string
}
