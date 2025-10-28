package set_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/set"
)

func TestMiddleware(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(set.Middleware("apiVersion", "v1"))
	router.Use(set.Middleware("config", map[string]int{"timeout": 30}))

	router.Get("/test", func(ctx *navaros.Context) {
		version := ctx.MustGet("apiVersion")
		if version == nil {
			t.Error("expected apiVersion to be set")
		}
		if version != "v1" {
			t.Errorf("expected 'v1', got %v", version)
		}

		config := ctx.MustGet("config")
		if config == nil {
			t.Error("expected config to be set")
		}
		configMap := config.(map[string]int)
		if configMap["timeout"] != 30 {
			t.Errorf("expected timeout 30, got %d", configMap["timeout"])
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

func TestMiddlewareValueDoesNotPersistAcrossRequests(t *testing.T) {
	router := navaros.NewRouter()
	router.Use(set.Middleware("counter", 0))

	router.Get("/increment", func(ctx *navaros.Context) {
		counter := ctx.MustGet("counter").(int)
		ctx.Set("counter", counter+1)
		ctx.Status = http.StatusOK
		ctx.Body = "incremented"
	})

	router.Get("/check", func(ctx *navaros.Context) {
		counter := ctx.MustGet("counter").(int)
		if counter != 0 {
			t.Errorf("expected counter to be 0 (reset), got %d", counter)
		}
		ctx.Status = http.StatusOK
		ctx.Body = "ok"
	})

	// First request - increment
	req1 := httptest.NewRequest("GET", "/increment", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Second request - check that counter is reset
	req2 := httptest.NewRequest("GET", "/check", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
}
