# Navaros

<p>
  <img src="https://github.com/user-attachments/assets/6a2ef6ff-edb9-4ed9-ad8c-9fff3ec71a41" width="400" />
</p>
<p>
  <a href="https://pkg.go.dev/github.com/RobertWHurst/navaros">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg">
  </a>
  <a href="https://goreportcard.com/report/github.com/RobertWHurst/navaros">
    <img src="https://goreportcard.com/badge/github.com/RobertWHurst/navaros">
  </a>
  <a href="https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml">
    <img src="https://github.com/RobertWHurst/Navaros/actions/workflows/ci.yml/badge.svg">
  </a>
  <a href="https://github.com/sponsors/RobertWHurst">
    <img src="https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86">
  </a>
</p>

> __If you encounter a bug please [report it][bug-report].__

_Navaros is a simple and fast HTTP router for Go. It's designed to be simple to
use and get out of your way so you can focus on building awesome things._

```go
import (
  "net/http"
  "github.com/RobertWHurst/navaros"
  "github.com/RobertWHurst/navaros/middleware/json"
)

func main() {
  router := navaros.New()

  widgets := []*Widget{}

  router.Use(json.Middleware(nil))

  router.Post("/widget", func (ctx *navaros.Context) {
    widget := &Widget{}
    if ctx.UnmarshalRequestBody(widget); err != nil {
      ctx.Status = 404
      ctx.Body = map[string]string{"error": "invalid request body"}
      return
    }
    
    widget.ID = GenId()
    widgets = append(widgets, widget)

    ctx.Status = 201
  })

  router.Get("/widget", func (ctx *navaros.Context) {
    ctx.Status = 200
    ctx.Body = widgets
  })

  router.Get("/widget/:id", func (ctx *navaros.Context) {
    id := ctx.Params().Get("id")

    for _, widget := range widgets {
      if widget.ID == id {
        ctx.Status = 200
        ctx.Body = widget
        return
      }
    }

    ctx.Status = 404
  })

  server := http.Server{
    Addr: ":8080",
    Handler: router
  }

  server.ListenAndServe()
}
```

## Features

- Simple and fast
- Middleware support
- Parameterized routes
- Streaming request and response bodies
- Marshal and unmarshal API for body related middleware
- Unmarshal request bodies into structs
- handler panic recovery
- Request and response data organized into a single context object
- context support for ending handler logic after cancelled requests
- No dependencies other than the standard library
- Powerful route pattern matching
- Nestable routers

## Installation

```sh
go get github.com/RobertWHurst/navaros
```

## Usage

Navaros is designed to be simple to use and get out of your way so you can focus
on building awesome things. It's API is designed to be simple and intuitive.

### Creating a router

In order to get started you need to create a router. A router is an http.Handler
so to use it you just need to assign it as http server's handler.

```go
router := navaros.New()

server := http.Server{
  Addr: ":8080",
  Handler: router
}

server.ListenAndServe()
```

### Adding middleware

Let's say your going to be building an API and you want to use JSON as your
request and response body format. Navaros comes with JSON middleware for this
purpose. Now the request body will be buffered, and you will be able to
unmarshal it into a struct within you handlers.

```go
router.Use(json.Middleware(nil))
```

### Adding a route and unmarshaling a request body

Now let's add a handler to create a widget. We'll use the `Post` method to
register a handler for the `POST` method on the `/widget` path.

```go
type Widget struct {
  Name string `json:"name"`
  ...
}

router.Post("/widget", func (ctx *navaros.Context) {
  widget := &Widget{}
  if ctx.UnmarshalRequestBody(widget); err != nil {
    ctx.Status = 400
    ctx.Body = map[string]string{"error": "invalid request body"}
    return
  }
  
  ...

  ctx.Status = 201
})
```

### Adding a parameterized route

Now let's add a handler to get a widget by it's id. We'll use the `Get` method
to register a handler for the `GET` method on the `/widget/:id` path. The `:id`
portion of the path is a parameter. Parameters are accessible via the `Params`
method on the context.

```go
router.Get("/widget/:id", func (ctx *navaros.Context) {
  id := ctx.Params().Get("id")

  ...
})
```

Note that you can do much more with route patterns. We'll cover that later in
the readme.

### Context cancellation

Now let's say we have a handler that might be doing some expensive work. We
don't want to waste resources on requests that have been cancelled. For any
logic in your handler that supports go's context API you can actually pass
the handler context to it as ctx implements the context interface.

```go
router.Post("/report", func (ctx *navaros.Context) {
  reportCfg := &ReportConfiguration{}
  if err := ctx.UnmarshalRequestBody(reportCfg); err != nil {
    ctx.Status = 400
    ctx.Body = map[string]string{"error": "invalid request body"}
    return
  }
  
  report, err := reportGenerator.GenerateReport(ctx, reportCfg)

  ctx.Status = 200
  ctx.Body = report
})
```

If the request is cancelled the GenerateReport function can check the context
for cancellation and return early.

### Streaming request and response bodies

Navaros supports streaming request and response bodies. This is useful for
handling large files or streaming data. Provided you are not using Middleware
that buffers the request body you can read the bytes as they come in.

```go
router.Post("/upload", func (ctx *navaros.Context) {
  ctx.MaxRequestBodySize = 1024 * 1024 * 100 // 100MB

  reader := ctx.RequestBodyReader()

  ...
})
```

You can also stream the response body.

