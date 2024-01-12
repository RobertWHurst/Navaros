<h1 align="center">
  <!-- TODO: Logo -->
</h1>

<p align="center">
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

__If you encounter a bug please [report it][bug-report].__

Navaros is a router package for go with a focus on robust middleware patterns and predictable route matching.

```go
import (
  "net/http"
  "github.com/RobertWHurst/navaros"
)

func main() {
  router := navaros.New()

  router.Get("/", helloWorld)

  server := http.Server{
    Addr: ":8080",
    Handler: router
  }

  server.ListenAndServe()
}

func helloWorld(ctx *navaros.Context) {
  ctx.Body = "Hello World"
}
```

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
