package setfn

import "github.com/RobertWHurst/navaros"

func Middleware[V any](key string, valueFn func() V) func(ctx *navaros.Context) {
	return func(ctx *navaros.Context) {
		ctx.Set(key, valueFn())
		ctx.Next()
	}
}
