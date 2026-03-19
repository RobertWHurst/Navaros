package navaros_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RobertWHurst/navaros"
)

func TestNewRouter(t *testing.T) {
	navaros.NewRouter()
}

func TestRouterGetSimpleHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 201
		ctx.Body = "Hello World"
	})

	m.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Error("expected 201")
	}
	if w.Body.String() != "Hello World" {
		t.Error("expected Hello World")
	}
}

func TestRouterGetReaderHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 201

		reader, writer := io.Pipe()
		go func() {
			_, err := writer.Write([]byte("Hello"))
			if err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond)
			_, err = writer.Write([]byte(" World"))
			if err != nil {
				t.Error(err)
			}
			_ = writer.Close()
		}()

		ctx.Body = reader
	})

	m.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Error("expected 201")
	}
	if w.Body.String() != "Hello World" {
		t.Error("expected Hello World")
	}
}

type readCloser struct {
	calledClose *bool
	reader      io.Reader
}

func (r *readCloser) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

func (r *readCloser) Close() error {
	*r.calledClose = true
	return nil
}

func TestRouterGetReadCloserHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	falsePtr := false
	calledClose := &falsePtr

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 201

		reader, writer := io.Pipe()
		go func() {
			_, err := writer.Write([]byte("Hello"))
			if err != nil {
				t.Error(err)
			}
			time.Sleep(100 * time.Millisecond)
			_, err = writer.Write([]byte(" World"))
			if err != nil {
				t.Error(err)
			}
			_ = writer.Close()
		}()

		readCloser := &readCloser{reader: reader, calledClose: calledClose}

		ctx.Body = readCloser
	})

	m.ServeHTTP(w, r)

	if *calledClose != true {
		t.Error("expected Close to be called")
	}

	if w.Code != 201 {
		t.Error("expected 201")
	}
	if w.Body.String() != "Hello World" {
		t.Error("expected Hello World")
	}
}

func TestRouterGetWithMiddlewareAndHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Set("str1", "Hello")
		ctx.Set("str2", "World")
		ctx.Next()
		ctx.Status = 201
	})
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		str1 := ctx.MustGet("str1").(string)
		str2 := ctx.MustGet("str2").(string)
		ctx.Body = str1 + " " + str2
	})

	m.ServeHTTP(w, r)

	if w.Code != 201 {
		t.Errorf("expected 201, got %d", w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("expected Hello World, got %s", w.Body.String())
	}
}

func TestRouterGetWithMiddlewareAndHandlerInline(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Set("str1", "Hello")
		ctx.Set("str2", "World")
		ctx.Next()
	}, func(ctx *navaros.Context) {
		ctx.Status = 200
		str1 := ctx.MustGet("str1").(string)
		str2 := ctx.MustGet("str2").(string)
		ctx.Body = string(str1) + " " + string(str2)
	})

	m.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("expected Hello World, got %s", w.Body.String())
	}
}

func TestRouterGetErroredHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledHandler := false

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Error = errors.New("Hello World")
		ctx.Next()
	})
	m.Get("/a/b/c", func(_ *navaros.Context) {
		calledHandler = true
	})

	m.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if calledHandler {
		t.Error("expected handler not to be called")
	}
}

func TestRouterGetPanickedHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledHandler := false

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		panic("Hello World")
	})
	m.Get("/a/b/c", func(_ *navaros.Context) {
		calledHandler = true
	})

	m.ServeHTTP(w, r)

	if w.Code != 500 {
		t.Errorf("expected 500, got %d", w.Code)
	}
	if calledHandler {
		t.Error("expected handler not to be called")
	}
}

