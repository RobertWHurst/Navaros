package navaros

import (
	"context"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestContextDoneChannelRace tests that accessing Done() concurrently with
// context cleanup doesn't cause a data race. This simulates the scenario where
// user code passes the context to a goroutine (like an S3 upload) that may
// still access the context after the handler returns.
func TestContextDoneChannelRace(t *testing.T) {
	router := NewRouter()

	var wg sync.WaitGroup
	router.Get("/test", func(ctx *Context) {
		wg.Add(1)
		// Spawn a goroutine that will access the context after the handler returns
		go func() {
			defer wg.Done()
			// Use the context as a context.Context (like AWS SDK does)
			var stdCtx context.Context = ctx

			// Keep accessing Done() in a loop to increase chance of catching races
			for i := 0; i < 100; i++ {
				select {
				case <-stdCtx.Done():
					return
				default:
					time.Sleep(time.Microsecond)
				}
			}
		}()

		// Handler returns immediately, which will trigger context cleanup
		ctx.Status = 200
	})

	// Make multiple requests to increase chance of catching races
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
	}

	// Wait for all goroutines to complete
	wg.Wait()
}

// TestContextDoneChannelConcurrentAccess tests that multiple goroutines
// can safely call Done() on the same context concurrently.
func TestContextDoneChannelConcurrentAccess(t *testing.T) {
	router := NewRouter()

	var wg sync.WaitGroup
	router.Get("/test", func(ctx *Context) {
		// Spawn multiple goroutines that all access Done() concurrently
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var stdCtx context.Context = ctx
				// Just get the Done channel multiple times to test concurrent access
				for j := 0; j < 100; j++ {
					ch := stdCtx.Done()
					if ch == nil {
						t.Error("Done() returned nil channel")
					}
				}
			}()
		}

		ctx.Status = 200
	})

	req := httptest.NewRequest("GET", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	// Wait for all goroutines to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for goroutines to complete")
	}
}
