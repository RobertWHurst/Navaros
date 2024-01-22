package navaros_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 201, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
}

func TestRouterGetMiddlewareAndHandler(t *testing.T) {
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

	assert.Equal(t, 201, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
}

func TestRouterGetMiddlewareAndHandlerInline(t *testing.T) {
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

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
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

	assert.Equal(t, 500, w.Code)
	assert.False(t, calledHandler)
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

	assert.Equal(t, 500, w.Code)
	assert.False(t, calledHandler)
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

	assert.True(t, calledFirstHandler)
	assert.True(t, calledSecondHandler)
	assert.True(t, calledThirdHandler)
}

func TestRouterPublicRouteDescriptors(t *testing.T) {
	m := navaros.NewRouter()
	m.PublicGet("/a/b/c", func(ctx *navaros.Context) {})
	m.PublicGet("/a/b/c", func(ctx *navaros.Context) {})
	m.PublicPost("/e/:f/*", func(ctx *navaros.Context) {})

	descriptors := m.RouteDescriptors()

	assert.Len(t, descriptors, 2)
	assert.Equal(t, navaros.Get, descriptors[0].Method)
	assert.Equal(t, "/a/b/c", descriptors[0].Pattern.String())
	assert.Equal(t, navaros.Post, descriptors[1].Method)
	assert.Equal(t, "/e/:f/*", descriptors[1].Pattern.String())
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

	assert.Len(t, descriptors, 4)

	assert.Equal(t, navaros.Get, descriptors[0].Method)
	assert.Equal(t, "/a/b/c", descriptors[0].Pattern.String())
	assert.Equal(t, navaros.Post, descriptors[1].Method)
	assert.Equal(t, "/a/b/c", descriptors[1].Pattern.String())

	assert.Equal(t, navaros.Get, descriptors[2].Method)
	assert.Equal(t, "/a/b/c/a/b/c", descriptors[2].Pattern.String())
	assert.Equal(t, navaros.Post, descriptors[3].Method)
	assert.Equal(t, "/a/b/c/a/b/c", descriptors[3].Pattern.String())
}
