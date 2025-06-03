package navaros

// HandlerNode is used to build the handler chains used by the context. The
// router builds a chain from these objects then attaches them to the context.
// It then calls Next on the context to execute the chain.
type HandlerNode struct {
	Method                  HTTPMethod
	Pattern                 *Pattern
	HandlersAndTransformers []any
	Next                    *HandlerNode
}

// tryMatch attempts to match the handler node's route pattern and http
// method to the a context. It will return true if the handler node
// matches, and false if it does not.
func (n *HandlerNode) tryMatch(ctx *Context) bool {
	if n.Method != All && n.Method != ctx.method {
		return false
	}
	if n.Pattern == nil {
		return true
	}
	return n.Pattern.MatchInto(ctx.path, &ctx.params)
}