func TestRouterGetSubRouter(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledFirstHandler := false
	calledSecondHandler := false
	calledThirdHandler := false

	m2 := navaros.NewRouter()
	m2.Get("/a/b/c", func(ctx *navaros.Context) {
		calledSecondHandler = true
		ctx.Next()
	})

	m1 := navaros.NewRouter()
	m1.Get("/a/b/c", func(ctx *navaros.Context) {
		calledFirstHandler = true
		ctx.Next()
	})
	m1.Get("/a/b/c", m2)
	m1.Get("/a/b/c", func(ctx *navaros.Context) {
		calledThirdHandler = true
		ctx.Next()
	})

	m1.ServeHTTP(w, r)

	if !calledFirstHandler {
		t.Error("expected first handler to be called")
	}
	if !calledSecondHandler {
		t.Error("expected second handler to be called")
	}
	if !calledThirdHandler {
		t.Error("expected third handler to be called")
	}
}

func TestRouterGetSubSubRouter(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledFirstHandler := false
	calledSecondHandler := false
	calledThirdHandler := false

	m3 := navaros.NewRouter()
	m3.Get("/a/b/c", func(ctx *navaros.Context) {
		calledThirdHandler = true
		ctx.Next()
	})

	m2 := navaros.NewRouter()
	m2.Get("/a/b/c", func(ctx *navaros.Context) {
		calledSecondHandler = true
		ctx.Next()
	})
	m2.Use(m3)

	m1 := navaros.NewRouter()
	m1.Get("/a/b/c", func(ctx *navaros.Context) {
		calledFirstHandler = true
		ctx.Next()
	})
	m1.Use(m2)

	m1.ServeHTTP(w, r)

	if !calledFirstHandler {
		t.Error("expected first handler to be called")
	}
	if !calledSecondHandler {
		t.Error("expected second handler to be called")
	}
	if !calledThirdHandler {
		t.Error("expected third handler to be called")
	}
}

func TestRouterPublicRouteDescriptors(t *testing.T) {
	m := navaros.NewRouter()
	m.PublicGet("/a/b/c", func(ctx *navaros.Context) {})
	m.PublicGet("/a/b/c", func(ctx *navaros.Context) {})
	m.PublicPost("/e/:f/*", func(ctx *navaros.Context) {})

	descriptors := m.RouteDescriptors()

	if len(descriptors) != 2 {
		t.Errorf("expected 2 descriptors, got %d", len(descriptors))
	}
	if descriptors[0].Method != navaros.Get {
		t.Errorf("expected Get method, got %s", descriptors[0].Method)
	}
	if descriptors[0].Pattern.String() != "/a/b/c" {
		t.Errorf("expected /a/b/c pattern, got %s", descriptors[0].Pattern.String())
	}
	if descriptors[1].Method != navaros.Post {
		t.Errorf("expected Post method, got %s", descriptors[1].Method)
	}
	if descriptors[1].Pattern.String() != "/e/:f/*" {
		t.Errorf("expected /e/:f/* pattern, got %s", descriptors[1].Pattern.String())
	}
}

func TestRouterPublicRouteDescriptorsWithSubRouter(t *testing.T) {
	m3 := navaros.NewRouter()
	m3.PublicGet("/a/b/c", func(ctx *navaros.Context) {})
	m3.PublicPost("/a/b/c", func(ctx *navaros.Context) {})

	m2 := navaros.NewRouter()
	m2.PublicGet("/a/b/c", func(ctx *navaros.Context) {})  //x
	m2.PublicPost("/a/b/c", func(ctx *navaros.Context) {}) //x

	m1 := navaros.NewRouter()
	m1.PublicGet("/a/b/c", func(ctx *navaros.Context) {})  //x
	m2.PublicPost("/a/b/c", func(ctx *navaros.Context) {}) //x

	m1.Use(m2)
	m1.Use("/a/b/c", m3)

	descriptors := m1.RouteDescriptors()

	if len(descriptors) != 4 {
		t.Errorf("expected 4 descriptors, got %d", len(descriptors))
	}

	if descriptors[0].Method != navaros.Get {
		t.Errorf("expected Get method, got %s", descriptors[0].Method)
	}
	if descriptors[0].Pattern.String() != "/a/b/c" {
		t.Errorf("expected /a/b/c pattern, got %s", descriptors[0].Pattern.String())
	}
	if descriptors[1].Method != navaros.Post {
		t.Errorf("expected Post method, got %s", descriptors[1].Method)
	}
	if descriptors[1].Pattern.String() != "/a/b/c" {
		t.Errorf("expected /a/b/c pattern, got %s", descriptors[1].Pattern.String())
	}

	if descriptors[2].Method != navaros.Get {
		t.Errorf("expected Get method, got %s", descriptors[2].Method)
	}
	if descriptors[2].Pattern.String() != "/a/b/c/a/b/c" {
		t.Errorf("expected /a/b/c/a/b/c pattern, got %s", descriptors[2].Pattern.String())
	}
	if descriptors[3].Method != navaros.Post {
		t.Errorf("expected Post method, got %s", descriptors[3].Method)
	}
	if descriptors[3].Pattern.String() != "/a/b/c/a/b/c" {
		t.Errorf("expected /a/b/c/a/b/c pattern, got %s", descriptors[3].Pattern.String())
	}
}

