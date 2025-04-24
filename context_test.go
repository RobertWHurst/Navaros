package navaros_test

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RobertWHurst/navaros"
)

func TestNewContext(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx == nil {
		t.Error("expected context")
	}
}

func TestContextHandler(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	handlerCalled := false
	var ctx *navaros.Context
	ctx = navaros.NewContext(res, req, func(givenCtx *navaros.Context) {
		handlerCalled = true

		if ctx != givenCtx {
			t.Error("expected the same context to be passed to the handler")
		}
	})
	ctx.Next()

	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

func TestContextHandlerAfterTimeout(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	handlerCalled := false
	ctx := navaros.NewContext(res, req, func(_ *navaros.Context) {
		handlerCalled = true
	})

	navaros.CtxSetDeadline(ctx, time.Now().Add(-1*time.Second))
	ctx.Next()

	if handlerCalled {
		t.Error("expected handler to not be called")
	}
	if ctx.Error == nil {
		t.Error("expected context to have an error")
	}
}

func TestContextWithPanic(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/a/b/c", bytes.NewBuffer([]byte("test")))

	ctx := navaros.NewContext(res, req, func(_ *navaros.Context) {
		panic("test panic")
	})

	ctx.Next()

	if ctx.Error == nil {
		t.Error("expected context to have an error")
	}
	if ctx.Error.Error() != "test panic" {
		t.Error("expected context to have the correct error")
	}
}

type testTransformer struct {
	transformRequestCalled  bool
	transformResponseCalled bool
}

func (t *testTransformer) TransformRequest(ctx *navaros.Context) {
	t.transformRequestCalled = true

	body, err := io.ReadAll(ctx.RequestBodyReader())
	if err != nil {
		ctx.Error = err
		return
	}
	newBody := bytes.NewBuffer([]byte(string(body) + " transformed"))

	ctx.SetRequestBodyReader(newBody)
}

func (t *testTransformer) TransformResponse(ctx *navaros.Context) {
	ctx.RequestBodyReader()

	t.transformResponseCalled = true
}

func TestContextWithTransformer(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	transformer := &testTransformer{
		transformRequestCalled:  false,
		transformResponseCalled: false,
	}

	ctx := navaros.NewContext(res, req, transformer)
	ctx.Next()

	if !transformer.transformRequestCalled {
		t.Error("expected transformer to be called")
	}
	if !transformer.transformResponseCalled {
		t.Error("expected transformer to be called")
	}
}

func TestContextMethod(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx.Method() != navaros.Get {
		t.Error("expected method to be GET")
	}
}

func TestContextPath(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://roberthurst.ca/a/b/c?z=yz", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx.Path() != "/a/b/c" {
		t.Error("expected path to be /a/b/c")
	}
}

func TestContextURL(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://roberthurst.ca/a/b/c?z=yz", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx.URL().Host != "roberthurst.ca" {
		t.Error("expected url host to be roberthurst.ca")
	}
	if ctx.URL().Path != "/a/b/c" {
		t.Error("expected url path to be /a/b/c")
	}
	if ctx.URL().RawQuery != "z=yz" {
		t.Error("expected url raw query to be z=yz")
	}
}

func TestContextParams(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	if len(ctx.Params()) != 0 {
		t.Error("expected params to be empty because no routes are registered")
	}
}

func TestContextQuery(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c?z=yz", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx.Query().Get("z") != "yz" {
		t.Error("expected query to have z=yz")
	}
}

func TestContextProtocol(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.Proto = "HTTP/1.2"

	ctx := navaros.NewContext(res, req, nil)

	if ctx.Protocol() != "HTTP/1.2" {
		t.Error("expected request protocol to be our made up HTTP/1.2 protocol")
	}
}

func TestContextProtocolMajor(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.ProtoMajor = 2

	ctx := navaros.NewContext(res, req, nil)

	if ctx.ProtocolMajor() != 2 {
		t.Error("expected request protocol major to be 2")
	}
}

func TestContextProtocolMinor(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.ProtoMinor = 2

	ctx := navaros.NewContext(res, req, nil)

	if ctx.ProtocolMinor() != 2 {
		t.Error("expected request protocol minor to be 2")
	}
}

func TestContextRequestHeaders(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.Header.Add("test", "test")

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestHeaders().Get("test") != "test" {
		t.Error("expected request headers to have test=test")
	}
}

