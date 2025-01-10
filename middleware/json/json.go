package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/RobertWHurst/navaros"
)

type Options struct {
	disableRequestBodyUnmarshaller bool
	disableResponseBodyMarshaller  bool
}

func Middleware(options *Options) func(ctx *navaros.Context) {
	if options == nil {
		options = &Options{}
	}

	return func(ctx *navaros.Context) {
		if !options.disableRequestBodyUnmarshaller {
			unmarshalRequestBody(ctx)
		}

		if !options.disableResponseBodyMarshaller {
			marshalResponseBody(ctx)
		}

		ctx.Next()
	}
}

func unmarshalRequestBody(ctx *navaros.Context) {
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

	ctx.SetRequestBodyUnmarshaller(func(into any) error {
		return json.Unmarshal(requestBodyBytes, into)
	})
}

func marshalResponseBody(ctx *navaros.Context) {
	ctx.SetResponseBodyMarshaller(func(from any) (io.Reader, error) {
		if from != nil {
			ctx.Headers.Add("Content-Type", "application/json")
		}
		switch str := from.(type) {
		case E:
			if ctx.Status == 0 {
				ctx.Status = 400
			}
			from = map[string]string{"error": string(str)}
		case string:
			from = map[string]string{"message": str}
		}
		jsonBytes, err := json.Marshal(from)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(jsonBytes), nil
	})
}
