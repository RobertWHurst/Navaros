package navaros_test

import (
	"strings"
	"testing"

	"github.com/RobertWHurst/navaros"
)

func TestNewPattern(t *testing.T) {
	pattern, err := navaros.NewPattern("/a/b/c")
	if err != nil {
		t.Error(err)
	}
	if pattern == nil {
		t.Error("pattern is nil")
	}
}

func TestNewPatternWithInvalid(t *testing.T) {
	_, err := navaros.NewPattern("/a/b/c(")
	if err == nil {
		t.Error("expected error")
	}
}

func TestPatternMatch(t *testing.T) {
	cases := []struct {
		message        string
		patternStr     string
		pathStr        string
		shouldMatch    bool
		expectedError  string
		expectedParams map[string]string
	}{
		{"should match a static path",
			"/a/b/c", "/a/b/c", true, "",
			nil,
		},
		{"should match a static path with a trailing slash",
			"/a/b/c", "/a/b/c/", true, "",
			nil,
		},
		{"should match a static path with an optional chunk",
			"/a/b?/c", "/a/c", true, "",
			nil,
		},
		{"should match a static path with an optional chunk and a trailing slash",
			"/a/b?/c", "/a/c/", true, "",
			nil,
		},
		{"should match a static path with a one or more chunks",
			"/a/b+/c", "/a/b/c", true, "",
			nil,
		},
		{"should match a static path with a one or more chunks and a trailing slash",
			"/a/b+/c", "/a/b/c/", true, "",
			nil,
		},
		{"should match a static path with a zero or more chunks",
			"/a/b*/c", "/a/c", true, "",
			nil,
		},
		{"should match a static path with a zero or more chunks and a trailing slash",
			"/a/b*/c", "/a/c/", true, "",
			nil,
		},
		{"should match a dynamic path",
			"/a/:b/c", "/a/123/c", true, "",
			map[string]string{
				"b": "123",
			},
		},
		{"should match a dynamic path with a trailing slash",
			"/a/:b/c", "/a/123/c/", true, "",
			map[string]string{
				"b": "123",
			},
		},
		{"should match a dynamic path with an optional chunk",
			"/a/:b?/c", "/a/c", true, "",
			map[string]string{
				"b": "",
			},
		},
		{"should match a dynamic path with an optional chunk and a trailing slash",
			"/a/:b?/c", "/a/c/", true, "",
			map[string]string{
				"b": "",
			},
		},
		{"should match a dynamic path with a one or more chunks",
			"/a/:b+/c", "/a/123/d/c", true, "",
			map[string]string{
				"b": "123/d",
			},
		},
		{"should match a dynamic path with a one or more chunks and a trailing slash",
			"/a/:b+/c", "/a/123/d/c/", true, "",
			map[string]string{
				"b": "123/d",
			},
		},
		{"should match a dynamic path with a zero or more chunks",
			"/a/:b*/c", "/a/c", true, "",
			map[string]string{
				"b": "",
			},
		},
		{"should match a dynamic path with a zero or more chunks and a trailing slash",
			"/a/:b*/c", "/a/c/", true, "",
			map[string]string{
				"b": "",
			},
		},
		{"should match a dynamic path with a custom sub pattern",
			"/a/:b(\\d+)/c", "/a/123/c", true, "",
			map[string]string{
				"b": "123",
			},
		},
		{"should match a dynamic path with a custom sub pattern and a trailing slash",
			"/a/:b(\\d+)/c", "/a/123/c/", true, "",
			map[string]string{
				"b": "123",
			},
		},
		{"should match a wildcard path",
			"/a/*/c", "/a/123/c", true, "",
			nil,
		},
		{"should match a wildcard path with a trailing slash",
			"/a/*/c", "/a/123/c/", true, "",
			nil,
		},
		{"should match a wildcard path with an optional chunk",
			"/a/*?/c", "/a/c", true, "",
			nil,
		},
		{"should match a wildcard path with an optional chunk and a trailing slash",
			"/a/*?/c", "/a/c/", true, "",
			nil,
		},
		{"should match a wildcard path with a custom sub pattern",
			"/a/*(\\d+)/c", "/a/123/c", true, "",
			nil,
		},
		{"should match a wildcard path with a custom sub pattern and a trailing slash",
			"/a/*(\\d+)/c", "/a/123/c/", true, "",
			nil,
		},
		{"should maatch a path with a custom sub pattern on its own",
			"/a/(\\d+)/c", "/a/123/c", true, "",
			nil,
		},
		{"should not match a non matching static path",
			"/a/b/c", "/a/b/d", false, "",
			nil,
		},
		{"should not match a non matching static path with a trailing slash",
			"/a/b/c", "/a/b/d/", false, "",
			nil,
		},
		{"should not match a non matching static path with an optional chunk",
			"/a/b?/c", "/a/d", false, "",
			nil,
		},
		{"should not match a non matching static path with an optional chunk and a trailing slash",
			"/a/b?/c", "/a/d/", false, "",
			nil,
		},
		{"should not match a non matching static path with a one or more chunks",
			"/a/b+/c", "/a/b/d", false, "",
			nil,
		},
		{"should not match a non matching static path with a one or more chunks and a trailing slash",
			"/a/b+/c", "/a/b/d/", false, "",
			nil,
		},
		{"should not match a non matching static path with a zero or more chunks",
			"/a/b*/c", "/a/d", false, "",
			nil,
		},
		{"should not match a non matching static path with a zero or more chunks and a trailing slash",
			"/a/b*/c", "/a/d/", false, "",
			nil,
		},
		{"should not match a non matching dynamic path",
			"/a/:b/c", "/a/123/d", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a trailing slash",
			"/a/:b/c", "/a/123/d/", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with an optional chunk",
			"/a/:b?/c", "/a/d", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with an optional chunk and a trailing slash",
			"/a/:b?/c", "/a/d/", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a one or more chunks",
			"/a/:b+/c", "/a/123/d", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a one or more chunks and a trailing slash",
			"/a/:b+/c", "/a/123/d/", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a zero or more chunks",
			"/a/:b*/c", "/a/d", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a zero or more chunks and a trailing slash",
			"/a/:b*/c", "/a/d/", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a custom sub pattern",
			"/a/:b(\\d+)/c", "/a/abc/c", false, "",
			nil,
		},
		{"should not match a non matching dynamic path with a custom sub pattern and a trailing slash",
			"/a/:b(\\d+)/c", "/a/abc/c/", false, "",
			nil,
		},
		{"should not match a non matching wildcard path",
			"/a/*/c", "/a/123/d", false, "",
			nil,
		},
		{"should not match a non matching wildcard path with a trailing slash",
			"/a/*/c", "/a/123/d/", false, "",
			nil,
		},
		{"should not match a non matching wildcard path with an optional chunk",
			"/a/*?/c", "/a/d", false, "",
			nil,
		},
		{"should not match a non matching wildcard path with an optional chunk and a trailing slash",
			"/a/*?/c", "/a/d/", false, "",
			nil,
		},
		{"should not match a non matching wildcard path with a custom sub pattern",
			"/a/*(\\d+)/c", "/a/abc/c", false, "",
			nil,
		},
		{"should not match a non matching wildcard path with a custom sub pattern and a trailing slash",
			"/a/*(\\d+)/c", "/a/abc/c/", false, "",
			nil,
		},
		{"should error if pattern does not start with a slash",
			"a/b/c", "/a/b/c", false, "must start with a leading slash",
			nil,
		},
		{"should error if the pattern contains an unclosed sub pattern",
			"/a/:b(\\d+/c", "/a/123/c", false, "invalid named capture",
			nil,
		},
		{"should error if a dynamic chunk does not have a name",
			"/a/:(\\d+)/c", "/a/123/c", false, "dynamic chunks must have a name",
			nil,
		},
	}

	for _, c := range cases {
		t.Run(c.message, func(t *testing.T) {
			pattern, err := navaros.NewPattern(c.patternStr)
			if c.expectedError == "" {
				params, isMatch := pattern.Match(c.pathStr)
				if isMatch != c.shouldMatch {
					t.Errorf("expected isMatch to be %v but got %v", c.shouldMatch, isMatch)
				}
				if c.expectedParams == nil {
					if len(params) != 0 {
						t.Errorf("expected params to be empty but got %v", params)
					}
				} else {
					if params == nil {
						t.Errorf("expected params to be %v but got nil", c.expectedParams)
					}
				}
			} else {
				if err == nil {
					t.Error("expected error")
				}
				if !strings.Contains(err.Error(), c.expectedError) {
					t.Errorf("expected error to contain %v but got %v", c.expectedError, err.Error())
				}
			}
		})
	}
}

