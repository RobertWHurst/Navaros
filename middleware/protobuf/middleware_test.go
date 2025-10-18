package protobuf_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/protobuf"
	"google.golang.org/protobuf/proto"
)

func TestMiddleware_RequestUnmarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(protobuf.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		var req protobuf.TestRequest
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
		ctx.Body = &protobuf.TestResponse{Message: "ok", Success: true}
	})

	reqData := &protobuf.TestRequest{Name: "test", Value: 42}
	reqBody, _ := proto.Marshal(reqData)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/protobuf")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp protobuf.TestResponse
	if err := proto.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "ok" {
		t.Errorf("expected message 'ok', got %q", resp.Message)
	}
}

func TestMiddleware_ResponseMarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(protobuf.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = http.StatusOK
		ctx.Body = &protobuf.TestResponse{Message: "hello", Success: true}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/protobuf" {
		t.Errorf("expected Content-Type application/protobuf, got %q", contentType)
	}

	var resp protobuf.TestResponse
	if err := proto.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "hello" {
		t.Errorf("expected message 'hello', got %q", resp.Message)
	}
	if !resp.Success {
		t.Errorf("expected success true")
	}
}

func TestMiddleware_AlternateContentType(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(protobuf.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		var req protobuf.TestRequest
		if err := ctx.UnmarshalRequestBody(&req); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}

		ctx.Status = http.StatusOK
		ctx.Body = &req
	})

	reqData := &protobuf.TestRequest{Name: "test", Value: 42}
	reqBody, _ := proto.Marshal(reqData)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/x-protobuf")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_InvalidProtoType(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(protobuf.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		var req string
		if err := ctx.UnmarshalRequestBody(&req); err == nil {
			t.Error("expected error for non-proto type")
		} else if err.Error() != "value must implement proto.Message (generated protobuf struct)" {
			t.Errorf("unexpected error message: %v", err)
		}
		ctx.Status = http.StatusBadRequest
		ctx.Body = "error"
	})

	reqData := &protobuf.TestRequest{Name: "test", Value: 42}
	reqBody, _ := proto.Marshal(reqData)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/protobuf")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