func TestSubRouterCorrectlyProceedsToNextHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/b", nil)
	w := httptest.NewRecorder()

	calledWrongHandler := false
	calledHandler := false

	m1 := navaros.NewRouter()
	m1.Get("/a", func(ctx *navaros.Context) {
		calledWrongHandler = true
		ctx.Next()
	})

	m2 := navaros.NewRouter()
	m2.Get("/b", func(ctx *navaros.Context) {
		calledHandler = true
		ctx.Next()
	})

	m3 := navaros.NewRouter()
	m3.Use(m1)
	m3.Use(m2)

	m3.ServeHTTP(w, r)

	if calledWrongHandler {
		t.Error("expected wrong handler not to be called")
	}
	if !calledHandler {
		t.Error("expected handler to be called")
	}
}

func BenchmarkRouter(b *testing.B) {
	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 201
		ctx.Body = "Hello World"
	})

	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ServeHTTP(w, r)
	}
}

func BenchmarkRouterOnHTTPServer(b *testing.B) {
	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 200
		ctx.Body = "Hello World"
	})

	s := httptest.NewServer(m)
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res *http.Response
		var err error
		b.StartTimer()
		res, err = http.Get(s.URL + "/a/b/c")
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
		if res.StatusCode != 200 {
			b.Errorf("expected 200, got %d", res.StatusCode)
		}
	}
}

func BenchmarkGoMux(b *testing.B) {
	m := http.NewServeMux()
	m.HandleFunc("/a/b/c", func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(200)
		_, err := res.Write([]byte("Hello World"))
		if err != nil {
			b.Error(err)
		}
	})

	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ServeHTTP(w, r)
	}
}

