package json_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/json"
)

type testRequest struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type testResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func TestMiddleware_RequestUnmarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(json.Middleware(nil))

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

	reqBody := `{"name":"test","value":42}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"message":"ok"`) {
		t.Errorf("expected response to contain message, got %q", body)
	}
}

func TestMiddleware_ResponseMarshalling(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(json.Middleware(nil))

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
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"message":"hello"`) {
		t.Errorf("expected response to contain message, got %q", body)
	}
	if !strings.Contains(body, `"success":true`) {
		t.Errorf("expected response to contain success, got %q", body)
	}
}

// String responses are passed through as-is in Navaros

func TestMiddleware_ErrorResponse(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(json.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Body = json.Error("something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"error":"something went wrong"`) {
		t.Errorf("expected error message, got %q", body)
	}
}

func TestMiddleware_FieldErrorResponse(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(json.Middleware(nil))

	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Body = json.FieldError{Field: "email", Error: "invalid format"}
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"error":"Validation error"`) {
		t.Errorf("expected validation error, got %q", body)
	}
	if !strings.Contains(body, `"email"`) {
		t.Errorf("expected field name, got %q", body)
	}
}

func TestMiddleware_NonJSONContentType(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(json.Middleware(nil))

	router.Post("/test", func(ctx *navaros.Context) {
		bodyReader := ctx.RequestBodyReader()
		bodyBytes, _ := io.ReadAll(bodyReader)

		ctx.Status = http.StatusOK
		ctx.Body = string(bodyBytes)
	})

	req := httptest.NewRequest("POST", "/test", strings.NewReader("plain text"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_RequestUnmarshallingWithContentTypeParameters(t *testing.T) {
	// Clients that name a charset — Go's own net/http, OkHttp, axios — send
	// "application/json; charset=utf-8". RFC 9110 section 8.3 makes parameters
	// no part of a media type's identity, so these requests must unmarshal too.
	headers := []string{
		"application/json; charset=utf-8",
		"application/json;charset=UTF-8",
		"Application/JSON",
		"application/vnd.api+json; charset=utf-8",
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			router := navaros.NewRouter()
			router.Use(json.Middleware(nil))

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

			reqBody := `{"name":"test","value":42}`
			req := httptest.NewRequest("POST", "/test", strings.NewReader(reqBody))
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
		"application/xml",
		"application/octet-stream, x+json",
		"application/+json",
		"*/*",
	}

	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			router := navaros.NewRouter()
			router.Use(json.Middleware(nil))

			router.Post("/test", func(ctx *navaros.Context) {
				var req testRequest
				if err := ctx.UnmarshalRequestBody(&req); err == nil {
					t.Errorf("Content-Type %q installed an unmarshaller; it names no JSON media type", header)
				}

				ctx.Status = http.StatusOK
			})

			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","value":42}`))
			req.Header.Set("Content-Type", header)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status 200, got %d", w.Code)
			}
		})
	}
}
