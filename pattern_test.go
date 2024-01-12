package navaros_test

import (
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/stretchr/testify/assert"
)

func TestPatternStaticChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/b/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternStaticChunkWithTrailingSlash(t *testing.T) {
	p, err := navaros.NewPattern("/a/b/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/b/c/")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternStaticOptionalChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/b?/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternStaticOneOrMoreChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/b+/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/b/b/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/c")
	assert.False(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternStaticZeroOrMoreChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/b*/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/b/b/b/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternDynamicChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])
}

func TestPatternDynamicChunkWithTrailingSlash(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c/")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])
}

func TestPatternDynamicOptionalChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b?/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/c")
	assert.True(t, isMatch)
	assert.Empty(t, "", params["b"])

	params, isMatch = p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])
}

func TestPatternDynamicOneOrMoreChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b+/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])

	params, isMatch = p.Match("/a/123/456/789/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123/456/789", params["b"])

	params, isMatch = p.Match("/a/c")
	assert.False(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternDynamicZeroOrMoreChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b*/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/c")
	assert.True(t, isMatch)
	assert.Empty(t, params["b"])

	params, isMatch = p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])

	params, isMatch = p.Match("/a/123/456/789/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123/456/789", params["b"])
}

func TestPatternDynamicChunkWithCustomSubPattern(t *testing.T) {
	p, err := navaros.NewPattern("/a/:b(\\d+)/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Equal(t, "123", params["b"])

	params, isMatch = p.Match("/a/abc/c")
	assert.False(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternWildcardChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/*/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternWildcardChunkWithTrailingSlash(t *testing.T) {
	p, err := navaros.NewPattern("/a/*/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c/")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternWildcardOptionalChunk(t *testing.T) {
	p, err := navaros.NewPattern("/a/*?/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)
}

func TestPatternWildcardChunkWithCustomSubPattern(t *testing.T) {
	p, err := navaros.NewPattern("/a/*(\\d+)/c")
	assert.Nil(t, err)

	params, isMatch := p.Match("/a/123/c")
	assert.True(t, isMatch)
	assert.Empty(t, params)

	params, isMatch = p.Match("/a/abc/c")
	assert.False(t, isMatch)
	assert.Empty(t, params)
}
