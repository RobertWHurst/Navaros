package navaros

import (
	"bufio"
	"net"
	"net/http"
)

type ContextResponseWriter struct {
	ctx        *Context
	bodyWriter http.ResponseWriter
}

var _ http.ResponseWriter = &ContextResponseWriter{}
var _ http.Hijacker = &ContextResponseWriter{}

func (c *ContextResponseWriter) Write(bytes []byte) (int, error) {
	c.ctx.hasWrittenBody = true
	c.ctx.flushHeaders()
	return c.bodyWriter.Write(bytes)
}

func (c *ContextResponseWriter) WriteHeader(status int) {
	c.ctx.Status = status
	c.ctx.flushHeaders()
}

func (c *ContextResponseWriter) Header() http.Header {
	return c.bodyWriter.Header()
}

func (c *ContextResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := c.bodyWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}
