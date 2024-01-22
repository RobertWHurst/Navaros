package navaros

import (
	"net/http"
	"reflect"
	"strings"
)

var PrintHandlerErrors = false

type Router struct {
	routeDescriptorMap map[HTTPMethod]map[string]bool
	routeDescriptors   []*RouteDescriptor
	firstHandlerNode   *HandlerNode
	lastHandlerNode    *HandlerNode
}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := NewContextWithNode(res, req, r.firstHandlerNode)
	ctx.Next()
	ctx.finalize()
}

// Handle is for the purpose of taking an existing context, and running it
// through the mux's handler chain. If the last handler calls next, it
// will call next on the original context.
func (r *Router) Handle(ctx *Context) {
	subCtx := NewSubContextWithNode(ctx, r.firstHandlerNode)
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

	r.bind(false, All, mountPath, handlersAndTransformers...)
}

func (r *Router) PublicAll(path string, handlersAndTransformers ...any) {
	r.bind(true, All, path, handlersAndTransformers...)
}

func (r *Router) PublicPost(path string, handlersAndTransformers ...any) {
	r.bind(true, Post, path, handlersAndTransformers...)
}

func (r *Router) PublicGet(path string, handlersAndTransformers ...any) {
	r.bind(true, Get, path, handlersAndTransformers...)
}

func (r *Router) PublicPut(path string, handlersAndTransformers ...any) {
	r.bind(true, Put, path, handlersAndTransformers...)
}

func (r *Router) PublicPatch(path string, handlersAndTransformers ...any) {
	r.bind(true, Patch, path, handlersAndTransformers...)
}

func (r *Router) PublicDelete(path string, handlersAndTransformers ...any) {
	r.bind(true, Delete, path, handlersAndTransformers...)
}

func (r *Router) PublicOptions(path string, handlersAndTransformers ...any) {
	r.bind(true, Options, path, handlersAndTransformers...)
}

func (r *Router) PublicHead(path string, handlersAndTransformers ...any) {
	r.bind(true, Head, path, handlersAndTransformers...)
}

func (r *Router) All(path string, handlersAndTransformers ...any) {
	r.bind(false, All, path, handlersAndTransformers...)
}

func (r *Router) Post(path string, handlersAndTransformers ...any) {
	r.bind(false, Post, path, handlersAndTransformers...)
}

func (r *Router) Get(path string, handlersAndTransformers ...any) {
	r.bind(false, Get, path, handlersAndTransformers...)
}

func (r *Router) Put(path string, handlersAndTransformers ...any) {
	r.bind(false, Put, path, handlersAndTransformers...)
}

func (r *Router) Patch(path string, handlersAndTransformers ...any) {
	r.bind(false, Patch, path, handlersAndTransformers...)
}

func (r *Router) Delete(path string, handlersAndTransformers ...any) {
	r.bind(false, Delete, path, handlersAndTransformers...)
}

func (r *Router) Options(path string, handlersAndTransformers ...any) {
	r.bind(false, Options, path, handlersAndTransformers...)
}

func (r *Router) Head(path string, handlersAndTransformers ...any) {
	r.bind(false, Head, path, handlersAndTransformers...)
}

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
		} else if _, ok := handlerOrTransformer.(func(*Context)); ok {
			continue
		}

		panic(
			"invalid handler type. Must be a Transformer, Handler, or " +
				"func(*Context). Got: " + reflect.TypeOf(handlerOrTransformer).String(),
		)
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
