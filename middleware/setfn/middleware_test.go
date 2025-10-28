package setfn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/setfn"
)

func TestMiddleware(t *testing.T) {
	callCount := 0
	router := navaros.NewRouter()
	router.Use(setfn.Middleware("requestID", func() string {
		callCount++
		return "req-123"
	}))

	router.Get("/test", func(ctx *navaros.Context) {
		requestID := ctx.MustGet("requestID")
		if requestID == nil {
			t.Error("expected requestID to be set")
		}
		if requestID != "req-123" {
			t.Errorf("expected 'req-123', got %v", requestID)
		}

		ctx.Status = http.StatusOK
		ctx.Body = "ok"
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	if callCount != 1 {
		t.Errorf("expected function to be called once, got %d", callCount)
	}
}

func TestMiddlewareFunctionCalledPerRequest(t *testing.T) {
	callCount := 0
	router := navaros.NewRouter()
	router.Use(setfn.Middleware("counter", func() int {
		callCount++
		return callCount
	}))

	router.Get("/test", func(ctx *navaros.Context) {
		counter := ctx.MustGet("counter").(int)
		ctx.Status = http.StatusOK
		ctx.Body = counter
	})

	// First request
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Second request - function should be called again
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if callCount != 2 {
		t.Errorf("expected function to be called twice, got %d", callCount)
	}
}
