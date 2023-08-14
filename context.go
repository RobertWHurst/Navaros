package navaros

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
)

type Context struct {
	parentContext *Context

	method            HttpVerb
	path              string
	url               *url.URL
	params            RequestParams
	requestHeaders    RequestHeaders
	requestBodyReader io.ReadCloser
	requestBodyBytes  []byte

	Status            int
	Headers           ResponseHeaders
	Body              any
	bodyWriter        http.ResponseWriter
	hasWrittenHeaders bool
	hasWrittenBody    bool

	Error error

	requestBodyUnmarshaller func(into any) error
	responseBodyMarshaller  func(from any) ([]byte, error)

	currentHandlerNode               *handlerNode
	matchingHandlerNode              *handlerNode
	currentHandlerOrTransformerIndex int
	currentHandlerOrTransformer      any
}

var contextData = make(map[*Context]map[string]any)

func NewContext(responseWriter http.ResponseWriter, request *http.Request, firstHandlerNode *handlerNode) *Context {
	return &Context{
		method:         HttpVerb(request.Method),
		path:           request.URL.Path,
		url:            request.URL,
		requestHeaders: RequestHeaders(request.Header),

		Headers:    map[string]string{},
		bodyWriter: responseWriter,

		currentHandlerNode: &handlerNode{
			method:                  All,
			handlersAndTransformers: []any{},
			next:                    firstHandlerNode,
		},
	}
}

func NewSubContext(ctx *Context, firstHandlerNode *handlerNode) *Context {
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
	return &subContext
}

