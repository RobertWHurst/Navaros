package navaros

import (
	"time"
)

// CtxGetParam enables tests to set parameters on a context. This is needed
// because tests often use NewContext to create a context to use with the
// handler being tested. Because the context wasn't matched against a real
// path, it will have no parameters. This function allows tests to
// set parameters on the context.
func CtxSetParam(ctx *Context, key, value string) {
	ctx.params[key] = value
}

// CtxDeleteParam allows tests to delete a parameter from the context.
func CtxDeleteParam(ctx *Context, key string) {
	delete(ctx.params, key)
}

// CtxFinalize allows libraries to call the finalize method on a context.
// finalize figures out the final status code, headers, and body for the
// response. This is normally called by the router, but can be called by
// libraries which wish to extend or encapsulate the functionality of Navaros.
func CtxFinalize(ctx *Context) {
	ctx.finalize()
}

// CtxFree allows libraries to free a context they created manually.
// This is important, DO NOT FORGET TO CALL THIS IF YOU CREATE A CONTEXT MANUALLY.
func CtxFree(ctx *Context) {
	ctx.free()
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
