package json

import (
	"encoding/json"
	"io"

	"github.com/robertwhurst/navaros"
)

type Options struct {
	disableRequestBodyUnmarshaller bool
	disableResponseBodyMarshaller  bool
}

func Middleware(options Options) navaros.HandlerFunc {
	return func(ctx *navaros.Context) {
		if !options.disableRequestBodyUnmarshaller {
			contentType := ctx.RequestHeaders().Get("Content-Type")
			if contentType != "application/json" {
				ctx.Next()
				return
			}

			requestBodyReader := ctx.RequestBodyReader()
			requestBodyBytes, err := io.ReadAll(requestBodyReader)
			if err != nil {
				ctx.Error = err
				return
			}
			ctx.SetRequestBodyBytes(requestBodyBytes)

			ctx.SetRequestBodyUnmarshaller(func(into any) error {
				return json.Unmarshal(requestBodyBytes, into)
			})
		}

		if !options.disableResponseBodyMarshaller {
			ctx.SetResponseBodyMarshaller(func(from any) ([]byte, error) {
				return json.Marshal(from)
			})
		}

		ctx.Next()
	}
}
