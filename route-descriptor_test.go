package navaros_test

import (
	"encoding/json"
	"testing"

	"github.com/RobertWHurst/navaros"
	"github.com/stretchr/testify/assert"
)

func TestRouteDescriptorMarshalJSON(t *testing.T) {
	pattern, err := navaros.NewPattern("/a/b/c")
	assert.Nil(t, err)

	r := &navaros.RouteDescriptor{
		Method:  navaros.Get,
		Pattern: pattern,
	}

	bytes, err := r.MarshalJSON()
	assert.Nil(t, err)

	jsonData := map[string]any{}
	err = json.Unmarshal(bytes, &jsonData)
	assert.Nil(t, err)

	assert.Equal(t, "GET", jsonData["Method"])
	assert.Equal(t, "/a/b/c", jsonData["Pattern"])
}

func TestRouteDescriptorUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{"Method":"GET","Pattern":"/a/b/c"}`)

	r := &navaros.RouteDescriptor{}
	err := r.UnmarshalJSON(jsonData)

	assert.Nil(t, err)
	assert.Equal(t, navaros.Get, r.Method)
	assert.Equal(t, "/a/b/c", r.Pattern.String())
}
