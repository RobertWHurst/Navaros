# Navaros

A lightweight, flexible HTTP router for Go. Build fast web applications with powerful route matching, middleware support, and composable handlers.

[![Go Reference](https://pkg.go.dev/badge/github.com/RobertWHurst/navaros.svg)](https://pkg.go.dev/github.com/RobertWHurst/navaros)
[![Go Report Card](https://goreportcard.com/badge/github.com/RobertWHurst/navaros)](https://goreportcard.com/report/github.com/RobertWHurst/navaros)
[![CI](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml/badge.svg)](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/RobertWHurst/dfe4585fccd1ef915602a113e05d9daf/raw/navaros-coverage.json)](https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml)
[![GitHub release](https://img.shields.io/github/v/release/RobertWHurst/Navaros)](https://github.com/RobertWHurst/Navaros/releases)
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
- [Microservices](#microservices)
  - [Public vs Private Routes](#public-vs-private-routes)
  - [Gateway Pattern](#gateway-pattern)
  - [Creating Services](#creating-services)
  - [Service-to-Service Communication](#service-to-service-communication)
  - [Client as Handler](#client-as-handler)
  - [Benefits](#benefits)
- [WebSockets](#websockets)
  - [Mounting WebSocket Routes](#mounting-websocket-routes)
  - [Shared Patterns and Concepts](#shared-patterns-and-concepts)
  - [Real-Time Applications](#real-time-applications)
  - [WebSocket Message Routing](#websocket-message-routing)
  - [Benefits of Using Both](#benefits-of-using-both)
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
- 📦 **Body Handling** - Streaming request/response bodies, with buffering and marshaling via middleware
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
		name := ctx.Params().Get("name")
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
router.Use(func(ctx *navaros.Context) {
	ctx.Set("requestID", generateID())
	ctx.Next()
})

router.Get("/user/:id", func(ctx *navaros.Context) {
	requestID := ctx.Get("requestID").(string)
	userID := ctx.Params().Get("id")
	ctx.Body = "User " + userID + " (Request: " + requestID + ")"
})
```

### Middleware

Middleware functions execute before and after handlers in a composable chain. They can inspect and modify the context, perform authentication, log requests, or short-circuit the chain by not calling `Next()`.

Middleware runs in the order it's registered. Each middleware can perform work before calling `Next()` (pre-processing) and after `Next()` returns (post-processing). This allows middleware to wrap handler execution with setup and teardown logic.

You can register middleware globally to run for all routes, or scope it to specific path patterns. When you provide a path to `Use()`, it automatically matches that path and all sub-paths - no need to add `/**` explicitly.

```go
router.Use(func(ctx *navaros.Context) {
	start := time.Now()
	ctx.Next()
	duration := time.Since(start)
	log.Printf("%s %s - %dms", ctx.Method(), ctx.Path(), duration.Milliseconds())
})

router.Use("/api", func(ctx *navaros.Context) {
	token := ctx.RequestHeaders().Get("Authorization")
	if token == "" {
		ctx.Status = http.StatusUnauthorized
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

**Important:** Context objects are pooled and reused for performance. When a handler returns, its context is immediately returned to the pool and may be reused for a different request. This means **handlers must block until all operations using the context are complete**.

If you spawn a goroutine or set up a callback that references the context after the handler returns, those operations will fail with an error: `"context cannot be used after handler returns - handlers must block until all operations complete"`.

**Wrong - Don't do this:**

```go
// ❌ This will fail - goroutine uses context after handler returns
router.Get("/async", func(ctx *navaros.Context) {
	go func() {
		time.Sleep(time.Second)
		ctx.Body = "Done" // ERROR: context already returned to pool
	}()
	// Handler returns immediately - context is freed
})
```

**Right - Handler blocks:**

```go
// ✓ Correct - handler blocks until operation completes
router.Get("/stream", func(ctx *navaros.Context) {
	ctx.Headers.Set("Content-Type", "text/event-stream")
	
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	// Handler blocks here, keeping context alive
	for {
		select {
		case <-ctx.Done():
			return // Client disconnected, clean exit
		case <-ticker.C:
			ctx.Write([]byte("data: ping\n\n"))
			ctx.Flush()
		}
	}
})
```

Handlers only need to block as long as they need to use the context. For typical request/response handlers, this means they return immediately after setting the response. For streaming responses, they block until the stream completes. The key rule: don't return from the handler while operations that use the context are still pending.

## Routing

### Route Patterns

Navaros supports fairly powerful route patterns. The following is a list of supported pattern segment types.

- Static - `/a/b/c` - Matches the exact path
- Wildcard - `/a/*/c` - Pattern segments with a single `*` match any path segment
- Dynamic - `/a/:b/c` - Pattern segments prefixed with `:` capture values. `/users/:id` matches `/users/123`, and `ctx.Params().Get("id")` returns `"123"`

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
	ctx.Body = "Numeric user ID: " + ctx.Params().Get("id")
})

router.Get("/users/:slug", func(ctx *navaros.Context) {
	ctx.Body = "User slug: " + ctx.Params().Get("slug")
})

router.Get("/files/:path+", func(ctx *navaros.Context) {
	ctx.Body = "File path: " + ctx.Params().Get("path")
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
	ctx.Status = http.StatusCreated
	ctx.Body = user
})
```

### Route Parameters

Parameters are captured from the request path based on the route pattern. They're accessed through the context's `Params()` method, which returns a map-like object.

Parameters are always strings since they come from the URL path. If you need other types, parse the parameter value in your handler.

```go
router.Get("/users/:id", func(ctx *navaros.Context) {
	userID := ctx.Params().Get("id")
	
	id, err := strconv.Atoi(userID)
	if err != nil {
		ctx.Status = http.StatusBadRequest
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

**Cookies** can be read by name. The context returns an `*http.Cookie` and an error if the cookie doesn't exist.

**URL components** like the protocol, host, and path are available directly through the context.

**TLS information** is available if the request came over HTTPS. This includes certificate details and negotiated protocol versions.

```go
router.Get("/info", func(ctx *navaros.Context) {
	userAgent := ctx.RequestHeaders().Get("User-Agent")
	search := ctx.Query().Get("search")
	
	sessionCookie, err := ctx.RequestCookie("session")
	if err == nil {
		log.Printf("Session: %s", sessionCookie.Value)
	}
	
	if tls := ctx.RequestTLS(); tls != nil {
		ctx.Body = fmt.Sprintf("Secure connection from %s searching for %s", userAgent, search)
	} else {
		ctx.Body = "Insecure connection"
	}
})
```

### Request Body

Most APIs use `ctx.UnmarshalRequestBody(&value)` with middleware like JSON, MessagePack, or Protocol Buffers. The middleware reads and decodes the body for you.

For large uploads like files, use `ctx.RequestBodyReader()` to stream the body without loading it all into memory. The reader respects `MaxRequestBodySize` limits (default 10MB) to prevent memory exhaustion. You can change the limit globally with `navaros.MaxRequestBodySize` or per-request with `ctx.MaxRequestBodySize`. Set to `-1` to disable the limit.

You can set custom unmarshallers with `ctx.SetRequestBodyUnmarshaller()` for other content types.

```go
import "github.com/RobertWHurst/navaros/middleware/json"

router.Use(json.Middleware(nil))

router.Post("/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = http.StatusBadRequest
		ctx.Body = "Invalid request body"
		return
	}
	
	ctx.Status = http.StatusCreated
	ctx.Body = user
})

// Allow larger uploads for this route
router.Post("/upload", func(ctx *navaros.Context) {
	ctx.MaxRequestBodySize = 100 * 1024 * 1024 // 100MB
	
	reader := ctx.RequestBodyReader()
	defer reader.Close()
	
	file, _ := os.Create("/tmp/upload")
	defer file.Close()
	
	io.Copy(file, reader)
	ctx.Status = http.StatusOK
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
	ctx.Status = http.StatusOK
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

**Structured data** can be set as any Go value like `ctx.Body = User{Name: "Alice"}`. Middleware will marshal it to the appropriate format - the JSON middleware marshals values to JSON by setting a marshaller with `ctx.SetResponseBodyMarshaller()`. This is the easiest approach for APIs since you just set the body to your response struct and let middleware handle encoding.

**Streaming responses** use `ctx.Write()` to send bytes directly to the client without buffering. The context implements `io.WriteCloser`, making it compatible with standard library functions like `io.Copy(ctx, reader)` for streaming from any source. This is essential for large responses like file downloads or server-sent events that don't fit in memory. Your handler must block until streaming completes.

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
	ctx.Status = http.StatusMovedPermanently
	ctx.Body = navaros.Redirect{To: "https://auth.example.com/login"}
})
```

## Built-in Middleware

### JSON Middleware

The JSON middleware automatically marshals and unmarshals JSON request and response bodies. It sets up the context's unmarshal and marshal functions to handle JSON encoding.

For requests with `Content-Type: application/json`, it reads the body and provides an unmarshal function that decodes JSON into Go values. For responses, it marshals any non-reader body value to JSON before writing it.

Pass `nil` for default configuration, or use `&json.Options{}` to customize:
- `DisableRequestBodyUnmarshaller` - Skip setting up request unmarshalling
- `DisableResponseBodyMarshaller` - Skip setting up response marshalling

```go
import "github.com/RobertWHurst/navaros/middleware/json"

// Default configuration
router.Use(json.Middleware(nil))

// Custom configuration - response marshalling only
router.Use(json.Middleware(&json.Options{
	DisableRequestBodyUnmarshaller: true,
}))

router.Post("/api/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = http.StatusBadRequest
		ctx.Body = "Invalid JSON"
		return
	}
	
	ctx.Status = http.StatusCreated
	ctx.Body = user
})
```

### MessagePack Middleware

The MessagePack middleware provides binary serialization support using MessagePack format. It automatically handles request unmarshalling and response marshalling for `Content-Type: application/msgpack`.

MessagePack is more compact and faster than JSON, making it ideal for high-performance APIs or bandwidth-constrained environments.

Pass `nil` for default configuration, or use `&msgpack.Options{}` to customize:
- `DisableRequestBodyUnmarshaller` - Skip setting up request unmarshalling
- `DisableResponseBodyMarshaller` - Skip setting up response marshalling

```go
import "github.com/RobertWHurst/navaros/middleware/msgpack"

router.Use(msgpack.Middleware(nil))

router.Post("/api/users", func(ctx *navaros.Context) {
	var user User
	if err := ctx.UnmarshalRequestBody(&user); err != nil {
		ctx.Status = http.StatusBadRequest
		ctx.Body = msgpack.Error("Invalid MessagePack")
		return
	}
	
	ctx.Status = http.StatusCreated
	ctx.Body = user
})
```

Like the JSON middleware, MessagePack middleware supports special response types:
- `msgpack.Error` - Returns `{"error": "message"}`
- `msgpack.FieldError` - Returns validation error format
- `msgpack.M` - Shorthand for `map[string]any`

### Protocol Buffers Middleware

The Protocol Buffers middleware provides efficient binary serialization using Protocol Buffers. It handles `Content-Type: application/protobuf`.

Protocol Buffers require you to define `.proto` schemas and generate Go code with `protoc`. The middleware works with any `proto.Message` implementation.

Pass `nil` for default configuration, or use `&protobuf.Options{}` to customize:
- `DisableRequestBodyUnmarshaller` - Skip setting up request unmarshalling
- `DisableResponseBodyMarshaller` - Skip setting up response marshalling

```go
import (
	"github.com/RobertWHurst/navaros/middleware/protobuf"
	"your-project/api/userpb"
)

router.Use(protobuf.Middleware(nil))

router.Post("/api/users", func(ctx *navaros.Context) {
	var req userpb.CreateUserRequest
	if err := ctx.UnmarshalRequestBody(&req); err != nil {
		ctx.Status = http.StatusBadRequest
		ctx.Body = "Invalid protobuf"
		return
	}
	
	ctx.Status = http.StatusCreated
	ctx.Body = &userpb.CreateUserResponse{
		Id:   123,
		Name: req.Name,
	}
})
```

The middleware automatically sets Content-Type headers and validates that request/response bodies implement `proto.Message`.

### Set Middleware Variants

The Set middleware family lets you store values on the context as middleware. This is useful for setting up common values that multiple handlers need. Each variant takes a key and value/function as parameters.

**set** stores static values on every request - takes a key and value.

**setfn** calls a function on every request and stores the result - takes a key and function. This is useful for values that change per request, like request IDs or timestamps.

**setvalue** dereferences a pointer and stores the value - takes a key and pointer. This is useful when the value might change between requests but you want to capture the current value.

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

maxItems := 100
router.Use(setvalue.Middleware("maxItems", &maxItems))

// Later, maxItems can be changed and setvalue captures the current value per request
maxItems = 200

router.Get("/info", func(ctx *navaros.Context) {
	version := ctx.Get("version").(string)
	requestID := ctx.Get("requestID").(string)
	maxItems := ctx.Get("maxItems").(int) // Gets the dereferenced int value
	
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
mainRouter.Use("/api", apiRouter)

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
		ctx.Status = http.StatusUnauthorized
		ctx.Body = "Unauthorized"
		return
	}
	
	user, err := validateToken(token)
	if err != nil {
		ctx.Status = http.StatusForbidden
		ctx.Body = "Forbidden"
		return
	}
	
	ctx.Set("user", user)
	ctx.Next()
}

router.Use("/api", authMiddleware)

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
		ctx.Status = http.StatusInternalServerError
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
		ctx.Status = http.StatusTooManyRequests
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

## Microservices

Navaros works with [Zephyr](https://github.com/telemetrytv/Zephyr), a microservice framework that routes HTTP requests over message transports like NATS. This lets you write services as regular HTTP handlers while getting service discovery and routing.

**Note:** The examples below require a NATS server. See [NATS documentation](https://nats.io/) for installation.

### Public vs Private Routes

Navaros distinguishes between public and private routes using `Public*()` methods:

- **Public routes** (`PublicGet`, `PublicPost`, etc.) - Accessible through a gateway from external clients
- **Private routes** (`Get`, `Post`, etc.) - Only accessible via service-to-service communication

```go
router := navaros.NewRouter()

// Public route - accessible via gateway
router.PublicGet("/api/users", func(ctx *navaros.Context) {
    ctx.Body = []User{{Name: "Alice"}}
})

// Private route - only accessible to other services
router.Get("/internal/stats", func(ctx *navaros.Context) {
    ctx.Body = getInternalStats()
})
```

### Gateway Pattern

A Zephyr gateway sits at the edge of your service network, routing external HTTP requests to the appropriate services based on their registered routes.

```go
// Gateway service
conn, _ := nats.Connect("nats://localhost:4222")
gateway := zephyr.NewGateway("api-gateway", natstransport.New(conn))
gateway.Start()

http.ListenAndServe(":8080", gateway)
```

### Creating Services

Services register themselves with the gateway and handle requests. When using Navaros, public routes are automatically discovered and registered.

```go
// User service
router := navaros.NewRouter()
router.Use(json.Middleware(nil))

router.PublicGet("/users", func(ctx *navaros.Context) {
    ctx.Body = []User{{ID: 1, Name: "Alice"}}
})

router.PublicPost("/users", func(ctx *navaros.Context) {
    var user User
    ctx.UnmarshalRequestBody(&user)
    ctx.Status = http.StatusCreated
    ctx.Body = user
})

conn, _ := nats.Connect("nats://localhost:4222")
service := zephyr.NewService("user-service", natstransport.New(conn), router)
service.Start()
```

### Service-to-Service Communication

Services can call each other directly using Zephyr's client, which can access both public and private routes.

```go
// Order service calling user service
conn, _ := nats.Connect("nats://localhost:4222")
client := zephyr.NewClient(natstransport.New(conn))

router := navaros.NewRouter()

router.PublicPost("/orders", func(ctx *navaros.Context) {
    // Call user service to verify user exists
    resp, _ := client.Service("user-service").Get("/users/123")
    if resp.StatusCode == http.StatusNotFound {
        ctx.Status = http.StatusBadRequest
        ctx.Body = "User not found"
        return
    }
    
    // Create order...
    ctx.Status = http.StatusCreated
})

service := zephyr.NewService("order-service", natstransport.New(conn), router)
service.Start()
```

### Client as Handler

Zephyr clients implement `navaros.Handler`, allowing you to proxy requests from one service to another directly in your routing:

```go
client := zephyr.NewClient(natstransport.New(conn))
router := navaros.NewRouter()

// Proxy all /users requests to user-service
router.PublicGet("/users/**", client.Service("user-service"))

// This service now acts as a proxy/facade
service := zephyr.NewService("api-facade", natstransport.New(conn), router)
service.Start()
```

### Benefits

- **Write services like HTTP servers** - No special message handling code
- **Automatic service discovery** - Services find each other through the transport
- **Public/private routes** - Control what's exposed externally vs internally
- **Standard HTTP** - Use all standard HTTP features (methods, headers, status codes)
- **Transport agnostic** - Works with NATS, or implement custom transports

For complete documentation, see [Zephyr on GitHub](https://github.com/telemetrytv/Zephyr).

## WebSockets

Navaros can be combined with [Velaros](https://github.com/RobertWHurst/Velaros), a WebSocket router that brings the same routing and middleware patterns to WebSocket connections. This lets you serve both HTTP and WebSocket traffic from the same server with a consistent API.

**Note:** Velaros cannot be used behind Zephyr services. WebSocket connections require persistent TCP connections that Zephyr's message-based architecture doesn't support. If you need WebSockets in a microservices architecture, either:
- Use Velaros at a socket gateway that sits in front of your Zephyr services
- Wait for the release of Eurus, the WebSocket equivalent of Zephyr that will provide service discovery and routing for WebSocket connections

### Mounting WebSocket Routes

Velaros routers provide a `Middleware()` method that returns a Navaros handler, allowing you to mount WebSocket routes on any path:

```go
import (
	"github.com/RobertWHurst/navaros"
	"github.com/RobertWHurst/navaros/middleware/json"
	"github.com/RobertWHurst/velaros"
	vjson "github.com/RobertWHurst/velaros/middleware/json"
)

// Create HTTP router
httpRouter := navaros.NewRouter()
httpRouter.Use(json.Middleware(nil))

// Create WebSocket router
wsRouter := velaros.NewRouter()
wsRouter.Use(vjson.Middleware())

// Add HTTP routes
httpRouter.Get("/api/users", func(ctx *navaros.Context) {
	ctx.Body = []User{{Name: "Alice"}}
})

// Add WebSocket routes
wsRouter.Bind("/chat/message", func(ctx *velaros.Context) {
	var msg ChatMessage
	ctx.Unmarshal(&msg)
	ctx.Reply(ChatResponse{Status: "received"})
})

// Mount WebSocket router using Middleware() method
httpRouter.Use("/ws", wsRouter.Middleware())

http.ListenAndServe(":8080", httpRouter)
```

### Shared Patterns and Concepts

Both routers use identical pattern syntax and middleware concepts:

**Routing Patterns:**

```go
// HTTP routing
httpRouter.Get("/users/:id", getUserHandler)
httpRouter.Post("/files/**", uploadHandler)

// WebSocket routing - same pattern syntax
wsRouter.Bind("/users/:id", getUserMessageHandler)
wsRouter.Bind("/files/**", fileMessageHandler)
```

**Middleware:**

```go
// HTTP middleware
httpRouter.Use("/admin", func(ctx *navaros.Context) {
	if !authenticated(ctx) {
		ctx.Status = http.StatusUnauthorized
		return
	}
	ctx.Next()
})

// WebSocket middleware - same structure
wsRouter.Use("/admin", func(ctx *velaros.Context) {
	if !authenticated(ctx) {
		ctx.Send(ErrorResponse{Error: "unauthorized"})
		return
	}
	ctx.Next()
})
```

**Context Storage:**

Both routers provide context storage with similar semantics:

```go
// HTTP: per-request storage (cleared after request completes)
httpRouter.Use(func(ctx *navaros.Context) {
	ctx.Set("requestID", generateID())
	ctx.Next()
})

// WebSocket: per-message storage (cleared after message processing completes)
wsRouter.Use(func(ctx *velaros.Context) {
	ctx.Set("messageID", generateID())
	ctx.Next()
})

// WebSocket: per-connection storage (persists for the connection lifetime)
wsRouter.UseOpen(func(ctx *velaros.Context) {
	ctx.SetOnSocket("sessionID", generateID())
})
```

### Real-Time Applications

Velaros enables real-time features while Navaros handles REST APIs:

```go
type ChatServer struct {
	httpRouter *navaros.Router
	wsRouter   *velaros.Router
	broadcast  chan ChatMessage
	clients    sync.Map
}

func NewChatServer() *ChatServer {
	s := &ChatServer{
		httpRouter: navaros.NewRouter(),
		wsRouter:   velaros.NewRouter(),
		broadcast:  make(chan ChatMessage, 100),
	}

	// Start broadcast handler
	go s.handleBroadcasts()

	// HTTP API for message history
	s.httpRouter.Get("/api/messages", func(ctx *navaros.Context) {
		ctx.Body = getMessageHistory()
	})

	// WebSocket connection setup
	s.wsRouter.UseOpen(func(ctx *velaros.Context) {
		socketID := ctx.SocketID()
		msgChan := make(chan ChatMessage, 10)
		s.clients.Store(socketID, msgChan)
	})

	s.wsRouter.UseClose(func(ctx *velaros.Context) {
		socketID := ctx.SocketID()
		if ch, ok := s.clients.LoadAndDelete(socketID); ok {
			close(ch.(chan ChatMessage))
		}
	})

	// Listen for broadcast messages and send to client
	s.wsRouter.Bind("/chat/listen", func(ctx *velaros.Context) {
		socketID := ctx.SocketID()
		msgChan, ok := s.clients.Load(socketID)
		if !ok {
			return
		}

		// Handler blocks - continuously send messages to client
		for {
			select {
			case msg := <-msgChan.(chan ChatMessage):
				if err := ctx.Send(msg); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	})

	// Receive message from client and broadcast
	s.wsRouter.Bind("/chat/send", func(ctx *velaros.Context) {
		var msg ChatMessage
		ctx.Unmarshal(&msg)

		// Send to broadcast channel
		s.broadcast <- msg

		ctx.Reply(ChatResponse{Status: "sent"})
	})

	s.httpRouter.Use("/ws", s.wsRouter.Middleware())

	return s
}

func (s *ChatServer) handleBroadcasts() {
	for msg := range s.broadcast {
		// Send to all connected clients
		s.clients.Range(func(key, value any) bool {
			if msgChan, ok := value.(chan ChatMessage); ok {
				select {
				case msgChan <- msg:
				default:
					// Channel full, skip this client
				}
			}
			return true
		})
	}
}

func (s *ChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.httpRouter.ServeHTTP(w, r)
}
```

### WebSocket Message Routing

While HTTP uses request methods and paths, WebSocket uses message paths for routing. Messages contain a path field that determines which handler processes them:

```javascript
// Client connects via HTTP upgrade
const ws = new WebSocket('ws://localhost:8080/ws');

// Once connected, send messages with paths
ws.send(JSON.stringify({
	path: '/chat/send',       // Routes to handler
	id: 'msg-123',            // For request/reply
	data: {                   // Message payload
		text: 'Hello!'
	}
}));

// Receive responses
ws.onmessage = (event) => {
	const msg = JSON.parse(event.data);
	console.log('Reply to', msg.id, ':', msg.data);
};
```

### Benefits of Using Both

- **Consistent API** - Same routing patterns, middleware structure, and context methods
- **Unified Server** - Serve HTTP and WebSocket from one server on one port
- **Complementary Features** - REST APIs for CRUD, WebSockets for real-time updates
- **Type-Safe Communication** - Both support JSON, MessagePack, and Protocol Buffers

For complete Velaros documentation, see [Velaros on GitHub](https://github.com/RobertWHurst/Velaros).

## Performance

Navaros is designed for high performance with minimal overhead.

**Context pooling** reuses context objects to reduce allocations. Contexts are reset and returned to a pool after each request.

**Pre-built handler chains** are constructed at registration time, not per-request. Each route's middleware and handler sequence is a linked list ready to execute.

**Zero allocations** in hot paths mean Navaros doesn't create garbage during request handling, reducing GC pressure.

**Minimal overhead** from simple, direct code paths. No reflection in request handling, no hidden costs.

## Architecture

Navaros uses a simple, predictable architecture based on sequential pattern matching and middleware chains.

Routes are matched in the order they're registered. This makes routing behavior predictable and easy to reason about, though you should register more specific patterns before general ones.

The middleware chain is built at registration time and executed sequentially. Each middleware and handler gets the context, performs its work, and optionally calls the next function in the chain.

Context pooling provides performance without complexity. Contexts are reset and reused, keeping allocations low while maintaining simplicity.

## Testing

Test your handlers using `httptest` from the standard library. Create a test request, record the response, and assert on the results.

```go
import (
	"net/http"
	"net/http/httptest"
	"testing"
	
	"github.com/RobertWHurst/navaros"
)

func TestHandler(t *testing.T) {
	router := navaros.NewRouter()
	
	router.Get("/hello", func(ctx *navaros.Context) {
		ctx.Body = "Hello World"
	})
	
	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()
	
	router.ServeHTTP(rec, req)
	
	if rec.Code != http.StatusOK {
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

If you want to support this project by throwing me some coffee money, it's greatly appreciated.

[![sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/RobertWHurst)

If you're interested in providing feedback or would like to contribute, please feel free to do so. I recommend first opening an issue expressing your feedback or intent to contribute a change, from there we can consider your feedback or guide your contribution efforts. Any and all help is greatly appreciated since this is an open source effort after all.

Thank you!

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

[Velaros](https://github.com/RobertWHurst/Velaros) - WebSocket router built as a companion to Navaros. Uses the same routing patterns and middleware concepts for WebSocket message handling.

[Zephyr](https://github.com/telemetrytv/Zephyr) - Microservice framework that brings HTTP directly to your services. Works seamlessly with Navaros for service-to-service communication and gateway routing.
