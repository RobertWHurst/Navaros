// Package navaros is a lightweight, flexible HTTP router for Go.
//
// It provides powerful route matching, middleware support, and efficient request handling
// with zero external dependencies. Designed for simplicity and performance, Navaros
// implements the standard http.Handler interface and supports nested routers, parameterized
// routes, streaming bodies, and context cancellation.
//
// # Quick Start
//
// router := navaros.NewRouter()
//
//	router.Get("/users/:id", func(ctx *navaros.Context) {
//	    id := ctx.Params().Get("id")
//	    ctx.Status = 200
//	    ctx.Body = map[string]string{"id": id}
//	})
//
// http.ListenAndServe(":8080", router)
//
// # Features
//
// - Sequential pattern matching with regex for dynamic segments
// - Middleware chain with Next() for pre/post processing
// - Parameter extraction with wildcards, regex, optional/greedy modifiers
// - JSON body parsing via middleware
// - Panic recovery and error handling
// - Context implements context.Context for cancellation
// - Nestable sub-routers for modular code
// - Streaming request/response bodies
// - No dependencies beyond stdlib
//
// # Middleware
//
// Use middleware for auth, logging, body parsing:
//
//	router.Use(func(ctx *navaros.Context) {
//	    // Pre-handler logic
//	    ctx.Next()
//	    // Post-handler logic
//	})
//
// Built-in JSON middleware: import "github.com/RobertWHurst/navaros/middleware/json"; router.Use(json.Middleware(nil))
//
// # Route Patterns
//
// - Static: /users
// - Param: /users/:id
// - Wildcard: /files/*
// - Optional: /users/:id?
// - Regex: /users/:id(\\d+)
// - Combine: /api/:v(\\d+)/:action*/static
//
// Specific patterns before general.
//
// # Context
//
// Unified access to request/response/params:
//
// ctx.Params().Get("id")
// ctx.UnmarshalRequestBody(&user)
// ctx.Status = 200
// ctx.Body = user
//
// Supports Set/Get for per-request data.
//
// See https://pkg.go.dev/github.com/RobertWHurst/navaros for full API docs.
package navaros