func TestContextRequestBodyReader(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", bytes.NewBuffer([]byte("test")))

	ctx := navaros.NewContext(res, req, nil)

	body, err := io.ReadAll(ctx.RequestBodyReader())
	if err != nil {
		t.Error("expected no error")
	}
	if string(body) != "test" {
		t.Error("expected body to be test")
	}
}

func TestContextSetRequestBodyReader(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", bytes.NewBuffer([]byte("test")))

	ctx := navaros.NewContext(res, req, nil)

	body, err := io.ReadAll(ctx.RequestBodyReader())
	if err != nil {
		t.Error("expected no error")
	}
	if string(body) != "test" {
		t.Error("expected body to be test")
	}

	ctx.SetRequestBodyReader(bytes.NewBuffer([]byte("test2")))
	body, err = io.ReadAll(ctx.RequestBodyReader())
	if err != nil {
		t.Error("expected no error")
	}
	if string(body) != "test2" {
		t.Error("expected body to be test2")
	}
}

func TestContextRequestBodyUnmarshaller(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", bytes.NewBuffer([]byte(`{"test":"test"}`)))

	ctx := navaros.NewContext(res, req, nil)
	ctx.SetRequestBodyUnmarshaller(func(into any) error {
		bodyBytes, err := io.ReadAll(ctx.RequestBodyReader())
		if err != nil {
			return err
		}
		return json.Unmarshal(bodyBytes, into)
	})

	var body struct {
		Test string `json:"test"`
	}
	if err := ctx.UnmarshalRequestBody(&body); err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	if body.Test != "test" {
		t.Error("expected body to be test")
	}
}

func TestContextSetResponseBodyMarshaller(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	ctx.SetResponseBodyMarshaller(func(from any) (io.Reader, error) {
		return nil, nil
	})
}

func TestContextRequestContentLength(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", bytes.NewBuffer([]byte("test")))

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestContentLength() != 4 {
		t.Error("expected content length to be 4")
	}
}

func TestContextRequestTransferEncoding(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.TransferEncoding = []string{"test"}

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestTransferEncoding()[0] != "test" {
		t.Error("expected transfer encoding to be test")
	}
}

func TestContextRequestHost(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.Host = "test"

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestHost() != "test" {
		t.Error("expected host to be test")
	}
}

func TestContextRequestRemoteAddress(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.RemoteAddr = "test"

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestRemoteAddress() != "test" {
		t.Error("expected remote address to be test")
	}
}

func TestContextRequestRawURI(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.RequestURI = "test"

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestRawURI() != "test" {
		t.Error("expected raw uri to be test")
	}
}

func TestContextRequestTLS(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)
	req.TLS = &tls.ConnectionState{
		Version: tls.VersionTLS13,
	}

	ctx := navaros.NewContext(res, req, nil)

	if ctx.RequestTLS().Version != tls.VersionTLS13 {
		t.Error("expected tls version to be tls.VersionTLS13")
	}
}

func TestContextRequest(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	if ctx.Request() != req {
		t.Error("expected request to be the same")
	}
}

func TestContextResponseWriter(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	resWriter := ctx.ResponseWriter()
	resWriter.WriteHeader(200)
	resWriter.Write([]byte("test"))

	if res.Code != 200 {
		t.Error("expected response code to be 200")
	}
	if res.Body.String() != "test" {
		t.Error("expected response body to be test")
	}
}

func TestContextWrite(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)
	if _, err := ctx.Write([]byte("test")); err != nil {
		t.Error("expected no error")
	}

	if res.Body.String() != "test" {
		t.Error("expected body to be test")
	}
}

func TestContextFlush(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)
	ctx.Flush()

	if !res.Flushed {
		t.Error("expected flushed to be true")
	}
}

func TestContextDeadline(t *testing.T) {
	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/b/c", nil)

	ctx := navaros.NewContext(res, req, nil)

	deadline := time.Now()
	navaros.CtxSetDeadline(ctx, deadline)

	ctxDeadline, ctxHasDeadline := ctx.Deadline()
	if !ctxHasDeadline {
		t.Error("expected deadline to be set")
	}
	if ctxDeadline != deadline {
		t.Error("expected deadline to be the one provided")
	}
}
