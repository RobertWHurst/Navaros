package navaros

type Transformer interface {
	TransformRequest(ctx *Context)
	TransformResponse(ctx *Context)
}

type handlerNode struct {
	method                  HttpVerb
	pattern                 *Pattern
	handlersAndTransformers []any
	next                    *handlerNode
}

type HandlerFunc func(ctx *Context)

type Handler interface {
	Handle(ctx *Context)
}
