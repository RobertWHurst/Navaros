package navaros

type Transformer interface {
	TransformRequest(ctx *Context)
	TransformResponse(ctx *Context)
}

type HandlerNode struct {
	Method                  HTTPMethod
	Pattern                 *Pattern
	HandlersAndTransformers []any
	Next                    *HandlerNode
}

type Handler interface {
	Handle(ctx *Context)
}

type RouterHandler interface {
	RouteDescriptors() []*RouteDescriptor
	Handle(ctx *Context)
}
