package navaros

import (
	"time"
)

// CtxFinalize allows libraries to call the finalize method on a context.
// finalize figures out the final status code, headers, and body for the
// response. This is nomally called by the router, but can be called by
// libraries which wish to extend or encapsulate the functionality of Navaros.
func CtxFinalize(ctx *Context) {
	ctx.finalize()
}

// CtxSetDeadline sets a deadline for the context. This allows libraries to
// limit the amount of time a handler can take to process a request.
func CtxSetDeadline(ctx *Context, deadline time.Time) {
	ctx.deadline = &deadline
}
