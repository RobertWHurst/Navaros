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
var _ http.Flusher = &ContextResponseWriter{}
var _ http.Hijacker = &ContextResponseWriter{}

func (c *ContextResponseWriter) Header() http.Header {
	return c.bodyWriter.Header()
}

func (c *ContextResponseWriter) WriteHeader(status int) {
	c.ctx.Status = status
	c.flushHeaders()
}

func (c *ContextResponseWriter) Write(bytes []byte) (int, error) {
	c.ctx.hasWrittenBody = true
	c.flushHeaders()
	return c.bodyWriter.Write(bytes)
}

func (c *ContextResponseWriter) Flush() {
	if f, ok := c.bodyWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (c *ContextResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := c.bodyWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (c *ContextResponseWriter) flushHeaders() {
	if c.ctx.hasWrittenHeaders {
		return
	}
	c.ctx.hasWrittenHeaders = true

	for key, values := range c.ctx.Headers {
		for _, value := range values {
			c.bodyWriter.Header().Add(key, value)
		}
	}
	for _, cookie := range c.ctx.Cookies {
		http.SetCookie(c.bodyWriter, cookie)
	}

	if c.ctx.inhibitResponse {
		return
	}

	status := c.ctx.Status
	if status == 0 {
		status = http.StatusOK
	}
	c.bodyWriter.WriteHeader(status)
}
