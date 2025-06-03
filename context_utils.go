package navaros

import (
	"time"
)

// CtxFinalize allows libraries to call the finalize method on a context.
// finalize figures out the final status code, headers, and body for the
// response. This is normally called by the router, but can be called by
// libraries which wish to extend or encapsulate the functionality of Navaros.
func CtxFinalize(ctx *Context) {
	ctx.finalize()
}

// CtxSetDeadline sets a deadline for the context. This allows libraries to
// limit the amount of time a handler can take to process a request.
func CtxSetDeadline(ctx *Context, deadline time.Time) {
	ctx.deadline = &deadline
}

// CtxInhibitResponse prevents the context from sending a response. This is
// useful for libraries which wish to handle the response themselves. For example,
// a upgraded websocket connection cannot have response headers written to it.
func CtxInhibitResponse(ctx *Context) {
	ctx.inhibitResponse = true
}
