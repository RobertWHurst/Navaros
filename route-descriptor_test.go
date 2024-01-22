package navaros_test

import (
	"encoding/json"
	"testing"

	"github.com/RobertWHurst/navaros"
)

func TestRouteDescriptorMarshalJSON(t *testing.T) {
	pattern, err := navaros.NewPattern("/a/b/c")
	if err != nil {
		t.Errorf("Failed to create pattern: %s", err.Error())
	}

	r := &navaros.RouteDescriptor{
		Method:  navaros.Get,
		Pattern: pattern,
	}

	bytes, err := r.MarshalJSON()
	if err != nil {
		t.Errorf("Failed to marshal route descriptor: %s", err.Error())
	}

	jsonData := map[string]any{}
	err = json.Unmarshal(bytes, &jsonData)
	if err != nil {
		t.Errorf("Failed to unmarshal route descriptor: %s", err.Error())
	}

	if len(jsonData) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(jsonData))
	}
	if jsonData["Method"] != "GET" {
		t.Errorf("Expected Method to be GET, got %s", jsonData["Method"])
	}
	if jsonData["Pattern"] != "/a/b/c" {
		t.Errorf("Expected Pattern to be /a/b/c, got %s", jsonData["Pattern"])
	}
}

func TestRouteDescriptorUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{"Method":"GET","Pattern":"/a/b/c"}`)

	r := &navaros.RouteDescriptor{}
	if err := r.UnmarshalJSON(jsonData); err != nil {
		t.Errorf("Failed to unmarshal route descriptor: %s", err.Error())
	}

	if r.Method != navaros.Get {
		t.Errorf("Expected Method to be GET, got %s", r.Method)
	}
	if r.Pattern.String() != "/a/b/c" {
		t.Errorf("Expected Pattern to be /a/b/c, got %s", r.Pattern)
	}
}
