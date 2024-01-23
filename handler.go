package navaros

// Transformer is a special type of handler object that can be used to
// transform the context before and after handlers have processed the request.
// This is most useful for modifying or re-encoding the request and response
// bodies.
type Transformer interface {
	TransformRequest(ctx *Context)
	TransformResponse(ctx *Context)
}

// HandlerNode is used to build the handler chains used by the context. The
// router builds a chain from these objects then attaches them to the context.
// It then calls Next on the context to execute the chain.
type HandlerNode struct {
	Method                  HTTPMethod
	Pattern                 *Pattern
	HandlersAndTransformers []any
	Next                    *HandlerNode
}

// Handler is a handler object interface. Any object that implements this
// interface can be used as a handler in a handler chain.
type Handler interface {
	Handle(ctx *Context)
}

// RouterHandler is handled nearly identically to a Handler, but it also
// provides a list of route descriptors which are collected by the router.
// These will be merged with the other route descriptors already collected.
// This use for situation where a handler may do more sub-routing, and the
// allows the handler to report the sub-routes to the router, rather than
// it's base path.
type RouterHandler interface {
	RouteDescriptors() []*RouteDescriptor
	Handle(ctx *Context)
}
