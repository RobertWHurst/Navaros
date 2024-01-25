package navaros

import (
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

// MaxRequestBodySize is the maximum size of a request body. Changing this
// value will affect all requests. Set to -1 to disable the limit. If
// MaxRequestBodySize is set on the context it will override this value. This
// setting is useful for preventing denial of service attacks. It is not
// recommended to set this value to -1 unless you know what you are doing!!!
var MaxRequestBodySize int64 = 1024 * 1024 * 10 // 10MB

// Context represents a request and response. Navaros handlers access the
// request and build the response through the context.
type Context struct {
	parentContext *Context

	request *http.Request

	method HTTPMethod
	path   string
	params RequestParams

	Status            int
	Headers           http.Header
	Cookies           []*http.Cookie
	Body              any
	bodyWriter        http.ResponseWriter
	hasWrittenHeaders bool
	hasWrittenBody    bool

	MaxRequestBodySize int64

	Error           error
	ErrorStack      string
	FinalError      error
	FinalErrorStack string

	requestBodyUnmarshaller func(into any) error
	responseBodyMarshaller  func(from any) (io.Reader, error)

	currentHandlerNode               *HandlerNode
	matchingHandlerNode              *HandlerNode
	currentHandlerOrTransformerIndex int
	currentHandlerOrTransformer      any

	associatedValues map[string]any

	deadline     *time.Time
	doneHandlers []func()
}

// NewContext creates a new Context from go's http.ResponseWriter and
// http.Request. It also takes a variadic list of handlers. This is useful for
// creating a new Context outside of a router, and can be used by libraries
// which wish to extend or encapsulate the functionality of Navaros.
func NewContext(res http.ResponseWriter, req *http.Request, handlers ...any) *Context {
	return NewContextWithNode(res, req, &HandlerNode{
		Method:                  All,
		HandlersAndTransformers: handlers,
	})
}

// NewContextWithNode creates a new Context from go's http.ResponseWriter and
// http.Request. It also takes a HandlerNode - a link in a chain of handlers.
// This is useful for creating a new Context outside of a router, and can be
// used by libraries which wish to extend or encapsulate the functionality of
// Navaros. For example, implementing a custom router.
func NewContextWithNode(res http.ResponseWriter, req *http.Request, firstHandlerNode *HandlerNode) *Context {
	return &Context{
		request: req,

		method: HTTPMethod(req.Method),
		path:   req.URL.Path,

		Headers:    http.Header{},
		bodyWriter: res,

		currentHandlerNode: &HandlerNode{
			Method: All,
			Next:   firstHandlerNode,
		},

		associatedValues: map[string]any{},

		doneHandlers: []func(){},
	}
}

// NewSubContextWithNode creates a new Context from an existing Context. This
// is useful when you want to create a new Context from an existing one, but
// with a different handler chain. Note that when the end of the sub context's
// handler chain is reached, the parent context's handler chain will continue.
func NewSubContextWithNode(ctx *Context, firstHandlerNode *HandlerNode) *Context {
	finalHandlerNode := firstHandlerNode
	for finalHandlerNode.Next != nil {
		finalHandlerNode = finalHandlerNode.Next
	}
	finalHandlerNode.Next = &HandlerNode{
		Method:                  All,
		HandlersAndTransformers: []any{func(_ *Context) { ctx.Next() }},
	}

	subContext := *ctx
	subContext.parentContext = ctx
	subContext.currentHandlerNode = &HandlerNode{
		Method:                  All,
		Pattern:                 nil,
		HandlersAndTransformers: []any{},
		Next:                    firstHandlerNode,
	}
	subContext.matchingHandlerNode = nil
	subContext.currentHandlerOrTransformerIndex = 0
	subContext.currentHandlerOrTransformer = nil
	return &subContext
}

// Next calls the next handler in the chain. This is useful for creating
// middleware style handlers that work on the context before and/or after the
// responding handler.
func (c *Context) Next() {
	c.next()
}

func (s Context) Set(key string, value any) {
	s.associatedValues[key] = value
}

func (s Context) Get(key string) any {
	return s.associatedValues[key]
}

// Method returns the HTTP method of the request.
func (c *Context) Method() HTTPMethod {
	return c.method
}

// Path returns the path of the request.
func (c *Context) Path() string {
	return c.path
}

// URL returns the URL of the request.
func (c *Context) URL() *url.URL {
	return c.request.URL
}

// Params returns the parameters of the request. These are defined by the
// route pattern used to bind each handler, and may be different for each
// time next is called.
func (c *Context) Params() RequestParams {
	return c.params
}

// Query returns the query parameters of the request.
func (c *Context) Query() url.Values {
	return c.request.URL.Query()
}

// Protocol returns the http protocol version of the request.
func (c *Context) Protocol() string {
	return c.request.Proto
}

// ProtocolMajor returns the major number in http protocol version.
func (c *Context) ProtocolMajor() int {
	return c.request.ProtoMajor
}

// ProtocolMinor returns the minor number in http protocol version.
func (c *Context) ProtocolMinor() int {
	return c.request.ProtoMinor
}

// RequestHeaders returns the request headers.
func (c *Context) RequestHeaders() http.Header {
	return c.request.Header
}

// RequestTrailers returns the trailing headers of the request if set.
func (c *Context) RequestTrailers() http.Header {
	return c.request.Trailer
}

// RequestCookies returns the value of a request cookie by name. Returns nil
// if the cookie does not exist.
func (c *Context) RequestCookie(name string) (*http.Cookie, error) {
	return c.request.Cookie(name)
}

// RequestBodyReader returns a reader setup to read in the request body. This
// is useful for streaming the request body, or for middleware which decodes
// the request body. Without body handling middleware, the request body reader
// is the only way to access request body data.
func (c *Context) RequestBodyReader() io.ReadCloser {
	maxRequestBodySize := c.MaxRequestBodySize
	if c.MaxRequestBodySize == 0 {
		maxRequestBodySize = MaxRequestBodySize
	}
	if maxRequestBodySize == -1 {
		return c.request.Body
	}
	return http.MaxBytesReader(c.bodyWriter, c.request.Body, maxRequestBodySize)
}

// Allows middleware to intercept the request body reader and replace it with
// their own. This is useful transformers that re-write the request body
// in a streaming fashion. It's also useful for transformers that re-encode
// the request body.
func (c *Context) SetRequestBodyReader(reader io.Reader) {
	if readCloser, ok := reader.(io.ReadCloser); ok {
		c.request.Body = readCloser
	} else {
		c.request.Body = io.NopCloser(reader)
	}
}

// UnmarshalRequestBody unmarshals the request body into a given value. Note
// that is method requires SetRequestBodyUnmarshaller to be called first. This
// likely is done by middleware for parsing request bodies.
func (c *Context) UnmarshalRequestBody(into any) error {
	if c.requestBodyUnmarshaller == nil {
		return errors.New("no request body unmarshaller set. use SetRequestBodyUnmarshaller() or add body parser middleware")
	}
	return c.requestBodyUnmarshaller(into)
}

// SetRequestBodyUnmarshaller sets the request body unmarshaller. Middleware
// that parses request bodies should call this method to set the unmarshaller.
func (c *Context) SetRequestBodyUnmarshaller(unmarshaller func(into any) error) {
	c.requestBodyUnmarshaller = unmarshaller
}

// SetResponseBodyMarshaller sets the response body marshaller. Middleware
// that encodes response bodies should call this method to set the marshaller.
func (c *Context) SetResponseBodyMarshaller(marshaller func(from any) (io.Reader, error)) {
	c.responseBodyMarshaller = marshaller
}

// RequestContentLength returns the length of the request body if provided by
// the client.
func (c *Context) RequestContentLength() int64 {
	return c.request.ContentLength
}

// RequestTransferEncoding returns the transfer encoding of the request
func (c *Context) RequestTransferEncoding() []string {
	return c.request.TransferEncoding
}

// RequestHost returns the host of the request. Useful for determining the
// source of the request.
func (c *Context) RequestHost() string {
	return c.request.Host
}

// RequestRemoteAddress returns the remote address of the request. Useful for
// determining the source of the request.
func (c *Context) RequestRemoteAddress() string {
	return c.request.RemoteAddr
}

// RequestRawURI returns the raw URI of the request. This will be the original
// value from the request headers.
func (c *Context) RequestRawURI() string {
	return c.request.RequestURI
}

// RequestTLS returns the TLS connection state of the request if the request
// is using TLS.
func (c *Context) RequestTLS() *tls.ConnectionState {
	return c.request.TLS
}

// Request returns the underlying http.Request object.
func (c *Context) Request() *http.Request {
	return c.request
}

// ResponseWriter returns the underlying http.ResponseWriter object.
func (c *Context) ResponseWriter() http.ResponseWriter {
	return c.bodyWriter
}

// Write writes bytes to the response body. This is useful for streaming the
// response body, or for middleware which encodes the response body.
func (c *Context) Write(bytes []byte) (int, error) {
	c.hasWrittenBody = true

	if !c.hasWrittenHeaders {
		c.hasWrittenHeaders = true
		if c.Status == 0 {
			c.Status = 200
		}
		c.bodyWriter.WriteHeader(c.Status)
	}

	return c.bodyWriter.Write(bytes)
}

// Flush sends any bytes buffered in the response body to the client. Buffering
// is controlled by go's http.ResponseWriter.
func (c *Context) Flush() {
	c.bodyWriter.(http.Flusher).Flush()
}

// Deadline returns the deadline of the request. Deadline is part of the go
// context.Context interface.
func (c *Context) Deadline() (time.Time, bool) {
	ok := c.deadline != nil
	deadline := time.Time{}
	if ok {
		deadline = *c.deadline
	}
	return deadline, ok
}

// Done added for compatibility with go's context.Context. Alias for
// UntilFinish(). Done is part of the go context.Context interface.
func (c *Context) Done() <-chan struct{} {
	doneChan := make(chan struct{})
	c.doneHandlers = append(c.doneHandlers, func() {
		doneChan <- struct{}{}
	})
	return doneChan
}

// Err returns the final error of the request. Will be nil if the request
// is still being served even if an error has occurred. Populated once the
// request is done. Err is part of the go context.Context interface.
func (c *Context) Err() error {
	return c.FinalError
}

// Value is a noop for compatibility with go's context.Context.
func (c *Context) Value(key any) any {
	return nil
}

// marshallResponseBody uses a responseBodyMarshaller to marshall the response
// body into a reader if one has been set with SetResponseBodyMarshaller.
// It will return an error if no marshaller has been set.
func (c *Context) marshallResponseBody() (io.Reader, error) {
	if c.responseBodyMarshaller == nil {
		return nil, errors.New("no response body marshaller set. use SetResponseBodyMarshaller() or add body encoder middleware")
	}
	return c.responseBodyMarshaller(c.Body)
}

// tryUpdateParent updates the parent context with the current context's
// state. This is called by Next() when the current context is a sub context.
func (c *Context) tryUpdateParent() {
	if c.parentContext == nil {
		return
	}

	c.parentContext.Status = c.Status
	c.parentContext.Headers = c.Headers
	c.parentContext.Body = c.Body
	c.parentContext.Error = c.Error
	c.parentContext.ErrorStack = c.ErrorStack
	c.parentContext.hasWrittenHeaders = c.hasWrittenHeaders
	c.parentContext.hasWrittenBody = c.hasWrittenBody
}
