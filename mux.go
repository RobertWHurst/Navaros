package navaros

import (
	"fmt"
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

	finalBody := make([]byte, 0)
	if !ctx.hasWrittenBody {
		switch body := ctx.Body.(type) {
		case string:
			finalBody = []byte(body)
		case []byte:
			finalBody = body
		default:
			marshalledBytes, err := ctx.marshallResponseBody()
			if err != nil {
				ctx.Status = 500
				finalBody = []byte(err.Error())
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

	if len(finalBody) != 0 {
		if _, err := ctx.bodyWriter.Write(finalBody); err != nil {
			fmt.Printf("Error occurred when writing response: %s", err)
		}
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
