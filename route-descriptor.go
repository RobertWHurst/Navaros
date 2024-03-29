package navaros

import (
	"encoding/json"
)

// RouteDescriptor is a struct that is generated by the router when the public
// variants of the handler binding methods (named after their respective
// http method) are called. The router has a RouteDescriptors method which
// returns these objects, and can be used to build a api map, or pre-filter
// requests before they are passed to the router. This is most useful for
// libraries that wish to extend the functionality of Navaros.
type RouteDescriptor struct {
	Method  HTTPMethod
	Pattern *Pattern
}

// MarshalJSON returns the JSON representation of the route descriptor.
func (r *RouteDescriptor) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Method  HTTPMethod
		Pattern string
	}{
		Method:  r.Method,
		Pattern: r.Pattern.String(),
	})
}

// UnmarshalJSON parses the JSON representation of the route descriptor.
func (r *RouteDescriptor) UnmarshalJSON(data []byte) error {
	fromJSONStruct := struct {
		Method  HTTPMethod
		Pattern string
	}{}
	if err := json.Unmarshal(data, &fromJSONStruct); err != nil {
		return err
	}

	pattern, err := NewPattern(fromJSONStruct.Pattern)
	if err != nil {
		return err
	}

	r.Method = fromJSONStruct.Method
	r.Pattern = pattern

	return nil
}
