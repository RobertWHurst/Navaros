package navaros

import (
	"net/http"
)

type Mux struct {
	firstHandlerNode *handlerNode
	lastHandlerNode  *handlerNode
}

func New() *Mux {
	return &Mux{}
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r, m.firstHandlerNode)

	ctx.Next()

	if ctx.Error != nil {
		ctx.Status = 500
		ctx.Body = ctx.Error.Error()
	}

	if ctx.Status == 0 {
		if ctx.hasWrittenBody {
			ctx.Status = 200
		} else {
			ctx.Status = 404
		}
	}

	if !ctx.hasWrittenHeaders {
		for key, value := range ctx.Headers {
			ctx.bodyWriter.Header().Set(key, value)
		}
		ctx.bodyWriter.WriteHeader(ctx.Status)
	}

	if !ctx.hasWrittenBody {
		finalBodyBytes := make([]byte, 0)
		if ctx.Body != nil {
			switch body := ctx.Body.(type) {
			case string:
				finalBodyBytes = []byte(body)
			case []byte:
				finalBodyBytes = body
			default:
				marshalledBytes, err := ctx.marshallResponseBody()
				if err != nil {
					ctx.Error = err
				}
				finalBodyBytes = marshalledBytes
			}
		}
		ctx.bodyWriter.Write(finalBodyBytes)
	}

	ctx.clearContextData()
}

func (m *Mux) Handle(ctx *Context) {
	subCtx := NewSubContext(ctx, m.firstHandlerNode)
	subCtx.Next()
	ctx.Next()
}

func (m *Mux) Use(handlersAndTransformers ...any) {
	m.bind(All, "/**", handlersAndTransformers...)
}

func (m *Mux) All(path string, handlersAndTransformers ...any) {
	m.bind(All, path, handlersAndTransformers...)
}

func (m *Mux) Post(path string, handlersAndTransformers ...any) {
	m.bind(Post, path, handlersAndTransformers...)
}

func (m *Mux) Get(path string, handlersAndTransformers ...any) {
	m.bind(Get, path, handlersAndTransformers...)
}

func (m *Mux) Put(path string, handlersAndTransformers ...any) {
	m.bind(Put, path, handlersAndTransformers...)
}

func (m *Mux) Patch(path string, handlersAndTransformers ...any) {
	m.bind(Patch, path, handlersAndTransformers...)
}

func (m *Mux) Delete(path string, handlersAndTransformers ...any) {
	m.bind(Delete, path, handlersAndTransformers...)
}

func (m *Mux) Options(path string, handlersAndTransformers ...any) {
	m.bind(Options, path, handlersAndTransformers...)
}

func (m *Mux) Head(path string, handlersAndTransformers ...any) {
	m.bind(Head, path, handlersAndTransformers...)
}

func (m *Mux) bind(method HttpVerb, path string, handlersAndTransformers ...any) {
	pattern, err := NewPattern(path)
	if err != nil {
		panic(err)
	}

	nextHandlerNode := handlerNode{
		method:                  method,
		pattern:                 pattern,
		handlersAndTransformers: handlersAndTransformers,
	}

	if m.firstHandlerNode == nil {
		m.firstHandlerNode = &nextHandlerNode
		m.lastHandlerNode = &nextHandlerNode
	} else {
		m.lastHandlerNode.next = &nextHandlerNode
		m.lastHandlerNode = &nextHandlerNode
	}
}
