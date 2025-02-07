package setvalue

import "github.com/RobertWHurst/navaros"

func Middleware[V *any](key string, value V) func(ctx *navaros.Context) {
	return func(ctx *navaros.Context) {
		ctx.Set(key, *value)
		ctx.Next()
	}
}
