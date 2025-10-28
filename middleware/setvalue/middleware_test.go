package setvalue_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/setvalue"
)

type Config struct {
	MaxRetries int
	Timeout    int
}

func TestMiddleware(t *testing.T) {
	config := &Config{MaxRetries: 3, Timeout: 30}
	router := navaros.NewRouter()
	router.Use(setvalue.Middleware("config", config))

	router.Get("/test", func(ctx *navaros.Context) {
		cfg := ctx.MustGet("config")
		if cfg == nil {
			t.Error("expected config to be set")
		}

		configVal := cfg.(Config)
		if configVal.MaxRetries != 3 {
			t.Errorf("expected MaxRetries 3, got %d", configVal.MaxRetries)
		}
		if configVal.Timeout != 30 {
			t.Errorf("expected Timeout 30, got %d", configVal.Timeout)
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
}

func TestMiddlewareStoresValueNotPointer(t *testing.T) {
	timeout := 30
	router := navaros.NewRouter()
	router.Use(setvalue.Middleware("timeout", &timeout))

	router.Get("/test", func(ctx *navaros.Context) {
		val := ctx.MustGet("timeout").(int)
		if val != 60 {
			t.Errorf("expected current dereferenced value 60, got %d", val)
		}

		ctx.Status = http.StatusOK
		ctx.Body = "ok"
	})

	// Change the value after middleware is created
	timeout = 60

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
