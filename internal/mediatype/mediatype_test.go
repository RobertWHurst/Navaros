package mediatype_test

import (
	"strings"
	"testing"

	"github.com/RobertWHurst/navaros/internal/mediatype"
)

func TestIs(t *testing.T) {
	tests := []struct {
		header    string
		mediaType string
		want      bool
	}{
		// Exact matches still hold.
		{"application/json", "application/json", true},
		{"application/msgpack", "application/msgpack", true},
		{"application/protobuf", "application/protobuf", true},

		// Parameters are carried separately from the type/subtype pair that
		// identifies the media type, and application/json registers none at all
		// (RFC 8259 section 11). The first is what OkHttp sends for any string
		// request body, which is how this reached production.
		{"application/json; charset=utf-8", "application/json", true},
		{"application/json;charset=utf-8", "application/json", true},
		{"application/json ; charset=utf-8", "application/json", true},
		{"application/msgpack; charset=binary", "application/msgpack", true},
		{"application/protobuf; proto=Widget", "application/protobuf", true},

		// Type and subtype are case-insensitive.
		{"Application/JSON", "application/json", true},
		{"APPLICATION/JSON; CHARSET=UTF-8", "application/json", true},

		// The structured syntax suffix convention (RFC 6838 section 4.2.8). Of
		// the three suffixes at issue here only "+json" is IANA-registered;
		// "+msgpack" and "+protobuf" are honoured as a deliberate
		// generalization, not as registered suffixes.
		{"application/problem+json", "application/json", true},
		{"application/vnd.api+json; charset=utf-8", "application/json", true},
		{"application/json-patch+json", "application/json", true},
		{"application/vnd.custom+msgpack", "application/msgpack", true},

		// Malformed parameters must not hide a usable media type.
		{"application/json; charset", "application/json", true},
		{"application/json;", "application/json", true},
		{"application/json; charset=utf-8; charset=utf-16", "application/json", true},
		{"application/json; charset=\"utf-8\"", "application/json", true},

		// Non-matches.
		{"", "application/json", false},
		{"   ", "application/json", false},
		{"text/plain", "application/json", false},
		{"text/json", "application/json", false},
		{"application/xml", "application/json", false},
		{"application/octet-stream", "application/json", false},
		{"multipart/form-data; boundary=x", "application/json", false},
		{"application/jsonish", "application/json", false},
		{"application/json+zip", "application/json", false},
		{"application/msgpack", "application/json", false},
		{"application/json", "application/msgpack", false},
		{"; charset=utf-8", "application/json", false},

		// A value that is not a single media type must not be matched on its
		// tail. Type and subtype are each a token, so a comma, a space or a
		// second slash means this is not one media type, whatever it ends with.
		{"application/octet-stream, x+json", "application/json", false},
		{"application/pdf, junk+json", "application/json", false},
		{"application/octet-stream x+json", "application/json", false},
		{"application/x/y+json", "application/json", false},
		{"application/json, text/plain", "application/json", false},
		{"application/protobuf, evil+msgpack", "application/msgpack", false},

		// A suffix needs a subtype root in front of it.
		{"application/+json", "application/json", false},

		// Wildcards name no concrete media type. The suffixed forms are the ones
		// that matter: "*" is a valid token character, so without an explicit
		// wildcard rejection they reach the suffix rule and match.
		{"*/*", "application/json", false},
		{"application/*", "application/json", false},
		{"application/*+json", "application/json", false},
		{"application/*+msgpack", "application/msgpack", false},
		{"application/*+protobuf", "application/protobuf", false},

		// A suffix only names the target when the type matches too. Without the
		// type-prefix check, "text/x+json" would name "application/json".
		{"text/x+json", "application/json", false},
		{"text/vnd.custom+json", "application/json", false},
		{"image/svg+xml", "application/xml", false},

		// A type or subtype name is at most 127 characters (RFC 6838 section
		// 4.2), so an over-long value names nothing and is rejected before its
		// case is folded. The subtype here is otherwise a well-formed "+json"
		// suffix under the right type, so it would match without the limit.
		{"application/" + strings.Repeat("a", 250) + "+json", "application/json", false},
		{strings.Repeat("a", 130) + "/" + strings.Repeat("b", 130) + "+json", "application/json", false},

		// Aliases seen in the wild are deliberately not accepted: which of them
		// to honour is a decision about what the library accepts, not something
		// RFC 9110 settles.
		{"application/x-msgpack", "application/msgpack", false},
		{"application/vnd.msgpack", "application/msgpack", false},
		{"application/x-protobuf", "application/protobuf", false},
		{"application/vnd.google.protobuf", "application/protobuf", false},
	}

	for _, test := range tests {
		got := mediatype.Is(test.header, test.mediaType)
		if got != test.want {
			t.Errorf("Is(%q, %q) = %v, want %v", test.header, test.mediaType, got, test.want)
		}
	}
}
