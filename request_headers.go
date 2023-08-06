package navaros

import "strings"

type RequestHeaders map[string][]string

func (h RequestHeaders) Get(key string) string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			return strings.Join(v, ", ")
		}
	}
	return ""
}

func (h RequestHeaders) GetSlice(key string) []string {
	for k, v := range h {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return []string{}
}

func (h RequestHeaders) Set(key string, value string) {
	for k := range h {
		if strings.EqualFold(k, key) {
			h[k] = []string{value}
			return
		}
	}
	h[key] = []string{value}
}

func (h RequestHeaders) SetSlice(key string, value []string) {
	for k := range h {
		if strings.EqualFold(k, key) {
			h[k] = value
			return
		}
	}
	h[key] = value
}

func (h RequestHeaders) Delete(key string) {
	for k := range h {
		if strings.EqualFold(k, key) {
			delete(h, k)
			return
		}
	}
}
