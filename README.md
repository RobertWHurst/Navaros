# Navaros

A lightweight, flexible HTTP router for Go. Build fast web applications with powerful route matching, middleware support, and composable handlers.

[![Go Reference](https://pkg.go.dev/badge/github.com/RobertWHurst/navaros.svg)](https://pkg.go.dev/github.com/RobertWHurst/navaros)
[![Go Report Card](https://goreportcard.com/badge/github.com/RobertWHurst/navaros)](https://goreportcard.com/report/github.com/RobertWHurst/navaros)
[![CI](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml/badge.svg)](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/RobertWHurst/dfe4585fccd1ef915602a113e05d9daf/raw/navaros-coverage.json)](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/RobertWHurst/navaros)](https://github.com/RobertWHurst/Navaros/releases)
[![License](https://img.shields.io/github/license/RobertWHurst/Navaros)](LICENSE)
[![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/RobertWHurst)

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Context](#context)
  - [Middleware](#middleware)
  - [Context Lifecycle](#context-lifecycle)
- [Routing](#routing)
  - [Route Patterns](#route-patterns)
  - [HTTP Methods](#http-methods)
  - [Route Parameters](#route-parameters)
- [Request Handling](#request-handling)
  - [Accessing Request Data](#accessing-request-data)
  - [Request Body](#request-body)
- [Response Handling](#response-handling)
  - [Setting Response Data](#setting-response-data)
  - [Response Body](#response-body)
  - [Redirects](#redirects)
- [Built-in Middleware](#built-in-middleware)
  - [JSON Middleware](#json-middleware)
  - [MessagePack Middleware](#messagepack-middleware)
  - [Protocol Buffers Middleware](#protocol-buffers-middleware)
  - [Set Middleware Variants](#set-middleware-variants)
- [Advanced Usage](#advanced-usage)
  - [Nested Routers](#nested-routers)
  - [Authentication](#authentication)
  - [Error Handling](#error-handling)
  - [Custom Middleware](#custom-middleware)
- [Integration with HTTP Servers](#integration-with-http-servers)
- [Performance](#performance)
- [Architecture](#architecture)
- [Testing](#testing)
- [Help Welcome](#help-welcome)
- [License](#license)
- [Related Projects](#related-projects)

## Features

- 🚀 **High Performance** - Efficient route matching and context pooling for minimal overhead
- 🔌 **Middleware Support** - Composable middleware chain for request/response processing
- 🎯 **Powerful Patterns** - Flexible routing with parameters, wildcards, regex constraints, and modifiers
- 📦 **Body Handling** - Streaming and buffered request/response bodies with unmarshal support
- 🗂️ **Multiple Formats** - Built-in middleware for JSON, MessagePack, and Protocol Buffers
- 🛡️ **Panic Recovery** - Built-in handler panic recovery prevents crashes
- 📋 **Unified Context** - Single context object for request, response, params, and cancellation
- 🔄 **Context Cancellation** - Implements Go's context interface for cancellable operations
- 📁 **Nestable Routers** - Modular route organization with sub-routers
- ⚡ **Minimal Dependencies** - Core router uses only Go standard library
- 🧩 **Extensible** - Simple interfaces for custom middleware and handlers

## Installation

```bash
go get github.com/RobertWHurst/navaros
```

## Quick Start

```go
package main

import (
	"net/http"
	"github.com/RobertWHurst/navaros"
)

func main() {
	router := navaros.NewRouter()

	router.Get("/hello/:name", func(ctx *navaros.Context) {
		name := ctx.Params()["name"]
		ctx.Body = "Hello, " + name + "!"
	})

	http.ListenAndServe(":8080", router)
}
```

## Core Concepts

### Context

The Context is the central object passed to every handler and middleware. It provides unified access to the HTTP request, response, route parameters, and implements Go's context.Context interface for cancellation support.

The context gives handlers everything they need to process a request and build a response. You can access request data like headers and query parameters, set response data like status codes and body content, and store per-request values for passing data between middleware and handlers.

**Storage:** Use `Set` and `Get` to store per-request data. This is useful for passing information between middleware and handlers, such as authenticated user details or request IDs.

```go
router.Get("/user/:id", func(ctx *navaros.Context) {
	ctx.Set("requestID", "abc123")
	ctx.Next()
})

router.Get("/user/:id", func(ctx *navaros.Context) {
	requestID := ctx.Get("requestID").(string)
	userID := ctx.Params()["id"]
	ctx.Body = "User " + userID + " (Request: " + requestID + ")"
})
```

### Middleware

Middleware functions execute before and after handlers in a composable chain. They can inspect and modify the context, perform authentication, log requests, or short-circuit the chain by not calling `Next()`.

Middleware runs in the order it's registered. Each middleware can perform work before calling `Next()` (pre-processing) and after `Next()` returns (post-processing). This allows middleware to wrap handler execution with setup and teardown logic.

You can register middleware globally to run for all routes, or scope it to specific path patterns. Pattern-specific middleware only runs when the request path matches the pattern.

```go
router.Use(func(ctx *navaros.Context) {
	start := time.Now()
	ctx.Next()
	duration := time.Since(start)
	log.Printf("%s %s - %dms", ctx.Method(), ctx.Path(), duration.Milliseconds())
})

router.Use("/api/**", func(ctx *navaros.Context) {
	token := ctx.RequestHeaders().Get("Authorization")
	if token == "" {
		ctx.Status = 401
		ctx.Body = "Unauthorized"
		return
	}
	ctx.Next()
})

router.Get("/api/users", func(ctx *navaros.Context) {
	ctx.Body = "User list"
})
```

### Context Lifecycle

Contexts are pooled and reused for performance. When a handler completes, its context is immediately returned to the pool and may be reused for a different request. This means handlers must block until all operations using the context are complete.

**Important:** Do not keep references to contexts after your handler returns. Do not store contexts in struct fields, pass them to goroutines that outlive the handler, or reference them in callbacks that fire after the handler completes. The context becomes invalid the moment your handler returns.

If you try to use a context after the handler returns, you'll get a clear error message: "context cannot be used after handler returns - handlers must block until all operations complete". This prevents race conditions and use-after-free bugs.

For long-running operations like streaming responses, the handler must block until the stream completes. You can use the context's cancellation support to detect when the client disconnects.

```go
router.Get("/stream", func(ctx *navaros.Context) {
	ctx.Headers.Set("Content-Type", "text/event-stream")
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ctx.Write([]byte("data: ping\n\n"))
			ctx.Flush()
		}
	}
})
```

## Routing

### Route Patterns

Navaros supports fairly powerful route patterns. The following is a list of supported pattern segment types.

- Static - `/a/b/c` - Matches the exact path
- Wildcard - `/a/*/c` - Pattern segments with a single `*` match any path segment
- Dynamic - `/a/:b/c` - Pattern segments prefixed with `:` match any path segment and the value of this segment from the matched path is available via the `Params` method, and will be filled under a key matching the name of the pattern segment, ie: pattern of `/a/:b/c` will match `/a/1/c` and the value of `b` in the params will be `1`

Pattern segments can also be suffixed with additional modifiers.

- `?` - Optional - `/a/:b?/c` - Matches `/a/c` and `/a/1/c`
- `*` - Greedy - `/a/:b*/c` - Matches `/a/c` and `/a/1/2/3/c`
- `+` - One or more - `/a/:b+/c` - Matches `/a/1/c` and `/a/1/2/3/c` but not `/a/c`

You can also provide a regular expression to restrict matches for a pattern segment.

- `/a/:b(\\d+)/c` - Matches `/a/1/c` and `/a/2/c` but not `/a/b/c`

You can escape any of the special characters used by these operators by prefixing them with a `\\`.

- `/a/\\:b/c` - Matches `/a/:b/c`

And all of these can be combined.

- `/a/:b(\\d+)/*?/(d|e)+` - Matches `/a/1/d`, `/a/1/e`, `/a/2/c/d/e/f/g`, and `/a/3/1/d` but not `/a/b/c`, `/a/1`, or `/a/1/c/f`

Register more specific patterns before general ones to ensure correct matching.

```go
router.Get("/users/:id(\\d+)", func(ctx *navaros.Context) {
	ctx.Body = "Numeric user ID: " + ctx.Params()["id"]
})

router.Get("/users/:slug", func(ctx *navaros.Context) {
	ctx.Body = "User slug: " + ctx.Params()["slug"]
})

router.Get("/files/:path+", func(ctx *navaros.Context) {
	ctx.Body = "File path: " + ctx.Params()["path"]
})
```

### HTTP Methods

Routes can be registered for specific HTTP methods using method-specific functions like `Get()`, `Post()`, `Put()`, `Patch()`, `Delete()`, `Options()`, and `Head()`. Each function takes a pattern and one or more handlers to execute when both the pattern and method match.

The `All()` method registers handlers that run for any HTTP method on the given pattern. This is useful for cross-cutting concerns like logging middleware that should run regardless of the request method, or for APIs that handle multiple methods on the same endpoint.

Routes are matched in registration order within each method. If you register both method-specific and `All()` handlers for the same pattern, all matching handlers will run in the order they were registered.

```go
router.All("/api/users", func(ctx *navaros.Context) {
	log.Printf("%s /api/users", ctx.Method())
	ctx.Next()
})

router.Get("/api/users", func(ctx *navaros.Context) {
	ctx.Body = []User{{Name: "Alice"}, {Name: "Bob"}}
})

router.Post("/api/users", func(ctx *navaros.Context) {
	var user User
	ctx.UnmarshalRequestBody(&user)
	ctx.Status = 201
	ctx.Body = user
})
```

### Route Parameters

Parameters are captured from the request path based on the route pattern. They're accessed through the context's `Params()` method, which returns a map-like object.

Parameters are always strings since they come from the URL path. If you need other types, parse the parameter value in your handler.

```go
router.Get("/users/:id", func(ctx *navaros.Context) {
	userID := ctx.Params()["id"]
	
	id, err := strconv.Atoi(userID)
	if err != nil {
		ctx.Status = 400
		ctx.Body = "Invalid user ID"
		return
	}
	
	ctx.Body = fmt.Sprintf("User ID: %d", id)
})
```

## Request Handling

### Accessing Request Data

The context provides direct access to all request data. You can read headers, query parameters, cookies, the request body, and URL components.

**Headers** are accessed as a standard `http.Header` map. You can read individual headers or iterate over all of them.

**Query parameters** come from the URL's query string. They're accessed as a `url.Values` map, which handles multiple values for the same parameter.

**Cookies** can be read by name. The context returns the cookie value or an error if the cookie doesn't exist.

**URL components** like the protocol, host, and path are available directly through the context.

**TLS information** is available if the request came over HTTPS. This includes certificate details and negotiated protocol versions.

```go
router.Get("/info", func(ctx *navaros.Context) {
	headers := ctx.RequestHeaders().Get("User-Agent")
	query := ctx.Query().Get("search")
	cookie, _ := ctx.RequestCookie("session")
	
	if tls := ctx.RequestTLS(); tls != nil {
		ctx.Body = fmt.Sprintf("Secure connection: %s", tls.Version)
	} else {
		ctx.Body = "Insecure connection"
	}
})
```

### Request Body

Request bodies can be accessed as a streaming reader or unmarshalled into Go values. The approach you choose depends on the size of the body and how you need to process it.

**Streaming** with `ctx.RequestBodyReader()` gives you an `io.ReadCloser` for reading the body incrementally. This is ideal for large uploads like file transfers where you don't want to buffer the entire body in memory. You can pipe it directly to a file, database, or other destination. The reader respects `MaxRequestBodySize` limits to prevent memory exhaustion from malicious requests.

**Unmarshalling** with `ctx.UnmarshalRequestBody(&value)` requires middleware to set up the unmarshaller function. The JSON middleware handles this automatically - it buffers the body and decodes it as JSON. This is the simplest approach for structured data like API requests where the entire body fits comfortably in memory.

You can set custom unmarshallers with `ctx.SetRequestBodyUnmarshaller()` for other content types or special parsing requirements. The unmarshaller function receives a pointer to your target value and returns an error if unmarshalling fails. This allows you to support XML, form data, Protocol Buffers, or any other format.

```go
import "github.com/RobertWHurst/navaros/middleware/json"

router.Use(json.Middleware(nil))

router.Post("/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = 400
		ctx.Body = "Invalid request body"
		return
	}
	
	ctx.Status = 201
	ctx.Body = user
})

router.Post("/upload", func(ctx *navaros.Context) {
	reader := ctx.RequestBodyReader()
	defer reader.Close()
	
	file, _ := os.Create("/tmp/upload")
	defer file.Close()
	
	io.Copy(file, reader)
	ctx.Body = "File uploaded"
})
```

## Response Handling

### Setting Response Data

Handlers build responses by setting fields on the context. The status code, headers, cookies, and body are all set directly on the context.

**Status codes** are set as integers. If you don't set a status code, Navaros will infer one based on the response body: 200 if there's a body, 404 if there's no body.

**Headers** are set as a standard `http.Header` map. Headers set on the context are written to the response before the status code.

**Cookies** are added as `http.Cookie` pointers. They're written to the response as Set-Cookie headers.

```go
router.Get("/set-cookie", func(ctx *navaros.Context) {
	ctx.Status = 200
	ctx.Headers.Set("Content-Type", "text/plain")
	ctx.Cookies = append(ctx.Cookies, &http.Cookie{
		Name:  "session",
		Value: "abc123",
		Path:  "/",
	})
	ctx.Body = "Cookie set"
})
```

### Response Body

The response body can be set in several ways depending on your needs. Navaros handles the details of writing the body to the client based on what type of value you provide.

**Simple bodies** like strings and byte slices are written directly to the response with no additional processing. Set `ctx.Body = "Hello World"` or `ctx.Body = []byte{...}` and Navaros writes it as-is. This is the fastest approach for static content or when you've already formatted the response.

**Structured data** can be set as any Go value like `ctx.Body = User{Name: "Alice"}`. Middleware will marshal it to the appropriate format - the JSON middleware marshals values to JSON. This is the easiest approach for APIs since you just set the body to your response struct and let middleware handle encoding.

**Streaming responses** are written by calling `ctx.Write()` in a loop. Each call immediately writes bytes to the client without buffering the entire response. This is essential for large responses like file downloads, server-sent events, or any response that might not fit in memory. Remember that your handler must block until streaming completes - don't start a goroutine that streams after your handler returns.

**io.Reader bodies** like `ctx.Body = file` allow you to set any reader as the response body. Navaros will copy from the reader to the response, closing it if it implements `io.Closer`. This is useful for proxying responses or serving files without loading them entirely into memory.

You can set custom marshallers with `ctx.SetResponseBodyMarshaller()` for other content types or special encoding requirements. The marshaller function receives your body value and returns an `io.Reader` that Navaros will copy to the response.

```go
import "github.com/RobertWHurst/navaros/middleware/json"

router.Use(json.Middleware(nil))

router.Get("/string", func(ctx *navaros.Context) {
	ctx.Body = "Hello World"
})

router.Get("/json", func(ctx *navaros.Context) {
	ctx.Body = User{Name: "Alice", Email: "alice@example.com"}
})

router.Get("/stream", func(ctx *navaros.Context) {
	for i := 0; i < 10; i++ {
		ctx.Write([]byte(fmt.Sprintf("Chunk %d\n", i)))
		ctx.Flush()
		time.Sleep(100 * time.Millisecond)
	}
})

router.Get("/file", func(ctx *navaros.Context) {
	file, _ := os.Open("/path/to/file.pdf")
	ctx.Headers.Set("Content-Type", "application/pdf")
	ctx.Body = file
})
```

### Redirects

Redirects are created using the `Redirect` type. Set it as the response body and Navaros will handle the Location header and status code automatically.

Relative redirects are resolved against the current request URL. Absolute redirects are used as-is.

```go
router.Get("/old-path", func(ctx *navaros.Context) {
	ctx.Body = navaros.Redirect{To: "/new-path"}
})

router.Get("/login", func(ctx *navaros.Context) {
	ctx.Status = 301
	ctx.Body = navaros.Redirect{To: "https://auth.example.com/login"}
})
```

## Built-in Middleware

### JSON Middleware

The JSON middleware automatically marshals and unmarshals JSON request and response bodies. It sets up the context's unmarshal and marshal functions to handle JSON encoding.

For requests, it reads the body and provides an unmarshal function that decodes JSON into Go values. For responses, it marshals any non-reader body value to JSON before writing it.

```go
import "github.com/RobertWHurst/navaros/middleware/json"

router.Use(json.Middleware(nil))

router.Post("/api/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = 400
		ctx.Body = "Invalid JSON"
		return
	}
	
	ctx.Status = 201
	ctx.Body = user
})
```

### MessagePack Middleware

The MessagePack middleware provides binary serialization support using MessagePack format. It automatically handles request unmarshalling and response marshalling for Content-Type `application/msgpack` or `application/x-msgpack`.

MessagePack is more compact and faster than JSON, making it ideal for high-performance APIs or bandwidth-constrained environments.

```go
import "github.com/RobertWHurst/navaros/middleware/msgpack"

router.Use(msgpack.Middleware(nil))

router.Post("/api/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = 400
		ctx.Body = msgpack.Error("Invalid MessagePack")
		return
	}
	
	ctx.Status = 201
	ctx.Body = user
})
```

Like the JSON middleware, MessagePack middleware supports special response types:
- `msgpack.Error` - Returns `{"error": "message"}`
- `msgpack.FieldError` - Returns validation error format
- `msgpack.M` - Shorthand for `map[string]any`

### Protocol Buffers Middleware

The Protocol Buffers middleware provides efficient binary serialization using Protocol Buffers. It handles Content-Type `application/protobuf` or `application/x-protobuf`.

Protocol Buffers require you to define `.proto` schemas and generate Go code with `protoc`. The middleware works with any `proto.Message` implementation.

```go
import (
	"github.com/RobertWHurst/navaros/middleware/protobuf"
	"your-project/api/userpb"
)

router.Use(protobuf.Middleware(nil))

router.Post("/api/users", func(ctx *navaros.Context) {
	var req userpb.CreateUserRequest
	if err := ctx.UnmarshalRequestBody(&req); err != nil {
		ctx.Status = 400
		ctx.Body = "Invalid protobuf"
		return
	}
	
	ctx.Status = 201
	ctx.Body = &userpb.CreateUserResponse{
		Id:   123,
		Name: req.Name,
	}
})
```

The middleware automatically sets Content-Type headers and validates that request/response bodies implement `proto.Message`.

### Set Middleware Variants

The Set middleware family lets you store values on the context as middleware. This is useful for setting up common values that multiple handlers need.

**set** stores static values on every request.

**setfn** calls a function on every request and stores the result. This is useful for values that change per request, like request IDs or timestamps.

**setvalue** dereferences a pointer and stores the value. This is useful when the value might change between requests but you want to capture the current value.

```go
import (
	"github.com/RobertWHurst/navaros/middleware/set"
	"github.com/RobertWHurst/navaros/middleware/setfn"
	"github.com/RobertWHurst/navaros/middleware/setvalue"
)

router.Use(set.Middleware("version", "1.0.0"))

router.Use(setfn.Middleware("requestID", func() string {
	return uuid.New().String()
}))

config := &Config{MaxItems: 100}
router.Use(setvalue.Middleware("maxItems", &config.MaxItems))

router.Get("/info", func(ctx *navaros.Context) {
	version := ctx.Get("version").(string)
	requestID := ctx.Get("requestID").(string)
	maxItems := ctx.Get("maxItems").(int)
	
	ctx.Body = fmt.Sprintf("v%s (request: %s, max: %d)", version, requestID, maxItems)
})
```

## Advanced Usage

### Nested Routers

Routers can be nested to organize routes modularly. Create separate routers for different parts of your application, then mount them on the main router.

Sub-routers inherit middleware from their parent, and you can add sub-router-specific middleware. This allows you to build up middleware stacks that apply to specific sections of your application.

```go
apiRouter := navaros.NewRouter()
apiRouter.Use(authMiddleware)

apiRouter.Get("/users", func(ctx *navaros.Context) {
	ctx.Body = []User{{Name: "Alice"}}
})

apiRouter.Get("/posts", func(ctx *navaros.Context) {
	ctx.Body = []Post{{Title: "Hello"}}
})

mainRouter := navaros.NewRouter()
mainRouter.Use(loggingMiddleware)
mainRouter.Use("/api/**", apiRouter)

mainRouter.Get("/", func(ctx *navaros.Context) {
	ctx.Body = "Welcome"
})
```

### Authentication

Authentication is typically implemented as middleware. The middleware runs before handlers, checks credentials, and either continues the chain or returns an error response.

Pattern-specific middleware lets you protect specific routes or route groups. Store authenticated user details on the context so handlers can access them.

```go
func authMiddleware(ctx *navaros.Context) {
	token := ctx.RequestHeaders().Get("Authorization")
	if token == "" {
		ctx.Status = 401
		ctx.Body = "Unauthorized"
		return
	}
	
	user, err := validateToken(token)
	if err != nil {
		ctx.Status = 403
		ctx.Body = "Forbidden"
		return
	}
	
	ctx.Set("user", user)
	ctx.Next()
}

router.Use("/api/**", authMiddleware)

router.Get("/api/profile", func(ctx *navaros.Context) {
	user := ctx.Get("user").(*User)
	ctx.Body = user
})
```

### Error Handling

Navaros automatically recovers from panics in handlers. When a panic occurs, the context's Error and ErrorStack fields are set, and a 500 status code is returned.

You can implement custom error handling middleware that runs after handlers, checks the Error field, and returns appropriate error responses. This lets you control error formatting and logging.

```go
func errorHandler(ctx *navaros.Context) {
	ctx.Next()
	
	if ctx.Error != nil {
		log.Printf("Handler error: %v\n%s", ctx.Error, ctx.ErrorStack)
		ctx.Status = 500
		ctx.Body = map[string]string{
			"error": "Internal server error",
		}
	}
}

router.Use(errorHandler)

router.Get("/panic", func(ctx *navaros.Context) {
	panic("something went wrong")
})
```

### Custom Middleware

Middleware is any function that takes a context pointer. It can perform work before calling `Next()`, after calling `Next()`, or both.

To short-circuit the chain, don't call `Next()`. This is useful for middleware that handles certain requests completely, like authentication middleware that returns 401 responses.

Middleware registered earlier runs first. Global middleware runs before pattern-specific middleware.

```go
func corsMiddleware(ctx *navaros.Context) {
	ctx.Headers.Set("Access-Control-Allow-Origin", "*")
	ctx.Next()
}

func rateLimitMiddleware(ctx *navaros.Context) {
	if !checkRateLimit(ctx.RequestRemoteAddress()) {
		ctx.Status = 429
		ctx.Body = "Too many requests"
		return
	}
	ctx.Next()
}

router.Use(corsMiddleware)
router.Use(rateLimitMiddleware)

router.Get("/api/data", func(ctx *navaros.Context) {
	ctx.Body = "Data"
})
```

## Integration with HTTP Servers

Navaros implements the standard `http.Handler` interface, so it works seamlessly with any Go HTTP server or framework.

For the standard `net/http` server, pass the router to `http.ListenAndServe` or register it with `http.Handle`.

Navaros also works with third-party frameworks. Most frameworks provide a way to wrap an `http.Handler`, letting you use Navaros as the routing layer.

```go
router := navaros.NewRouter()

router.Get("/hello", func(ctx *navaros.Context) {
	ctx.Body = "Hello World"
})

http.ListenAndServe(":8080", router)
```

## Performance

Navaros is designed for high performance with minimal overhead.

**Context pooling** reuses context objects to reduce allocations. Contexts are reset and returned to a pool after each request.

**Efficient route matching** uses sequential pattern matching. While not as fast as radix trees for many routes, it's simple, predictable, and fast enough for most applications.

**Zero allocations** in hot paths mean Navaros doesn't create garbage during request handling, reducing GC pressure.

## Architecture

Navaros uses a simple, predictable architecture based on sequential pattern matching and middleware chains.

Routes are matched in the order they're registered. This makes routing behavior predictable and easy to reason about, though you should register more specific patterns before general ones.

The middleware chain is built at registration time and executed sequentially. Each middleware and handler gets the context, performs its work, and optionally calls the next function in the chain.

Context pooling provides performance without complexity. Contexts are reset and reused, keeping allocations low while maintaining simplicity.

## Testing

Test your handlers by creating contexts with test request and response recorders. The context constructor takes any `http.ResponseWriter` and `*http.Request`, so `httptest` types work perfectly.

Run the full test suite with coverage reporting.

```go
func TestHandler(t *testing.T) {
	router := navaros.NewRouter()
	
	router.Get("/hello", func(ctx *navaros.Context) {
		ctx.Body = "Hello World"
	})
	
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()
	
	router.ServeHTTP(rec, req)
	
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", rec.Body.String())
	}
}
```

Run tests:

```bash
go test ./...
go test -cover ./...
```

## Help Welcome

Navaros is open source and welcomes contributions. If you find a bug or have a feature request, please open an issue on GitHub.

Financial support through sponsorship is appreciated and helps maintain the project.

[![Sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/RobertWHurst)

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

[Velaros](https://github.com/RobertWHurst/Velaros) - WebSocket router built as a companion to Navaros. Uses the same routing patterns and middleware concepts for WebSocket message handling.
