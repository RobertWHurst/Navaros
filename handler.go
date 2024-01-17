package navaros

type Transformer interface {
	TransformRequest(ctx *Context)
	TransformResponse(ctx *Context)
}

type handlerNode struct {
	method                  HTTPMethod
	pattern                 *Pattern
	handlersAndTransformers []any
	next                    *handlerNode
}

type HandlerFunc func(ctx *Context)

type Handler interface {
	Handle(ctx *Context)
}

type RouterHandler interface {
	RouteDescriptors() []*RouteDescriptor
	Handle(ctx *Context)
}
