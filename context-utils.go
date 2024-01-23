package navaros

import (
	"fmt"
	"reflect"
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

// CtxSet associates a value by it's type with a context. This is for handlers
// and middleware to share data with other handlers and middleware associated
// with the context.
func CtxSet(ctx *Context, value any) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	valueType := reflect.TypeOf(value).String()

	if contextData[ctx] == nil {
		contextData[ctx] = make(map[string]any)
	}
	contextData[ctx][valueType] = value
}

// CtxGet retrieves a value by it's type from a context. This is for handlers
// and middleware to retrieve data set in association with the context by
// other handlers and middleware.
func CtxGet[V any](ctx *Context) (V, bool) {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	var v V
	targetType := reflect.TypeOf(v).String()

	var target V
	contextData, ok := contextData[ctx]
	if !ok {
		return target, false
	}
	value, ok := contextData[targetType]
	if !ok {
		return target, false
	}

	return value.(V), true
}

// CtxMustGet like CtxGet retrieves a value by it's type from a context, but
// unlike CtxGet it panics if the value is not found.
func CtxMustGet[V any](ctx *Context) V {
	for ctx.parentContext != nil {
		ctx = ctx.parentContext
	}

	var v V
	targetType := reflect.TypeOf(v).String()

	contextData, ok := contextData[ctx]
	if !ok {
		panic("Context data not found for context")
	}
	value, ok := contextData[targetType]
	if !ok {
		panic(fmt.Sprintf("Context data not found for type: %s", targetType))
	}

	return value.(V)
}