func BenchmarkGoMuxOnHTTPServer(b *testing.B) {
	m := http.NewServeMux()
	m.HandleFunc("/a/b/c", func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(200)
		_, err := res.Write([]byte("Hello World"))
		if err != nil {
			b.Error(err)
		}
	})

	s := httptest.NewServer(m)
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res *http.Response
		var err error
		b.StartTimer()
		res, err = http.Get(s.URL + "/a/b/c")
		b.StopTimer()
		if err != nil {
			b.Error(err)
		}
		if res.StatusCode != 200 {
			b.Errorf("expected 200, got %d", res.StatusCode)
		}
	}
}
func TestRouterPost(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Post("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("POST", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterPut(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Put("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRouterPatch(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Patch("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRouterDelete(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Delete("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRouterOptions(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Options("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRouterHead(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.Head("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestRouterAll(t *testing.T) {
	router := navaros.NewRouter()
	called := false
	router.All("/test", func(ctx *navaros.Context) {
		called = true
		ctx.Status = 200
	})

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	for _, method := range methods {
		called = false
		req := httptest.NewRequest(method, "/test", nil)
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)

		if !called {
			t.Errorf("expected handler to be called for %s", method)
		}
	}
}

func TestRouterPublicAll(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicAll("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	descriptors := router.RouteDescriptors()
	if len(descriptors) != 1 {
		t.Errorf("expected 1 descriptor, got %d", len(descriptors))
	}
}

func TestRouterPublicPut(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicPut("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("PUT", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterPublicPatch(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicPatch("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("PATCH", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterPublicDelete(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicDelete("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("DELETE", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterPublicOptions(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicOptions("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterPublicHead(t *testing.T) {
	router := navaros.NewRouter()
	router.PublicHead("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("HEAD", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterLookup(t *testing.T) {
	router := navaros.NewRouter()
	handler := func(ctx *navaros.Context) {}
	router.Get("/users/:id", handler)

	method, pattern, found := router.Lookup(handler)
	if !found {
		t.Error("expected to find handler")
	}
	if method != navaros.Get {
		t.Errorf("expected GET method, got %s", method)
	}
	if pattern.String() != "/users/:id" {
		t.Errorf("expected /users/:id pattern, got %s", pattern.String())
	}

	_, _, found = router.Lookup(func(ctx *navaros.Context) {})
	if found {
		t.Error("expected not to find nonexistent handler")
	}
}

func TestSetPrintHandlerErrors(t *testing.T) {
	navaros.SetPrintHandlerErrors(false)
	navaros.SetPrintHandlerErrors(true)
}

func TestRouterNotFound(t *testing.T) {
	router := navaros.NewRouter()
	router.Get("/exists", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("expected 404 for nonexistent route, got %d", res.Code)
	}
}

func TestRouterMethodNotAllowed(t *testing.T) {
	router := navaros.NewRouter()
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	req := httptest.NewRequest("POST", "/test", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 404 {
		t.Errorf("expected 404 for method not allowed, got %d", res.Code)
	}
}

func TestRouterNestedParams(t *testing.T) {
	router := navaros.NewRouter()
	router.Get("/users/:userId/posts/:postId", func(ctx *navaros.Context) {
		userId := ctx.Params().Get("userId")
		postId := ctx.Params().Get("postId")

		if userId != "123" || postId != "456" {
			t.Errorf("expected userId=123 and postId=456, got userId=%s postId=%s", userId, postId)
		}
		ctx.Status = 200
	})

	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if res.Code != 200 {
		t.Errorf("expected 200, got %d", res.Code)
	}
}

func TestRouterMiddlewareShortCircuit(t *testing.T) {
	router := navaros.NewRouter()
	handlerCalled := false

	router.Use("/admin/*", func(ctx *navaros.Context) {
		ctx.Status = 401
		ctx.Body = "unauthorized"
	})

	router.Get("/admin/users", func(ctx *navaros.Context) {
		handlerCalled = true
		ctx.Status = 200
	})

	req := httptest.NewRequest("GET", "/admin/users", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	if handlerCalled {
		t.Error("expected handler not to be called when middleware short-circuits")
	}
	if res.Code != 401 {
		t.Errorf("expected 401, got %d", res.Code)
	}
}

func TestRouterPanicRecovery(t *testing.T) {
	router := navaros.NewRouter()
	router.Get("/panic", func(ctx *navaros.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != 500 {
		t.Errorf("expected 500 after panic, got %d", res.Code)
	}
}

func Test_Router_SubRouterHandlerNotCallingNextDoesNotPropagateToParent(t *testing.T) {
	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	calledParentHandler := false

	subRouter := navaros.NewRouter()
	subRouter.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
		ctx.Body = "handled by sub-router"
	})

	parentRouter := navaros.NewRouter()
	parentRouter.Use(subRouter)
	parentRouter.Get("/test", func(ctx *navaros.Context) {
		calledParentHandler = true
	})

	parentRouter.ServeHTTP(w, r)

	if calledParentHandler {
		t.Error("expected parent handler not to be called when sub-router handler does not call next")
	}
	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "handled by sub-router" {
		t.Errorf("expected 'handled by sub-router', got %q", w.Body.String())
	}
}

func TestRouterLookupNonexistent(t *testing.T) {
	router := navaros.NewRouter()
	router.Get("/users", func(ctx *navaros.Context) {})

	nonExistentHandler := func(ctx *navaros.Context) {}
	_, _, found := router.Lookup(nonExistentHandler)

	if found {
		t.Error("expected not to find non-existent handler")
	}
}

func TestRouter_Wrap_CallsWrapAroundHandler(t *testing.T) {
	var order []string

	router := navaros.NewRouter()
	router.Wrap(func(ctx *navaros.Context) {
		order = append(order, "wrap-before")
		ctx.Next()
		order = append(order, "wrap-after")
	})
	router.Get("/test", func(ctx *navaros.Context) {
		order = append(order, "handler")
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	if len(order) != 3 || order[0] != "wrap-before" || order[1] != "handler" || order[2] != "wrap-after" {
		t.Errorf("expected [wrap-before handler wrap-after], got %v", order)
	}
}

func TestRouter_Wrap_MultipleWraps(t *testing.T) {
	var order []string

	router := navaros.NewRouter()
	router.Wrap(func(ctx *navaros.Context) {
		order = append(order, "outer-before")
		ctx.Next()
		order = append(order, "outer-after")
	})
	router.Wrap(func(ctx *navaros.Context) {
		order = append(order, "inner-before")
		ctx.Next()
		order = append(order, "inner-after")
	})
	router.Get("/test", func(ctx *navaros.Context) {
		order = append(order, "handler")
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	expected := []string{"outer-before", "inner-before", "handler", "inner-after", "outer-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

func TestRouter_Wrap_WrapsEachHandlerIndividually(t *testing.T) {
	wrapCount := 0

	router := navaros.NewRouter()
	router.Wrap(func(ctx *navaros.Context) {
		wrapCount++
		ctx.Next()
	})
	router.Use(func(ctx *navaros.Context) {
		ctx.Next()
	})
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	if wrapCount != 2 {
		t.Errorf("expected wrap to be called 2 times (middleware + handler), got %d", wrapCount)
	}
}

func TestRouter_Wrap_OnlyAffectsSubsequentHandlers(t *testing.T) {
	wrapCount := 0

	router := navaros.NewRouter()
	router.Use(func(ctx *navaros.Context) {
		ctx.Next()
	})
	router.Wrap(func(ctx *navaros.Context) {
		wrapCount++
		ctx.Next()
	})
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	if wrapCount != 1 {
		t.Errorf("expected wrap to be called 1 time (only handler, not middleware before Wrap), got %d", wrapCount)
	}
}

func TestRouter_Wrap_WrappedHandlerReturnsHandler(t *testing.T) {
	var wrapped any

	router := navaros.NewRouter()
	router.Wrap(func(ctx *navaros.Context) {
		wrapped = ctx.WrappedHandler()
		ctx.Next()
	})
	router.Get("/test", func(ctx *navaros.Context) {
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	if wrapped == nil {
		t.Error("expected WrappedHandler to return the handler being wrapped")
	}
}

func TestRouter_Wrap_WrapsMultipleHandlersOnSameNode(t *testing.T) {
	wrapCount := 0
	var order []string

	router := navaros.NewRouter()
	router.Wrap(func(ctx *navaros.Context) {
		wrapCount++
		order = append(order, "wrap")
		ctx.Next()
	})
	router.Get("/test", func(ctx *navaros.Context) {
		order = append(order, "handler1")
		ctx.Next()
	}, func(ctx *navaros.Context) {
		order = append(order, "handler2")
		ctx.Status = 200
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	if wrapCount != 2 {
		t.Errorf("expected wrap called 2 times, got %d", wrapCount)
	}
	expected := []string{"wrap", "handler1", "wrap", "handler2"}
	if len(order) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

type wrapTestTransformer struct{}

func (wrapTestTransformer) TransformRequest(*navaros.Context)  {}
func (wrapTestTransformer) TransformResponse(*navaros.Context) {}

func TestRouter_Wrap_PanicsOnTransformer(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when passing transformer to Wrap")
		}
	}()

	router := navaros.NewRouter()
	router.Wrap(&wrapTestTransformer{})
}
