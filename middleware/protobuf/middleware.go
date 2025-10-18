package protobuf

import (
	"bytes"
	"errors"
	"io"

	"github.com/RobertWHurst/navaros"
	"google.golang.org/protobuf/proto"
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
	if contentType != "application/protobuf" && contentType != "application/x-protobuf" {
		return
	}

	requestBodyReader := ctx.RequestBodyReader()
	requestBodyBytes, err := io.ReadAll(requestBodyReader)
	if err != nil {
		ctx.Error = err
		return
	}

	ctx.SetRequestBodyUnmarshaller(func(into any) error {
		protoMsg, ok := into.(proto.Message)
		if !ok {
			return errors.New("value must implement proto.Message (generated protobuf struct)")
		}
		return proto.Unmarshal(requestBodyBytes, protoMsg)
	})
}

func marshalResponseBody(ctx *navaros.Context) {
	ctx.SetResponseBodyMarshaller(func(from any) (io.Reader, error) {
		protoMsg, ok := from.(proto.Message)
		if !ok {
			return nil, errors.New("value must implement proto.Message (generated protobuf struct)")
		}

		if from != nil {
			ctx.Headers.Add("Content-Type", "application/protobuf")
		}

		protoBytes, err := proto.Marshal(protoMsg)
		if err != nil {
			return nil, err
		}
		return bytes.NewBuffer(protoBytes), nil
	})
}
