package navaros

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

// Context represents a request and response. Navaros handlers access the
// request and build the response through the context.
type Context struct {
	parentContext *Context

	method            HTTPMethod
	path              string
	url               *url.URL
	params            RequestParams
	query             url.Values
	requestHeaders    RequestHeaders
	requestBodyReader io.ReadCloser
	requestBodyBytes  []byte

	Status            int
	Headers           ResponseHeaders
	Body              any
	bodyWriter        http.ResponseWriter
	hasWrittenHeaders bool
	hasWrittenBody    bool

	Error           error
	ErrorStack      string
	FinalError      error
	FinalErrorStack string

	FinalNext func()

	requestBodyUnmarshaller func(into any) error
	responseBodyMarshaller  func(from any) ([]byte, error)

	currentHandlerNode               *handlerNode
	matchingHandlerNode              *handlerNode
	currentHandlerOrTransformerIndex int
	currentHandlerOrTransformer      any

	deadline     *time.Time
	doneHandlers []func()
}

var contextData = make(map[*Context]map[string]any)

// NewContext creates a new Context from go's http.ResponseWriter and
// http.Request. It also takes a handler node - the start of the handler
// chain.
func NewContext(responseWriter http.ResponseWriter, request *http.Request, firstHandlerNode *handlerNode) *Context {
	return &Context{
		method:            HTTPMethod(request.Method),
		path:              request.URL.Path,
		url:               request.URL,
		query:             request.URL.Query(),
		requestHeaders:    RequestHeaders(request.Header),
		requestBodyReader: request.Body,

		Headers:    map[string]string{},
		bodyWriter: responseWriter,

		currentHandlerNode: &handlerNode{
			method:                  All,
			handlersAndTransformers: []any{},
			next:                    firstHandlerNode,
		},

		doneHandlers: []func(){},
	}
}

// NewSubContext creates a new Context from an existing Context. This is useful
// when you want to create a new Context from an existing one, but with a
// different handler chain.
func NewSubContext(ctx *Context, firstHandlerNode *handlerNode, finalNext func()) *Context {
	subContext := *ctx
	subContext.parentContext = ctx
	subContext.currentHandlerNode = &handlerNode{
		method:                  All,
		pattern:                 nil,
		handlersAndTransformers: []any{},
		next:                    firstHandlerNode,
	}
	subContext.matchingHandlerNode = nil
	subContext.currentHandlerOrTransformerIndex = 0
	subContext.currentHandlerOrTransformer = nil
	subContext.FinalNext = finalNext
	return &subContext
}

// Next calls the next handler in the chain. This is useful for creating
// middleware style handlers that work on the context before and/or after the
// responding handler.
func (c *Context) Next() {
	shouldRunFinalNext := false

	// In the case that this is a sub context, we need to update the parent
	// context with the current context's state.
	defer func() {
		c.tryUpdateParent()
		if shouldRunFinalNext {
			c.FinalNext()
		}
	}()

	if c.Error != nil {
		return
	}

	// If we have exceeded the deadline, we can return early.
	if c.deadline != nil && time.Now().After(*c.deadline) {
		c.Error = errors.New("request exceeded timeout deadline")
		return
	}

	// walk the chain looking for a handler with a pattern that matches the method
	// and path of the request, or until we reach the end of the chain
	for c.currentHandlerNode != nil {

		// Because handlers can have multiple handler functions or transformers,
		// we may save a matching handler node to the context so that we can
		// continue from the same handler until we have executed all of it's
		// handlers and transformers.
		//
		// If we do not have a matching handler node, we will walk the chain
		// until we find a matching handler node.
		if c.matchingHandlerNode == nil {
			for c.currentHandlerNode != nil {
				if c.tryMatchHandlerNode(c.currentHandlerNode) {
					c.matchingHandlerNode = c.currentHandlerNode
					break
				}
				c.currentHandlerNode = c.currentHandlerNode.next
			}
			if c.matchingHandlerNode == nil {
				shouldRunFinalNext = true
				return
			}
		}

		// Grab a handler function or transformer from the matching handler node.
		// If there are more than one, we will continue from the same handler node
		// the next time Next is called. We iterate through the handler functions
		// and transformers until we have executed all of them.
		if c.currentHandlerOrTransformerIndex < len(c.matchingHandlerNode.handlersAndTransformers) {
			c.currentHandlerOrTransformer = c.currentHandlerNode.handlersAndTransformers[c.currentHandlerOrTransformerIndex]
			c.currentHandlerOrTransformerIndex += 1
			break
		}

		// We only get here if we had a matching handler node, and we have
		// executed all of it's handlers and transformers. We can now clear the
		// matching handler node, and continue to the next handler node.
		c.matchingHandlerNode = nil
		c.currentHandlerNode = c.currentHandlerNode.next
		c.currentHandlerOrTransformerIndex = 0
		c.currentHandlerOrTransformer = nil
	}

	// If we didn't find a handler function or transformer, check for a final next
	// function, execute it, and return.
	if c.currentHandlerOrTransformer == nil {
		shouldRunFinalNext = true
		return
	}

	// Execute the handler function or transformer. Throw an error if it's not
	// an expected type.
	if currentTransformer, ok := c.currentHandlerOrTransformer.(Transformer); ok {
		execWithCtxRecovery(c, func() {
			currentTransformer.TransformRequest(c)
			c.Next()
			currentTransformer.TransformResponse(c)
		})
	} else if currentHandler, ok := c.currentHandlerOrTransformer.(Handler); ok {
		execWithCtxRecovery(c, func() {
			currentHandler.Handle(c)
		})
	} else if currentHandler, ok := c.currentHandlerOrTransformer.(HandlerFunc); ok {
		execWithCtxRecovery(c, func() {
			currentHandler(c)
		})
	} else if currentHandler, ok := c.currentHandlerOrTransformer.(func(*Context)); ok {
		execWithCtxRecovery(c, func() {
			currentHandler(c)
		})
	} else {
		panic(fmt.Sprintf("Unknown handler type: %s", reflect.TypeOf(c.currentHandlerOrTransformer)))
	}
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
	return c.url
}

