package navaros_test

import (
	"errors"
	"net/http/httptest"
	"testing"

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

func TestRouterGetWithMiddlewareAndHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	type myStr1 string
	type myStr2 string

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		navaros.CtxSet(ctx, myStr1("Hello"))
		navaros.CtxSet(ctx, myStr2("World"))
		ctx.Next()
		ctx.Status = 201
	})
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		str1 := navaros.CtxMustGet[myStr1](ctx)
		str2 := navaros.CtxMustGet[myStr2](ctx)
		ctx.Body = string(str1) + " " + string(str2)
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

	type myStr1 string
	type myStr2 string

	m := navaros.NewRouter()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		navaros.CtxSet(ctx, myStr1("Hello"))
		navaros.CtxSet(ctx, myStr2("World"))
		ctx.Next()
	}, func(ctx *navaros.Context) {
		ctx.Status = 200
		str1 := navaros.CtxMustGet[myStr1](ctx)
		str2 := navaros.CtxMustGet[myStr2](ctx)
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
