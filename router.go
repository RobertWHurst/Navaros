package navaros

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

var PrintHandlerErrors = false

type Router struct {
	routeDescriptorMap map[HTTPMethod]map[string]bool
	routeDescriptors   []*RouteDescriptor
	firstHandlerNode   *handlerNode
	lastHandlerNode    *handlerNode
}

func NewRouter() *Router {
	return &Router{
		routeDescriptorMap: map[HTTPMethod]map[string]bool{},
		routeDescriptors:   []*RouteDescriptor{},
	}
}

func (r *Router) ServeHTTP(wtr http.ResponseWriter, req *http.Request) {
	ctx := NewContext(wtr, req, r.firstHandlerNode)

	ctx.Next()

	if ctx.Error != nil {
		ctx.Status = 500
		if PrintHandlerErrors {
			fmt.Printf("Error occurred when handling request: %s\n%s", ctx.Error, ctx.ErrorStack)
		}
	}

	finalBody := make([]byte, 0)
	if !ctx.hasWrittenBody && ctx.Body != nil {
		switch body := ctx.Body.(type) {
		case string:
			finalBody = []byte(body)
		case []byte:
			finalBody = body
		default:
			marshalledBytes, err := ctx.marshallResponseBody()
			if err != nil {
				ctx.Status = 500
				fmt.Printf("Error occurred when marshalling response body: %s", err)
			} else {
				finalBody = marshalledBytes
			}
		}
	}

	if ctx.Status == 0 {
		if len(finalBody) == 0 {
			ctx.Status = 404
		} else {
			ctx.Status = 200
		}
	}

	if !ctx.hasWrittenHeaders {
		for key, value := range ctx.Headers {
			ctx.bodyWriter.Header().Set(key, value)
		}
		ctx.bodyWriter.WriteHeader(ctx.Status)
	}

	hasBody := len(finalBody) != 0
	is100Range := ctx.Status >= 100 && ctx.Status < 200
	is204Or304 := ctx.Status == 204 || ctx.Status == 304

	if hasBody {
		if is100Range || is204Or304 {
			fmt.Printf("Response with status %d has body but no content is expected", ctx.Status)
		} else if _, err := ctx.bodyWriter.Write(finalBody); err != nil {
			fmt.Printf("Error occurred when writing response: %s", err)
		}
	}

	ctx.clearContextData()
	ctx.markDone()
}

// Handle is for the purpose of taking an existing context, and running it
// through the mux's handler chain. If the last handler calls next, it
// will call next on the original context.
func (r *Router) Handle(ctx *Context) {
	subCtx := NewSubContext(ctx, r.firstHandlerNode, func() {
		ctx.Next()
	})
	subCtx.Next()
}

// RouteDescriptors returns a list of all the route descriptors that this
// router is responsible for. Useful for gateway configuration.
func (r *Router) RouteDescriptors() []*RouteDescriptor {
	return r.routeDescriptors
}

func (r *Router) Use(handlersAndTransformers ...any) {
	mountPath := "/**"
	if len(handlersAndTransformers) != 0 {
		if customMountPath, ok := handlersAndTransformers[0].(string); ok {
			if !strings.HasSuffix(customMountPath, "/**") {
				customMountPath = strings.TrimSuffix(customMountPath, "/")
				customMountPath += "/**"
			}
			mountPath = customMountPath
			handlersAndTransformers = handlersAndTransformers[1:]
		}
	}
	r.bind(All, mountPath, handlersAndTransformers...)
}

func (r *Router) All(path string, handlersAndTransformers ...any) {
	r.bind(All, path, handlersAndTransformers...)
}

func (r *Router) Post(path string, handlersAndTransformers ...any) {
	r.bind(Post, path, handlersAndTransformers...)
}

func (r *Router) Get(path string, handlersAndTransformers ...any) {
	r.bind(Get, path, handlersAndTransformers...)
}

func (r *Router) Put(path string, handlersAndTransformers ...any) {
	r.bind(Put, path, handlersAndTransformers...)
}

func (r *Router) Patch(path string, handlersAndTransformers ...any) {
	r.bind(Patch, path, handlersAndTransformers...)
}

func (r *Router) Delete(path string, handlersAndTransformers ...any) {
	r.bind(Delete, path, handlersAndTransformers...)
}

func (r *Router) Options(path string, handlersAndTransformers ...any) {
	r.bind(Options, path, handlersAndTransformers...)
}

func (r *Router) Head(path string, handlersAndTransformers ...any) {
	r.bind(Head, path, handlersAndTransformers...)
}

func (r *Router) bind(method HTTPMethod, path string, handlersAndTransformers ...any) {
	if len(handlersAndTransformers) == 0 {
		panic("no handlers or transformers provided")
	}

	pattern, err := NewPattern(path)
	if err != nil {
		panic(err)
	}

	for _, handlerOrTransformer := range handlersAndTransformers {
		if _, ok := handlerOrTransformer.(Transformer); ok {
			continue
		} else if _, ok := handlerOrTransformer.(Handler); ok {
			continue
		} else if _, ok := handlerOrTransformer.(HandlerFunc); ok {
			continue
		} else if _, ok := handlerOrTransformer.(func(*Context)); ok {
			continue
		}

		handlerOrTransformerRefType := reflect.TypeOf(handlerOrTransformer)
		panic(fmt.Errorf("invalid handler or transformer type: %s", handlerOrTransformerRefType.String()))
	}

	hasAddedOwnRouteDescriptor := false
	for _, handlerOrTransformer := range handlersAndTransformers {
		if routerHandler, ok := handlerOrTransformer.(RouterHandler); ok {
			for _, routeDescriptor := range routerHandler.RouteDescriptors() {
				mountPath := strings.TrimSuffix(path, "/**")
				subPattern, err := NewPattern(mountPath + routeDescriptor.Pattern.String())
				if err != nil {
					panic(err)
				}
				r.addRouteDescriptor(routeDescriptor.Method, subPattern)
			}
		} else if !hasAddedOwnRouteDescriptor {
			r.addRouteDescriptor(method, pattern)
			hasAddedOwnRouteDescriptor = true
		}
	}

	nextHandlerNode := handlerNode{
		method:                  method,
		pattern:                 pattern,
		handlersAndTransformers: handlersAndTransformers,
	}

	if r.firstHandlerNode == nil {
		r.firstHandlerNode = &nextHandlerNode
		r.lastHandlerNode = &nextHandlerNode
	} else {
		r.lastHandlerNode.next = &nextHandlerNode
		r.lastHandlerNode = &nextHandlerNode
	}
}

func (r *Router) addRouteDescriptor(method HTTPMethod, pattern *Pattern) {
	path := pattern.String()
	if _, ok := r.routeDescriptorMap[method]; !ok {
		r.routeDescriptorMap[method] = map[string]bool{}
	}
	if _, ok := r.routeDescriptorMap[method][path]; ok {
		return
	}
	r.routeDescriptorMap[method][path] = true
	r.routeDescriptors = append(r.routeDescriptors, &RouteDescriptor{
		Method:  method,
		Pattern: pattern,
	})
}
