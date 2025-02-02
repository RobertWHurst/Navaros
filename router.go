package navaros

import (
	"net/http"
	"reflect"
	"strings"
)

var PrintHandlerErrors = false

// SetPrintHandlerErrors toggles the printing of handler errors.
func SetPrintHandlerErrors(enable bool) {
	PrintHandlerErrors = enable
}

// Router is the main component of Navaros. It is an HTTP handler that can be
// used to handle requests, and route them to the appropriate handlers. It
// implements the http.Handler interface, and can be used as a handler in
// standard http servers. It also implements Navaros' own Handler interface,
// which allows nesting routers for better code organization.
type Router struct {
	routeDescriptorMap map[HTTPMethod]map[string]bool
	routeDescriptors   []*RouteDescriptor
	firstHandlerNode   *HandlerNode
	lastHandlerNode    *HandlerNode
}

// NewRouter creates a new router.
func NewRouter() *Router {
	return &Router{}
}

// ServeHTTP allows the router to be used as a handler in standard go http
// servers. It handles the incoming request - creating a context and running
// the handler chain over it, then finalizing the response.
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := NewContextWithNode(res, req, r.firstHandlerNode)
	ctx.Next()
	ctx.finalize()
	ctx.free()
}

// Handle is for the purpose of taking an existing context, and running it
// through the mux's handler chain. If the last handler calls next, it
// will call next on the original context.
func (r *Router) Handle(ctx *Context) {
	subCtx := NewSubContextWithNode(ctx, r.firstHandlerNode)
	subCtx.Next()
	subCtx.free()
	if subCtx.currentHandlerNode == nil {
		ctx.Next()
	}
}

// RouteDescriptors returns a list of all the route descriptors that this
// router is responsible for. Useful for gateway configuration.
func (r *Router) RouteDescriptors() []*RouteDescriptor {
	return r.routeDescriptors
}

// Use is for middleware handlers. It allows the handlers to be executed on
// every request. If a path is provided, the middleware will only be executed
// on requests that match the path.
//
// Note that routers are middleware handlers, and so can be passed to Use to
// attach them as sub-routers. It's also important to know that if you provide
// a path with a router, the router will set the mount path as the base path
// for all of it's handlers. This means that if you have a router with a path
// of "/foo", and you bind a handler with a path of "/bar", the handler will
// only be executed on requests with a path of "/foo/bar".
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

	r.bind(false, All, mountPath, handlersAndTransformers...)
}

// All allows binding handlers to all HTTP methods at a given route path
// pattern.
func (r *Router) All(path string, handlersAndTransformers ...any) {
	r.bind(false, All, path, handlersAndTransformers...)
}

// Post allows binding handlers to the POST HTTP method at a given route path
// pattern.
func (r *Router) Post(path string, handlersAndTransformers ...any) {
	r.bind(false, Post, path, handlersAndTransformers...)
}

// Get allows binding handlers to the GET HTTP method at a given route path
// pattern.
func (r *Router) Get(path string, handlersAndTransformers ...any) {
	r.bind(false, Get, path, handlersAndTransformers...)
}

// Put allows binding handlers to the PUT HTTP method at a given route path
// pattern.
func (r *Router) Put(path string, handlersAndTransformers ...any) {
	r.bind(false, Put, path, handlersAndTransformers...)
}

// Patch allows binding handlers to the PATCH HTTP method at a given route path
// pattern.
func (r *Router) Patch(path string, handlersAndTransformers ...any) {
	r.bind(false, Patch, path, handlersAndTransformers...)
}

// Delete allows binding handlers to the DELETE HTTP method at a given route
// path pattern.
func (r *Router) Delete(path string, handlersAndTransformers ...any) {
	r.bind(false, Delete, path, handlersAndTransformers...)
}

// Options allows binding handlers to the OPTIONS HTTP method at a given
// route path pattern.
func (r *Router) Options(path string, handlersAndTransformers ...any) {
	r.bind(false, Options, path, handlersAndTransformers...)
}

// Head allows binding handlers to the HEAD HTTP method at a given route
// path pattern.
func (r *Router) Head(path string, handlersAndTransformers ...any) {
	r.bind(false, Head, path, handlersAndTransformers...)
}

// PublicAll is the same as All, but it also adds the route descriptor to the
// router's list of public route descriptors.
func (r *Router) PublicAll(path string, handlersAndTransformers ...any) {
	r.bind(true, All, path, handlersAndTransformers...)
}

// PublicPost is the same as Post, but it also adds the route descriptor to the
// router's list of public route descriptors.
func (r *Router) PublicPost(path string, handlersAndTransformers ...any) {
	r.bind(true, Post, path, handlersAndTransformers...)
}

// PublicGet is the same as Get, but it also adds the route descriptor to the
// router's list of public route descriptors.
func (r *Router) PublicGet(path string, handlersAndTransformers ...any) {
	r.bind(true, Get, path, handlersAndTransformers...)
}

// PublicPut is the same as Put, but it also adds the route descriptor to the
// router's list of public route descriptors.
func (r *Router) PublicPut(path string, handlersAndTransformers ...any) {
	r.bind(true, Put, path, handlersAndTransformers...)
}

// PublicPatch is the same as Patch, but it also adds the route descriptor to
// the router's list of public route descriptors.
func (r *Router) PublicPatch(path string, handlersAndTransformers ...any) {
	r.bind(true, Patch, path, handlersAndTransformers...)
}

// PublicDelete is the same as Delete, but it also adds the route descriptor
// to the router's list of public route descriptors.
func (r *Router) PublicDelete(path string, handlersAndTransformers ...any) {
	r.bind(true, Delete, path, handlersAndTransformers...)
}

// PublicOptions is the same as Options, but it also adds the route descriptor
// to the router's list of public route descriptors.
func (r *Router) PublicOptions(path string, handlersAndTransformers ...any) {
	r.bind(true, Options, path, handlersAndTransformers...)
}

// PublicHead is the same as Head, but it also adds the route descriptor to the
// router's list of public route descriptors.
func (r *Router) PublicHead(path string, handlersAndTransformers ...any) {
	r.bind(true, Head, path, handlersAndTransformers...)
}

// bind creates a pattern object from the route pattern as well as a handler
// node. It then attaches the new link to the end of the router's handler
// chain.
func (r *Router) bind(isPublic bool, method HTTPMethod, path string, handlersAndTransformers ...any) {
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

		panic("invalid handler type. Must be a Transformer, Handler, or " +
			"HandlerFunc. Got: " + reflect.TypeOf(handlerOrTransformer).String())
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
		} else if isPublic && !hasAddedOwnRouteDescriptor {
			r.addRouteDescriptor(method, pattern)
			hasAddedOwnRouteDescriptor = true
		}
	}

	nextHandlerNode := &HandlerNode{
		Method:                  method,
		Pattern:                 pattern,
		HandlersAndTransformers: handlersAndTransformers,
	}

	if r.firstHandlerNode == nil {
		r.firstHandlerNode = nextHandlerNode
		r.lastHandlerNode = nextHandlerNode
	} else {
		r.lastHandlerNode.Next = nextHandlerNode
		r.lastHandlerNode = nextHandlerNode
	}
}

// addRouteDescriptor adds a route descriptor to the router's list of route
// descriptors, but only if it doesn't already exist.
func (r *Router) addRouteDescriptor(method HTTPMethod, pattern *Pattern) {
	path := pattern.String()
	if r.routeDescriptorMap == nil {
		r.routeDescriptorMap = map[HTTPMethod]map[string]bool{}
	}
	if r.routeDescriptors == nil {
		r.routeDescriptors = []*RouteDescriptor{}
	}
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