func TestPatternString(t *testing.T) {
	pattern, err := navaros.NewPattern("/a/b/c")
	if err != nil {
		t.Error(err)
	}
	if pattern.String() != "/a/b/c" {
		t.Error("pattern string does not match")
	}
}

func TestPatternPath(t *testing.T) {
	pattern, err := navaros.NewPattern("/users/:id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := navaros.RequestParams{"id": "123"}
	path, err := pattern.Path(params, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/users/123" {
		t.Errorf("expected /users/123, got %s", path)
	}
}

func TestPatternRootPath(t *testing.T) {
	pattern, err := navaros.NewPattern("/")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	params, matched := pattern.Match("/")
	if !matched {
		t.Error("expected root path to match")
	}
	if len(params) != 0 {
		t.Errorf("expected no params, got %v", params)
	}
}

func TestPatternEmptyPathMatchesRoot(t *testing.T) {
	pattern, err := navaros.NewPattern("/")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	_, matched := pattern.Match("")
	if !matched {
		t.Error("expected empty string to match root pattern")
	}
}

func TestPatternTrailingSlashNormalization(t *testing.T) {
	pattern, err := navaros.NewPattern("/users")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	_, matched1 := pattern.Match("/users")
	_, matched2 := pattern.Match("/users/")

	if !matched1 {
		t.Error("expected /users to match")
	}
	if !matched2 {
		t.Error("expected /users/ to match")
	}
}

func TestPatternCaseSensitivity(t *testing.T) {
	pattern, err := navaros.NewPattern("/users")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	_, matched := pattern.Match("/Users")
	if matched {
		t.Error("expected case-sensitive matching - /Users should not match /users")
	}
}

func TestPatternSpecialCharactersInStaticPath(t *testing.T) {
	pattern, err := navaros.NewPattern("/api-v1/users.json")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	_, matched := pattern.Match("/api-v1/users.json")
	if !matched {
		t.Error("expected special characters in path to work")
	}
}

func TestPatternMultipleParams(t *testing.T) {
	pattern, err := navaros.NewPattern("/users/:userId/posts/:postId")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	params, matched := pattern.Match("/users/123/posts/456")
	if !matched {
		t.Error("expected to match")
	}
	if params.Get("userId") != "123" {
		t.Errorf("expected userId=123, got %s", params.Get("userId"))
	}
	if params.Get("postId") != "456" {
		t.Errorf("expected postId=456, got %s", params.Get("postId"))
	}
}

func TestPatternWildcardAtEnd(t *testing.T) {
	pattern, err := navaros.NewPattern("/files/*")
	if err != nil {
		t.Fatalf("failed to create pattern: %v", err)
	}

	_, matched := pattern.Match("/files/doc.pdf")
	if !matched {
		t.Error("expected /files/doc.pdf to match")
	}
}

func TestPatternEmptyParamName(t *testing.T) {
	_, err := navaros.NewPattern("/users/:/posts")
	if err == nil {
		t.Error("expected error for empty param name")
	}
}
