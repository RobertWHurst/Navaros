package navaros

import "strings"

type RequestParams map[string]string

func (p RequestParams) Get(key string) string {
	for k, v := range p {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}
