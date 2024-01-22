package navaros_test

import (
	"github.com/RobertWHurst/navaros"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pattern", func() {

	Describe("NewPattern", func() {
		It("should return a pattern", func() {
			pattern, err := navaros.NewPattern("/a/b/c")
			Expect(err).To(BeNil())
			Expect(pattern).ToNot(BeNil())
		})

		It("should return an error if the pattern is invalid", func() {
			_, err := navaros.NewPattern("/a/b/c(")
			Expect(err).ToNot(BeNil())
		})
	})

	DescribeTable("Match",
		func(patternStr string, pathStr string, shouldMatch bool, expectedError string, expectedParams map[string]string) {
			pattern, err := navaros.NewPattern(patternStr)
			if expectedError == "" {
				params, isMatch := pattern.Match(pathStr)
				Expect(isMatch).To(Equal(shouldMatch))
				if expectedParams == nil {
					Expect(params).To(BeEmpty())
				} else {
					Expect(params).To(Equal(expectedParams))
				}
			} else {
				Expect(err).To(MatchError(ContainSubstring(expectedError)))
			}
		},
		Entry("should match a static path",
			"/a/b/c", "/a/b/c", true, "",
			nil,
		),
		Entry("should match a static path with a trailing slash",
			"/a/b/c", "/a/b/c/", true, "",
			nil,
		),
		Entry("should match a static path with an optional chunk",
			"/a/b?/c", "/a/c", true, "",
			nil,
		),
		Entry(
			"should match a static path with an optional chunk and a trailing slash",
			"/a/b?/c", "/a/c/", true, "",
			nil,
		),
		Entry("should match a static path with a one or more chunks",
			"/a/b+/c", "/a/b/c", true, "",
			nil,
		),
		Entry("should match a static path with a one or more chunks and a trailing slash",
			"/a/b+/c", "/a/b/c/", true, "",
			nil,
		),
		Entry("should match a static path with a zero or more chunks",
			"/a/b*/c", "/a/c", true, "",
			nil,
		),
		Entry("should match a static path with a zero or more chunks and a trailing slash",
			"/a/b*/c", "/a/c/", true, "",
			nil,
		),
		Entry("should match a dynamic path",
			"/a/:b/c", "/a/123/c", true, "",
			map[string]string{
				"b": "123",
			},
		),
		Entry("should match a dynamic path with a trailing slash",
			"/a/:b/c", "/a/123/c/", true, "",
			map[string]string{
				"b": "123",
			},
		),
		Entry("should match a dynamic path with an optional chunk",
			"/a/:b?/c", "/a/c", true, "",
			map[string]string{
				"b": "",
			},
		),
		Entry("should match a dynamic path with an optional chunk and a trailing slash",
			"/a/:b?/c", "/a/c/", true, "",
			map[string]string{
				"b": "",
			},
		),
		Entry("should match a dynamic path with a one or more chunks",
			"/a/:b+/c", "/a/123/d/c", true, "",
			map[string]string{
				"b": "123/d",
			},
		),
		Entry("should match a dynamic path with a one or more chunks and a trailing slash",
			"/a/:b+/c", "/a/123/d/c/", true, "",
			map[string]string{
				"b": "123/d",
			},
		),
		Entry("should match a dynamic path with a zero or more chunks",
			"/a/:b*/c", "/a/c", true, "",
			map[string]string{
				"b": "",
			},
		),
		Entry("should match a dynamic path with a zero or more chunks and a trailing slash",
			"/a/:b*/c", "/a/c/", true, "",
			map[string]string{
				"b": "",
			},
		),
		Entry("should match a dynamic path with a custom sub pattern",
			"/a/:b(\\d+)/c", "/a/123/c", true, "",
			map[string]string{
				"b": "123",
			},
		),
		Entry("should match a dynamic path with a custom sub pattern and a trailing slash",
			"/a/:b(\\d+)/c", "/a/123/c/", true, "",
			map[string]string{
				"b": "123",
			},
		),
		Entry("should match a wildcard path",
			"/a/*/c", "/a/123/c", true, "",
			nil,
		),
		Entry("should match a wildcard path with a trailing slash",
			"/a/*/c", "/a/123/c/", true, "",
			nil,
		),
		Entry("should match a wildcard path with an optional chunk",
			"/a/*?/c", "/a/c", true, "",
			nil,
		),
		Entry("should match a wildcard path with an optional chunk and a trailing slash",
			"/a/*?/c", "/a/c/", true, "",
			nil,
		),
		Entry("should match a wildcard path with a custom sub pattern",
			"/a/*(\\d+)/c", "/a/123/c", true, "",
			nil,
		),
		Entry("should match a wildcard path with a custom sub pattern and a trailing slash",
			"/a/*(\\d+)/c", "/a/123/c/", true, "",
			nil,
		),
		Entry("should maatch a path with a custom sub pattern on its own",
			"/a/(\\d+)/c", "/a/123/c", true, "",
			nil,
		),
		Entry("should not match a non matching static path",
			"/a/b/c", "/a/b/d", false, "",
			nil,
		),
		Entry("should not match a non matching static path with a trailing slash",
			"/a/b/c", "/a/b/d/", false, "",
			nil,
		),
		Entry("should not match a non matching static path with an optional chunk",
			"/a/b?/c", "/a/d", false, "",
			nil,
		),
		Entry("should not match a non matching static path with an optional chunk and a trailing slash",
			"/a/b?/c", "/a/d/", false, "",
			nil,
		),
		Entry("should not match a non matching static path with a one or more chunks",
			"/a/b+/c", "/a/b/d", false, "",
			nil,
		),
		Entry("should not match a non matching static path with a one or more chunks and a trailing slash",
			"/a/b+/c", "/a/b/d/", false, "",
			nil,
		),
		Entry("should not match a non matching static path with a zero or more chunks",
			"/a/b*/c", "/a/d", false, "",
			nil,
		),
		Entry("should not match a non matching static path with a zero or more chunks and a trailing slash",
			"/a/b*/c", "/a/d/", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path",
			"/a/:b/c", "/a/123/d", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a trailing slash",
			"/a/:b/c", "/a/123/d/", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with an optional chunk",
			"/a/:b?/c", "/a/d", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with an optional chunk and a trailing slash",
			"/a/:b?/c", "/a/d/", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a one or more chunks",
			"/a/:b+/c", "/a/123/d", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a one or more chunks and a trailing slash",
			"/a/:b+/c", "/a/123/d/", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a zero or more chunks",
			"/a/:b*/c", "/a/d", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a zero or more chunks and a trailing slash",
			"/a/:b*/c", "/a/d/", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a custom sub pattern",
			"/a/:b(\\d+)/c", "/a/abc/c", false, "",
			nil,
		),
		Entry("should not match a non matching dynamic path with a custom sub pattern and a trailing slash",
			"/a/:b(\\d+)/c", "/a/abc/c/", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path",
			"/a/*/c", "/a/123/d", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path with a trailing slash",
			"/a/*/c", "/a/123/d/", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path with an optional chunk",
			"/a/*?/c", "/a/d", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path with an optional chunk and a trailing slash",
			"/a/*?/c", "/a/d/", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path with a custom sub pattern",
			"/a/*(\\d+)/c", "/a/abc/c", false, "",
			nil,
		),
		Entry("should not match a non matching wildcard path with a custom sub pattern and a trailing slash",
			"/a/*(\\d+)/c", "/a/abc/c/", false, "",
			nil,
		),
		Entry("should error if pattern does not start with a slash",
			"a/b/c", "/a/b/c", false, "must start with a leading slash",
			nil,
		),
		Entry("should error if the pattern contains an unclosed sub pattern",
			"/a/:b(\\d+/c", "/a/123/c", false, "invalid named capture",
			nil,
		),
		Entry("should error if a dynamic chunk does not have a name",
			"/a/:(\\d+)/c", "/a/123/c", false, "dynamic chunks must have a name",
			nil,
		),
	)

	Describe("String", func() {
		It("should return the pattern string", func() {
			pattern, err := navaros.NewPattern("/a/b/c")
			Expect(err).To(BeNil())
			Expect(pattern.String()).To(Equal("/a/b/c"))
		})
	})
})
