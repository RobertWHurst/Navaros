package navaros_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
)

func BenchmarkPatternMatch(b *testing.B) {
	pattern, err := navaros.NewPattern("/users/:id/posts/:postId")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pattern.Match("/users/123/posts/456")
	}
}

func BenchmarkPatternMatchInto(b *testing.B) {
	pattern, err := navaros.NewPattern("/users/:id/posts/:postId")
	if err != nil {
		b.Fatal(err)
	}
	params := navaros.RequestParams{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pattern.MatchInto("/users/123/posts/456", &params)
	}
}

func BenchmarkPatternMatchStatic(b *testing.B) {
	pattern, err := navaros.NewPattern("/users/123/posts/456")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pattern.Match("/users/123/posts/456")
	}
}

func BenchmarkPatternMatchWildcard(b *testing.B) {
	pattern, err := navaros.NewPattern("/api/*")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pattern.Match("/api/v1/users/123")
	}
}

func BenchmarkContextGetSet(b *testing.B) {
	router := navaros.NewRouter()
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Set("key", "value")
		_ = ctx.MustGet("key")
		ctx.Status = http.StatusOK
		ctx.Body = "ok"
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRequestRouting(b *testing.B) {
	router := navaros.NewRouter()
	router.Get("/users/:id", func(ctx *navaros.Context) {
		ctx.Status = http.StatusOK
		ctx.Body = ctx.Params().Get("id")
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkRequestRoutingWithMiddleware(b *testing.B) {
	router := navaros.NewRouter()
	router.Use(func(ctx *navaros.Context) {
		ctx.Set("middleware", "active")
		ctx.Next()
	})
	router.Get("/users/:id", func(ctx *navaros.Context) {
		ctx.Status = http.StatusOK
		ctx.Body = ctx.Params().Get("id")
	})

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPatternCompilation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = navaros.NewPattern("/users/:id/posts/:postId/comments/:commentId")
	}
}

func BenchmarkParamsExtraction(b *testing.B) {
	router := navaros.NewRouter()
	router.Get("/users/:userId/posts/:postId", func(ctx *navaros.Context) {
		_ = ctx.Params().Get("userId")
		_ = ctx.Params().Get("postId")
		ctx.Status = http.StatusOK
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}

func BenchmarkMiddlewareChain(b *testing.B) {
	router := navaros.NewRouter()
	router.Use(func(ctx *navaros.Context) {
		ctx.Set("m1", true)
		ctx.Next()
	})
	router.Use(func(ctx *navaros.Context) {
		ctx.Set("m2", true)
		ctx.Next()
	})
	router.Use(func(ctx *navaros.Context) {
		ctx.Set("m3", true)
		ctx.Next()
	})
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = http.StatusOK
		ctx.Body = "ok"
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, req)
	}
}
