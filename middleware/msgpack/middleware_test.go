package msgpack_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/msgpack"
	msgpacklib "github.com/vmihailenco/msgpack/v5"
)

type testRequest struct {
	Name  string `msgpack:"name"`
	Value int    `msgpack:"value"`
}

type testResponse struct {
	Message string `msgpack:"message"`
	Success bool   `msgpack:"success"`
}

func TestMiddleware_RequestUnmarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(msgpack.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		var req testRequest
		if err := ctx.UnmarshalRequestBody(&req); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		if req.Name != "test" {
			t.Errorf("expected name 'test', got %q", req.Name)
		}
		if req.Value != 42 {
			t.Errorf("expected value 42, got %d", req.Value)
		}

		ctx.Status = http.StatusOK
		ctx.Body = testResponse{Message: "ok", Success: true}
	})

	reqData := testRequest{Name: "test", Value: 42}
	reqBody, _ := msgpacklib.Marshal(reqData)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/msgpack")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp testResponse
	if err := msgpacklib.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "ok" {
		t.Errorf("expected message 'ok', got %q", resp.Message)
	}
}

func TestMiddleware_ResponseMarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(msgpack.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = http.StatusOK
		ctx.Body = testResponse{Message: "hello", Success: true}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/msgpack" {
		t.Errorf("expected Content-Type application/msgpack, got %q", contentType)
	}

	var resp testResponse
	if err := msgpacklib.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", resp.Message)
	}
	if !resp.Success {
		t.Errorf("expected success true")
	}
}

func TestMiddleware_ErrorResponse(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(msgpack.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Body = msgpack.Error("something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]any
	if err := msgpacklib.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "something went wrong" {
		t.Errorf("expected error message, got %v", resp)
	}
}

func TestMiddleware_FieldErrorResponse(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(msgpack.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Body = msgpack.FieldError{Field: "email", Error: "invalid format"}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var resp map[string]any
	if err := msgpacklib.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["error"] != "Validation error" {
		t.Errorf("expected validation error, got %v", resp)
	}
}

func TestMiddleware_AlternateContentType(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(msgpack.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		var req testRequest
		if err := ctx.UnmarshalRequestBody(&req); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		ctx.Status = http.StatusOK
		ctx.Body = req
	})

	reqData := testRequest{Name: "test", Value: 42}
	reqBody, _ := msgpacklib.Marshal(reqData)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/msgpack")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_RequestUnmarshallingWithContentTypeParameters(t *testing.T) {
	// Parameters are not part of a media type's identity (RFC 9110 section 8.3),
	// and type and subtype are case-insensitive.
	headers := []string{
		"application/msgpack; charset=binary",
		"Application/MsgPack",
		"application/vnd.custom+msgpack",
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			router := navaros.NewRouter()
			router.Use(msgpack.Middleware(nil))

			router.Post("/test", func(ctx *navaros.Context) {
				var req testRequest
				if err := ctx.UnmarshalRequestBody(&req); err != nil {
					t.Errorf("failed to unmarshal: %v", err)
					return
				}
				if req.Name != "test" || req.Value != 42 {
					t.Errorf("expected {test 42}, got %+v", req)
				}

				ctx.Status = http.StatusOK
				ctx.Body = testResponse{Message: "ok", Success: true}
			})

			reqBody, _ := msgpacklib.Marshal(testRequest{Name: "test", Value: 42})
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestMiddleware_NonMatchingContentTypeInstallsNoUnmarshaller(t *testing.T) {
	// The mirror of the test above, and the one that actually pins the contract:
	// a value that does not name this media type must leave the unmarshaller
	// uninstalled. Asserting only the positive direction passes even if the
	// comparison matches everything.
	headers := []string{
		"text/plain",
		"application/json",
		"application/protobuf, evil+msgpack",
		"application/+msgpack",
		"*/*",
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			router := navaros.NewRouter()
			router.Use(msgpack.Middleware(nil))

			router.Post("/test", func(ctx *navaros.Context) {
				var req testRequest
				if err := ctx.UnmarshalRequestBody(&req); err == nil {
					t.Errorf("Content-Type %q installed an unmarshaller; it names no msgpack media type", header)
				}

				ctx.Status = http.StatusOK
			})

			reqBody, _ := msgpacklib.Marshal(testRequest{Name: "test", Value: 42})
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}
