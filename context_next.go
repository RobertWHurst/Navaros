package navaros

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"time"
)

// Next calls the next handler in the chain. This is useful for creating
// middleware style handlers that work on the context before and/or after the
// responding handler.
func (c *Context) Next() {
	// In the case that this is a sub context, we need to update the parent
	// context with the current context's state.
	defer c.tryUpdateParent()

	if c.Error != nil {
		return
	}

	// If we have exceeded the deadline, we can return early.
	if c.deadline != nil && time.Now().After(*c.deadline) {
		c.Error = errors.New("request exceeded timeout deadline")
		return
	}

	// If we're mid-wrap (a wrap handler called ctx.Next()), advance the wrap
	// index and proceed to the next wrap or the real handler. Skip chain
	// walking.
	if c.currentHandlerOrTransformer != nil {
		if c.currentWrapHandlerIndex < len(c.wrapHandlers) {
			c.currentWrapHandler = c.wrapHandlers[c.currentWrapHandlerIndex]
			c.currentWrapHandlerIndex++
		}
	} else {
		// walk the chain looking for a handler with a pattern that matches the method
		// and path of the request, or until we reach the end of the chain
		for c.currentHandlerNode != nil {

			// Because handlers can have multiple handler functions or transformers,
			// we may save a matching handler node to the context so that we can
			// continue from the same handler until we have executed all of it's
			// handlers and transformers.
			//
			// If we do not have a matching handler node, we will walk the chain
			// until we find a matching handler node.
			if !c.currentHandlerNodeMatches {
				for c.currentHandlerNode != nil {
					if len(c.currentHandlerNode.WrapHandlers) > 0 && len(c.currentHandlerNode.HandlersAndTransformers) == 0 {
						c.wrapHandlers = append(c.wrapHandlers, c.currentHandlerNode.WrapHandlers...)
						c.currentHandlerNode = c.currentHandlerNode.Next
						continue
					}
					if c.currentHandlerNode.tryMatch(c) {
						c.currentHandlerNodeMatches = true
						break
					}
					c.currentHandlerNode = c.currentHandlerNode.Next
				}
				if !c.currentHandlerNodeMatches {
					break
				}
			}

			// Grab a handler function or transformer from the matching handler node.
			// If there are more than one, we will continue from the same handler node
			// the next time Next is called. We iterate through the handler functions
			// and transformers until we have executed all of them.
			if c.currentHandlerOrTransformerIndex < len(c.currentHandlerNode.HandlersAndTransformers) {
				c.currentHandlerOrTransformer = c.currentHandlerNode.HandlersAndTransformers[c.currentHandlerOrTransformerIndex]
				c.currentHandlerOrTransformerIndex += 1
				c.currentWrapHandlerIndex = 0

				// If there are wrap handlers on the context, start the wrap chain.
				if c.currentWrapHandlerIndex < len(c.wrapHandlers) {
					c.currentWrapHandler = c.wrapHandlers[c.currentWrapHandlerIndex]
					c.currentWrapHandlerIndex++
				}

				break
			}

			// We only get here if we had a matching handler node, and we have
			// executed all of it's handlers and transformers. We can now clear the
			// matching handler node, and continue to the next handler node.
			c.currentHandlerNodeMatches = false
			c.currentHandlerNode = c.currentHandlerNode.Next
			c.currentHandlerOrTransformerIndex = 0
			c.currentHandlerOrTransformer = nil
			c.currentWrapHandlerIndex = 0
		}
	}

	// If we didn't find a handler function or transformer and we have reached
	// the end of the chain, we can return early.
	if c.currentHandlerOrTransformer == nil && c.currentWrapHandler == nil {
		c.nextBeyondEnd = true
		return
	}

	// If there is a pending wrap handler, execute it. The wrap handler calls
	// ctx.Next() to proceed to the next wrap or the actual handler.
	if c.currentWrapHandler != nil {
		wrapHandler := c.currentWrapHandler
		c.currentWrapHandler = nil
		execWithCtxRecovery(c, func() {
			wrapHandler(c)
		})
		return
	}

	// Clear the current handler before executing so that if the handler calls
	// ctx.Next(), the chain advances normally instead of re-entering the wrap
	// dispatch.
	handlerOrTransformer := c.currentHandlerOrTransformer
	c.currentHandlerOrTransformer = nil

	// Execute the handler function or transformer. Throw an error if it's not
	// an expected type.
	if currentTransformer, ok := handlerOrTransformer.(Transformer); ok {
		execWithCtxRecovery(c, func() {
			currentTransformer.TransformRequest(c)
			c.Next()
			currentTransformer.TransformResponse(c)
		})
	} else if currentHandler, ok := handlerOrTransformer.(Handler); ok {
		execWithCtxRecovery(c, func() {
			currentHandler.Handle(c)
		})
	} else if currentHandler, ok := handlerOrTransformer.(HandlerFunc); ok {
		execWithCtxRecovery(c, func() {
			currentHandler(c)
		})
	} else if currentHandler, ok := handlerOrTransformer.(func(*Context)); ok {
		execWithCtxRecovery(c, func() {
			currentHandler(c)
		})
	} else if _, ok := handlerOrTransformer.(func(res http.ResponseWriter, req *http.Request)); ok {
		panic("http.HandlerFunc are not yet supported")
	} else {
		panic(fmt.Sprintf("Unknown handler type: %s", reflect.TypeOf(handlerOrTransformer)))
	}

	c.currentHandlerNodeMatches = false
	c.currentHandlerNode = nil
	c.currentHandlerOrTransformerIndex = 0
	c.currentWrapHandlerIndex = 0
}

func execWithCtxRecovery(ctx *Context, fn func()) {
	defer func() {
		if maybeErr := recover(); maybeErr != nil {
			if err, ok := maybeErr.(error); ok {
				ctx.Error = err
			} else {
				ctx.Error = fmt.Errorf("%s", maybeErr)
			}

			stack := string(debug.Stack())
			stackLines := strings.Split(stack, "\n")
			ctx.ErrorStack = strings.Join(stackLines[6:], "\n")
		}
	}()
	fn()
}
