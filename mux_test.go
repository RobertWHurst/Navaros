package navaros_test

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/robertwhurst/navaros"
	"github.com/stretchr/testify/assert"
)

func TestNewNavaros(t *testing.T) {
	navaros.New()
}

func TestNavarosGetSimpleHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	m := navaros.New()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Status = 201
		ctx.Body = "Hello World"
	})

	m.ServeHTTP(w, r)

	assert.Equal(t, 201, w.Code)
	assert.Equal(t, "Hello World", w.Body.String())
}

func TestNavarosGetMiddlewareAndHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	type myStr1 string
	type myStr2 string

	m := navaros.New()
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

func TestNavarosGetMiddlewareAndHandlerInline(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	type myStr1 string
	type myStr2 string

	m := navaros.New()
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

func TestNavarosGetErroredHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledHandler := false

	m := navaros.New()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		ctx.Error = errors.New("Hello World")
		ctx.Next()
	})
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		calledHandler = true
	})

	m.ServeHTTP(w, r)

	assert.Equal(t, 500, w.Code)
	assert.False(t, calledHandler)
}

func TestNavarosGetPanickedHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledHandler := false

	m := navaros.New()
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		panic("Hello World")
	})
	m.Get("/a/b/c", func(ctx *navaros.Context) {
		calledHandler = true
	})

	m.ServeHTTP(w, r)

	assert.Equal(t, 500, w.Code)
	assert.False(t, calledHandler)
}

func TestNavarosGetSubNavaros(t *testing.T) {
	r := httptest.NewRequest("GET", "/a/b/c", nil)
	w := httptest.NewRecorder()

	calledFirstHandler := false
	calledSecondHandler := false
	calledThirdHandler := false

	m2 := navaros.New()
	m2.Get("/a/b/c", func(ctx *navaros.Context) {
		calledSecondHandler = true
		ctx.Next()
	})

	m1 := navaros.New()
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
