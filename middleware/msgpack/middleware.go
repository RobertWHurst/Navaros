package msgpack

import (
	"bytes"
	"io"

	"github.com/RobertWHurst/navaros"
	"github.com/vmihailenco/msgpack/v5"
)

type Options struct {
	DisableRequestBodyUnmarshaller bool
	DisableResponseBodyMarshaller  bool
}

func Middleware(options *Options) func(ctx *navaros.Context) {
	if options == nil {
		options = &Options{}
	}

	return func(ctx *navaros.Context) {
		if !options.DisableRequestBodyUnmarshaller {
			unmarshalRequestBody(ctx)
		}

		if !options.DisableResponseBodyMarshaller {
			marshalResponseBody(ctx)
		}

		ctx.Next()
	}
}

func unmarshalRequestBody(ctx *navaros.Context) {
	contentType := ctx.RequestHeaders().Get("Content-Type")
	if contentType != "application/msgpack" {
		return
	}

	requestBodyReader := ctx.RequestBodyReader()
	requestBodyBytes, err := io.ReadAll(requestBodyReader)
	if err != nil {
		ctx.Error = err
		return
	}

	ctx.SetRequestBodyUnmarshaller(func(into any) error {
		return msgpack.Unmarshal(requestBodyBytes, into)
	})
}

func marshalResponseBody(ctx *navaros.Context) {
	ctx.SetResponseBodyMarshaller(func(from any) (io.Reader, error) {
		if from != nil {
			ctx.Headers.Add("Content-Type", "application/msgpack")
		}

		switch v := from.(type) {

		case []FieldError:
			if ctx.Status == 0 {
				ctx.Status = 400
			}
			from = M{
				"error":  "Validation error",
				"fields": genFieldsField(v),
			}

		case FieldError:
			if ctx.Status == 0 {
				ctx.Status = 400
			}
			from = M{
				"error":  "Validation error",
				"fields": genFieldsField([]FieldError{v}),
			}

		case Error:
			if ctx.Status == 0 {
				ctx.Status = 400
			}
			from = M{"error": string(v)}

		case string:
			from = M{"message": v}
		}

		msgpackBytes, err := msgpack.Marshal(from)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(msgpackBytes), nil
	})
}

func genFieldsField(errors []FieldError) []M {
	var fields []M
	for _, err := range errors {
		field := M{}
		field[err.Field] = err.Error
		fields = append(fields, field)
	}
	return fields
}
