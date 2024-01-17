package navaros

import (
	"encoding/json"
)

type RouteDescriptor struct {
	Method  HTTPMethod
	Pattern *Pattern
}

func (r *RouteDescriptor) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Method  HTTPMethod
		Pattern string
	}{
		Method:  r.Method,
		Pattern: r.Pattern.String(),
	})
}

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
