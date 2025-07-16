package navarostest

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"regexp"
	"strings"

	"github.com/RobertWHurst/navaros"
)

// WithHandlerRequest defines the options for executing a request against a
// handler.
type WithHandlerRequest struct {
	Method string
	Path   string
	Query  map[string]string
	Params map[string]string
	Body   []byte
}

// WithHandler executes a request against a handler and returns the result.
// It is used to test handlers in isolation, allowing for assertions on the
// response headers, status codes, and body content.
//
// The handler can be a Navaros router or any compatible http.Handler or
// http.HandlerFunc.
func WithHandler(t *testing.T, reqOpts *WithHandlerRequest, handler any) WithHandlerResult {
	req := httptest.NewRequest(reqOpts.Method, reqOpts.Path, bytes.NewReader(reqOpts.Body))
	res := httptest.NewRecorder()

	ctx := navaros.NewContext(res, req, handler)

	ctx.Next()
	navaros.CtxFinalize(ctx)
	navaros.CtxFree(ctx)

	return WithHandlerResult{
		t:   t,
		req: req,
		res: res.Result(),
	}
}

// WithHandlerResult provides methods to check the response from a handler.
// It is used to assert various properties of the response, such as headers,
// status codes, and body content.
type WithHandlerResult struct {
	t   *testing.T
	req *http.Request
	res *http.Response
}

// Header returns the response headers.
func (ra *WithHandlerResult) Header() http.Header {
	return ra.res.Header
}

// HasHeader checks if the response has a specific header.
func (ra *WithHandlerResult) HasHeader(name string) {
	_, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
	}
}

// HeaderCount checks if a specific header has the expected number of values.
// In other words, if a header is set multiple times, this checks how many
// times it was set.
func (ra *WithHandlerResult) HeaderCount(name string, expected int) {
	values, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
		return
	}
	if len(values) != expected {
		ra.t.Errorf("Expected response header %q to have %d values, but got %d", name, expected, len(values))
	}
}

// HeaderEquals checks if a specific header has a specific value.
func (ra *WithHandlerResult) HeaderEquals(name, value string) {
	values, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
		return
	}
	for _, v := range values {
		if v == value {
			return
		}
	}
	ra.t.Errorf("Expected response header %q to equal %q, but got %v", name, value, values)
}

// HeaderStartsWith checks if a specific header starts with a specific prefix.
func (ra *WithHandlerResult) HeaderStartsWith(name, prefix string) {
	values, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
		return
	}
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return
		}
	}
	ra.t.Errorf("Expected response header %q to start with %q, but got %v", name, prefix, values)
}

// HeaderContains checks if a specific header contains a specific substring.
func (ra *WithHandlerResult) HeaderContains(name, substr string) {
	values, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
		return
	}
	for _, value := range values {
		if strings.Contains(value, substr) {
			return
		}
	}
	ra.t.Errorf("Expected response header %q to contain %q, but got %v", name, substr, values)
}

// HeaderEndsWith checks if a specific header ends with a specific suffix.
func (ra *WithHandlerResult) HeaderEndsWith(name, suffix string) {
	values, ok := ra.res.Header[name]
	if !ok {
		ra.t.Errorf("Expected response to have header %q, but it was not found", name)
		return
	}
	for _, value := range values {
		if strings.HasSuffix(value, suffix) {
			return
		}
	}
	ra.t.Errorf("Expected response header %q to end with %q, but got %v", name, suffix, values)
}

// StatusCode returns the response status code.
func (ra *WithHandlerResult) StatusCode() int {
	return ra.res.StatusCode
}

// IsSuccessRange checks if the response status code is in the 2xx range.
func (ra *WithHandlerResult) IsSuccessRange() {
	if ra.res.StatusCode > 299 || ra.res.StatusCode < 200 {
		ra.t.Errorf("Expected response status code to be 2xx, but got %d", ra.res.StatusCode)
	}
}

// IsRedirectionRange checks if the response status code is in the 3xx range.
func (ra *WithHandlerResult) IsRedirectionRange() {
	if ra.res.StatusCode < 300 || ra.res.StatusCode >= 400 {
		ra.t.Errorf("Expected response status code to be 3xx, but got %d", ra.res.StatusCode)
	}
}

// IsClientErrorRange checks if the response status code is in the 4xx range.
func (ra *WithHandlerResult) IsClientErrorRange() {
	if ra.res.StatusCode < 400 || ra.res.StatusCode >= 500 {
		ra.t.Errorf("Expected response status code to be 4xx, but got %d", ra.res.StatusCode)
	}
}

// IsServerErrorRange checks if the response status code is in the 5xx range.
func (ra *WithHandlerResult) IsServerErrorRange() {
	if ra.res.StatusCode < 500 || ra.res.StatusCode >= 600 {
		ra.t.Errorf("Expected response status code to be 5xx, but got %d", ra.res.StatusCode)
	}
}

// IsStatusCode checks if the response status code matches the expected value.
func (ra *WithHandlerResult) IsStatusCode(expected int) {
	if ra.res.StatusCode != expected {
		ra.t.Errorf("Expected response status code to be %d, but got %d", expected, ra.res.StatusCode)
	}
}

// Body returns the response body as an io.ReadCloser.
func (ra *WithHandlerResult) Body() io.ReadCloser {
	return ra.res.Body
}

// BodyString reads the response body and returns it as a string.
func (ra *WithHandlerResult) BodyString() string {
	body, err := io.ReadAll(ra.res.Body)
	if err != nil {
		ra.t.Errorf("Failed to read response body: %v", err)
		return ""
	}
	return string(body)
}

// HasBody checks if the response has a body and is not empty.
func (ra *WithHandlerResult) HasBody() {
	body := ra.BodyString()
	if body == "" {
		ra.t.Error("Expected response to have a body, but it was empty")
	}
}

// BodyEquals checks if the response body equals the expected string.
func (ra *WithHandlerResult) BodyEquals(expected string) {
	body := ra.BodyString()
	if body != expected {
		ra.t.Errorf("Expected response body to equal %q, but got %q", expected, body)
	}
}

// BodyContains checks if the response body contains a specific substring.
func (ra *WithHandlerResult) BodyContains(substr string) {
	body := ra.BodyString()
	if !strings.Contains(body, substr) {
		ra.t.Errorf("Expected response body to contain %q, but got %q", substr, body)
	}
}

// BodyMatches checks if the response body matches a regular expression.
func (ra *WithHandlerResult) BodyMatches(regex *regexp.Regexp) {
	body := ra.BodyString()
	if !regex.MatchString(body) {
		ra.t.Errorf("Expected response body to match regex %q, but got %q", regex.String(), body)
	}
}