// Params returns the parameters of the request.
func (c *Context) Params() RequestParams {
	return c.params
}

// Query returns the query parameters of the request.
func (c *Context) Query() url.Values {
	return c.query
}

// RequestHeaders returns the request headers.
func (c *Context) RequestHeaders() RequestHeaders {
	return c.requestHeaders
}

// RequestBodyReader returns a requestBodyReader. This is useful for streaming
// the request body, or for middleware which collects/parses the request body.
func (c *Context) RequestBodyReader() io.ReadCloser {
	return c.requestBodyReader
}

// RequestBodyBytes returns the request body bytes of the request.
func (c *Context) RequestBodyBytes() ([]byte, error) {
	if c.requestBodyBytes == nil {
		return nil, errors.New("no request body set. use RequestBodyReader() or add body parser middleware")
	}
	return c.requestBodyBytes, nil
}

// UnmarshalRequestBody unmarshals the request body into a given value. Note
// that is method requires SetRequestBodyUnmarshaller to be called first. This
// likely is done by middleware for parsing request bodies.
func (c *Context) UnmarshalRequestBody(into any) error {
	if c.requestBodyBytes == nil {
		return errors.New("no request body set. use RequestBodyReader() or add body parser middleware")
	}
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
func (c *Context) SetResponseBodyMarshaller(marshaller func(from any) ([]byte, error)) {
	c.responseBodyMarshaller = marshaller
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

// SetStatus sets the response status code.
func (c *Context) SetRequestBodyBytes(body []byte) {
	c.requestBodyBytes = body
}

// Deadline returns the deadline of the request.
func (c *Context) Deadline() (time.Time, bool) {
	ok := c.deadline != nil
	deadline := time.Time{}
	if ok {
		deadline = *c.deadline
	}
	return deadline, ok
}

// Done added for compatibility with go's context.Context. Alias for
// UntilFinish().
func (c *Context) Done() <-chan struct{} {
	doneChan := make(chan struct{})
	c.doneHandlers = append(c.doneHandlers, func() {
		doneChan <- struct{}{}
	})
	return doneChan
}

// Err returns the final error of the request. Will be nil if the request
// is still being served even if an error has occurred. Populated once the
// request is done.
func (c *Context) Err() error {
	return c.FinalError
}

// Value is a noop for compatibility with go's context.Context.
func (c *Context) Value(key any) any {
	return nil
}

// CtxSet associates a value by it's type with a context. This is for handlers
// and middleware to share data with other handlers and middleware associated
// with the context.
func CtxSet(ctx *Context, value any) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	valueType := reflect.TypeOf(value).String()

	if contextData[ctx] == nil {
		contextData[ctx] = make(map[string]any)
	}
	contextData[ctx][valueType] = value
}

// CtxGet retrieves a value by it's type from a context. This is for handlers
// and middleware to retrieve data set in association with the context by
// other handlers and middleware.
func CtxGet[V any](ctx *Context) (V, bool) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	var v V
	targetType := reflect.TypeOf(v).String()

	var target V
	contextData, ok := contextData[ctx]
	if !ok {
		return target, false
	}
	value, ok := contextData[targetType]
	if !ok {
		return target, false
	}

	return value.(V), true
}

// CtxMustGet like CtxGet retrieves a value by it's type from a context, but
// unlike CtxGet it panics if the value is not found.
func CtxMustGet[V any](ctx *Context) V {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	var v V
	targetType := reflect.TypeOf(v).String()

	contextData, ok := contextData[ctx]
	if !ok {
		panic("Context data not found for context")
	}
	value, ok := contextData[targetType]
	if !ok {
		panic(fmt.Sprintf("Context data not found for type: %s", targetType))
	}

	return value.(V)
}

func (c *Context) marshallResponseBody() ([]byte, error) {
	if c.responseBodyMarshaller == nil {
		return nil, errors.New("no response body marshaller set. use SetResponseBodyMarshaller() or add body encoder middleware")
	}
	return c.responseBodyMarshaller(c.Body)
}

func (c *Context) clearContextData() {
	delete(contextData, c)
}

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

func (c *Context) tryMatchHandlerNode(node *handlerNode) bool {
	if node.method != All && node.method != c.method {
		return false
	}

	if node.pattern != nil {
		params, ok := node.pattern.Match(c.path)
		if !ok {
			return false
		}
		c.params = params
	} else {
		c.params = make(map[string]string)
	}

	return true
}

func (c *Context) markDone() {
	c.FinalError = c.Error
	c.FinalErrorStack = c.ErrorStack
	for _, doneHandler := range c.doneHandlers {
		doneHandler()
	}
}

func execWithCtxRecovery(ctx *Context, fn func()) {
	defer func() {
		if maybeErr := recover(); maybeErr != nil {
			if err, ok := maybeErr.(error); ok {
				ctx.Error = err
			} else {
				ctx.Error = fmt.Errorf("%s", maybeErr)
			}

			stack := string(debug.Stack())
			stackLines := strings.Split(stack, "\n")
			ctx.ErrorStack = strings.Join(stackLines[6:], "\n")
		}
	}()
	fn()
}