func (c *Context) Next() {
	if c.Error != nil {
		return
	}

	for c.currentHandlerNode != nil {

		// Iterate the chain until we find a matching handler node
		if c.matchingHandlerNode == nil {
			for c.currentHandlerNode != nil {
				if c.tryMatchHandlerNode(c.currentHandlerNode) {
					c.matchingHandlerNode = c.currentHandlerNode
					break
				}
				c.currentHandlerNode = c.currentHandlerNode.next
			}
			if c.matchingHandlerNode == nil {
				return
			}
		}

		if c.currentHandlerOrTransformerIndex < len(c.matchingHandlerNode.handlersAndTransformers) {
			c.currentHandlerOrTransformer = c.currentHandlerNode.handlersAndTransformers[c.currentHandlerOrTransformerIndex]
			c.currentHandlerOrTransformerIndex += 1
			break
		}

		c.matchingHandlerNode = nil
		c.currentHandlerNode = c.currentHandlerNode.next
		c.currentHandlerOrTransformerIndex = 0
		c.currentHandlerOrTransformer = nil
	}

	if c.currentHandlerOrTransformer == nil {
		return
	}

	if currentTransformer, ok := c.currentHandlerOrTransformer.(Transformer); ok {
		execWithCtxRecovery(c, func() {
			currentTransformer.TransformRequest(c)
			c.Next()
			currentTransformer.TransformResponse(c)
		})
	} else if currentMux, ok := c.currentHandlerOrTransformer.(*Mux); ok {
		execWithCtxRecovery(c, func() {
			currentMux.Handle(c)
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

	c.tryUpdateParent()
}

func (c *Context) Method() HttpVerb {
	return c.method
}

func (c *Context) Path() string {
	return c.path
}

func (c *Context) Url() *url.URL {
	return c.url
}

func (c *Context) Params() RequestParams {
	return c.params
}

func (c *Context) RequestHeaders() RequestHeaders {
	return c.requestHeaders
}

func (c *Context) RequestBodyReader() io.ReadCloser {
	return c.requestBodyReader
}

func (c *Context) RequestBodyBytes() ([]byte, error) {
	if c.requestBodyBytes == nil {
		return nil, errors.New("no request body set. use RequestBodyReader() or add body parser middleware")
	}
	return c.requestBodyBytes, nil
}

func (c *Context) UnmarshalRequestBody(into any) error {
	if c.requestBodyBytes == nil {
		return errors.New("no request body set. use RequestBodyReader() or add body parser middleware")
	}
	if c.requestBodyUnmarshaller == nil {
		return errors.New("no request body unmarshaller set. use SetRequestBodyUnmarshaller() or add body parser middleware")
	}
	return c.requestBodyUnmarshaller(into)
}

func (c *Context) MarshallResponseBody() ([]byte, error) {
	if c.responseBodyMarshaller == nil {
		return nil, errors.New("no response body marshaller set. use SetResponseBodyMarshaller() or add body encoder middleware")
	}
	return c.responseBodyMarshaller(c.Body)
}

func (c *Context) SetRequestBodyUnmarshaller(unmarshaller func(into any) error) {
	c.requestBodyUnmarshaller = unmarshaller
}

func (c *Context) SetResponseBodyMarshaller(marshaller func(from any) ([]byte, error)) {
	c.responseBodyMarshaller = marshaller
}

func (c *Context) Write(bytes []byte) (int, error) {
	if !c.hasWrittenHeaders {
		if c.Status == 0 {
			c.Status = 200
		}
		c.bodyWriter.WriteHeader(c.Status)
		c.hasWrittenHeaders = true
	}

	c.hasWrittenBody = true
	return c.bodyWriter.Write(bytes)
}

func (c *Context) Flush() {
	c.bodyWriter.(http.Flusher).Flush()
}

func (c *Context) SetRequestBodyBytes(body []byte) {
	c.requestBodyBytes = body
}

func CtxSet(ctx *Context, value any) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	vr := reflect.ValueOf(value)

	var vt reflect.Type
	if vr.Kind() == reflect.Ptr {
		vt = vr.Elem().Type()
	} else {
		vt = vr.Type()
	}

	valueType := vt.PkgPath() + "." + vt.Name()
	if contextData[ctx] == nil {
		contextData[ctx] = make(map[string]any)
	}
	contextData[ctx][valueType] = value
}

func CtxGet[V any](ctx *Context) (V, bool) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	var target V
	tr := reflect.ValueOf(target)
	targetType := tr.Elem().Type().PkgPath() + "." + tr.Elem().Type().Name()

	contextData, ok := contextData[ctx]
	if !ok {
		return target, false
	}
	value, ok := contextData[targetType]
	if !ok {
		return target, false
	}

	vr := reflect.ValueOf(value)
	tr.Elem().Set(vr.Elem())

	return target, true
}

func CtxMustGet[V any](ctx *Context) V {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	fmt.Println("1")
	var tt reflect.Type
	if reflect.TypeOf((*V)(nil)).Kind() == reflect.Ptr {
		fmt.Println("2")
		tt = reflect.TypeOf((*V)(nil)).Elem()
	} else {
		fmt.Println("3")
		tt = reflect.TypeOf((*V)(nil))
	}
	fmt.Println("4")

	targetType := tt.PkgPath() + "." + tt.Name()

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

func execWithCtxRecovery(ctx *Context, fn func()) {
	defer func() {
		if maybeErr := recover(); maybeErr != nil {
			if err, ok := maybeErr.(error); ok {
				ctx.Error = err
			} else {
				ctx.Error = fmt.Errorf("%s", maybeErr)
			}
		}
	}()
	fn()
}

func (c *Context) clearContextData() {
	delete(contextData, c)
}

func (c *Context) tryUpdateParent() {
	if c.parentContext == nil {
		return
	}

	// Copy the current context to the parent context while preserving the parent
	// context's routing state, and it's parent context.

	ancestorContext := c.parentContext.parentContext
	currentHandlerNode := c.parentContext.currentHandlerNode
	matchingHandlerNode := c.parentContext.matchingHandlerNode
	currentHandlerOrTransformerIndex := c.parentContext.currentHandlerOrTransformerIndex
	currentHandlerOrTransformer := c.parentContext.currentHandlerOrTransformer

	*c.parentContext = *c

	c.parentContext.parentContext = ancestorContext
	c.parentContext.currentHandlerNode = currentHandlerNode
	c.parentContext.matchingHandlerNode = matchingHandlerNode
	c.parentContext.currentHandlerOrTransformerIndex = currentHandlerOrTransformerIndex
	c.parentContext.currentHandlerOrTransformer = currentHandlerOrTransformer
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