```go
router.Get("/download/:id", func (ctx *navaros.Context) {
  id := ctx.Params().Get("id")
  fileReader, err := fileManager.GetFileReader(id)
  if err != nil {
    if err == fileManager.ErrFileNotFound {
      ctx.Status = 404
      return
    }
    panic(err)
  }

  ctx.Status = 200
  ctx.Body = fileReader
})
```

### Nesting routers

Navaros supports nesting routers. This is useful for breaking up your routes
into logical groups. For example you might break your resources into packages -
each package having it's own router.

Your root router might look something like this.

```go
router := navaros.New()

router.Use(json.Middleware(nil))

router.Use(widgets.Router)
router.Use(gadgets.Router)
router.Use(gizmos.Router)
```
And each of your resource routers might look something like this.

```go
package widgets

import (
  "github.com/RobertWHurst/navaros"
)

var Router = navaros.New()

func init() {
  Router.Post("/widget", CreateWidget)
  Router.Get("/widget", GetWidgets)
  Router.Get("/widget/:id", GetWidgetByID)
}

func CreateWidget(ctx *navaros.Context) {
  ...
}

func GetWidgets(ctx *navaros.Context) {
  ...
}

func GetWidgetByID(ctx *navaros.Context) {
  ...
}
```

### Route patterns

Navaros supports fairly powerful route patterns. The following is a list of
supported pattern chunk types.

- Static - `/a/b/c` - Matches the exact path
- Wildcard - `/a/*/c` - Pattern segments with a single `*` match any path segment
- Dynamic - `/a/:b/c` - Pattern segments prefixed with `:` match any path segment
  and the value of this segment from the matched path is available via the
  `Params` method, and will be filled under a key matching the name of the
  pattern segment, ie: pattern of `/a/:b/c` will match `/a/1/c` and the value
  of `b` in the params will be `1`

Pattern chunks can also be suffixed with additional modifiers.

- `?` - Optional - `/a/:b?/c` - Matches `/a/c` and `/a/1/c`
- `*` - Greedy - `/a/:b*/c` - Matches `/a/c` and `/a/1/2/3/c`
- `+` - One or more - `/a/:b+/c` - Matches `/a/1/c` and `/a/1/2/3/c` but not `/a/c`

You can also provide a regular expression to restrict matches for a pattern chunk.

- `/a/:b(\d+)/c` - Matches `/a/1/c` and `/a/2/c` but not `/a/b/c`

You can escape any of the special characters used by these operators by prefixing
them with a `\`.

- `/a/\:b/c` - Matches `/a/:b/c`

And all of these can be combined.

- `/a/:b(\d+)/*?/(d|e)+` - Matches `/a/1/d`, `/a/1/e`, `/a/2/c/d/e/f/g`, and
`/a/3/1/d` but not `/a/b/c`, `/a/1`, or `/a/1/c/f`

This is all most likely overkill, but if you ever need it, it's here.

### Handler and Middleware ordering

Handlers and middleware are executed in the order they are added to the router.
This means that a handler added before another will always be checked for a
match against the incoming request first regardless of the path pattern. This
means you can easily predict how your handlers will be executed.

It also means that your handlers with more specific patterns should be
added before any others that may share a common match.

```go
router.Get("/album/:id(\d{24})", GetWidgetByID)
router.Get("/album/:name", GetWidgetsByName)
```

### Binding different HTTP methods

The router has a method for the following HTTP methods.

- `Get` - Handles `GET` requests - Should be used for reading
- `Post` - Handles `POST` requests - Should be used for creating
- `Put` - Handles `PUT` requests - Should be used for updating
- `Patch` - Handles `PATCH` requests - Should be used for updating
- `Delete` - Handles `DELETE` requests - Should be used for deleting
- `Head` - Handles `HEAD` requests - Should be used checking for existence
- `Options` - Handles `OPTIONS` requests - Used for CORS

There is also a few non HTTP method methods.

- `All` - Handles all requests matching the pattern regardless of method
- `Use` - Handles all requests like all, but a pattern can be omitted

### Passing a request on to the next handler - middleware

If you want to pass a request on to the next handler you can call the `Next`
method on the context. This will execute the next handler in the chain.

```go
router.All("/admin/**", func (ctx *navaros.Context) {
  if !IsAuthenticated(ctx) {
    ctx.Status = 401
    return
  }
  ctx.Next()
})
```

This is what makes a handler into middleware.

Interestingly you can also add logic after the `Next` call. This is useful for
doing things with the response after a handler executed during the `Next` call
has a written it.

```go
router.Use(func (ctx *navaros.Context) {
  ctx.Next()
  if ctx.Status == 404 && ctx.Body == nil {
    ctx.Status = 404
    ctx.Body = map[string]string{"error": "not found"}
  }
})
```

## Roadmap

Now that the core of Navaros is complete I'm going to be focusing on adding more
middleware and adding websocket support.

Websockets are going to be an important part of this project as I want to
support building realtime applications with Navaros.

## Help Welcome

If you want to support this project by throwing be some coffee money It's
greatly appreciated.

[![sponsor](https://img.shields.io/static/v1?label=Sponsor&message=%E2%9D%A4&logo=GitHub&color=%23fe8e86)](https://github.com/sponsors/RobertWHurst)

If your interested in providing feedback or would like to contribute please feel
free to do so. I recommend first [opening an issue][feature-request] expressing
your feedback or intent to contribute a change, from there we can consider your
feedback or guide your contribution efforts. Any and all help is greatly
appreciated since this is an open source effort after all.

Thank you!

[bug-report]: https://github.com/RobertWHurst/Relign/issues/new?template=bug_report.md
[feature-request]: https://github.com/RobertWHurst/Relign/issues/new?template=feature_request.md
