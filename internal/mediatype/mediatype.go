// Package mediatype compares Content-Type header values against media types.
package mediatype

import (
	"strings"
)

// Is reports whether a Content-Type header value names the given media type.
//
// Comparison follows RFC 9110 section 8.3.1, which makes the type and subtype
// tokens case-insensitive and carries parameters separately from the
// type/subtype pair that identifies the media type. RFC 9110 does caution that
// a parameter may be significant depending on the media type's own
// registration; for these three it is not — RFC 8259 section 11 defines no
// parameters at all for application/json — so "application/json; charset=utf-8"
// names "application/json".
//
// The structured syntax suffix convention (RFC 6838 section 4.2.8) also
// matches, so "application/problem+json" and "application/vnd.api+json" both
// name "application/json". This is deliberately wider than the IANA structured
// syntax suffix registry: of the three suffixes at issue here, only "+json" is
// registered, and "+msgpack" and "+protobuf" are honoured anyway, because the
// suffix is how a client says its type is structured as that format.
//
// mediaType must be a bare lowercase type/subtype pair, e.g. "application/json".
func Is(header string, mediaType string) bool {
	base, ok := baseType(header)
	if !ok {
		return false
	}
	if base == mediaType {
		return true
	}

	slash := strings.IndexByte(mediaType, '/')
	if slash == -1 {
		return false
	}
	prefix, subtype := mediaType[:slash+1], mediaType[slash+1:]

	// A suffix match needs a non-empty subtype root before the "+", so
	// "application/+json" does not name "application/json".
	if len(base) <= len(prefix)+len(subtype)+1 {
		return false
	}
	return strings.HasPrefix(base, prefix) && strings.HasSuffix(base, "+"+subtype)
}

// baseType extracts the lowercased type/subtype from a Content-Type header,
// discarding parameters.
//
// Only the bytes before the first ";" are ever needed: RFC 9110 section 8.3
// defines the media type as token "/" token, so no quoted string can precede
// the first parameter separator, and a malformed parameter cannot hide the
// media type naming it. Requiring both halves to be tokens is what keeps a
// value that is not a single media type — "application/pdf, junk+json", or a
// comma-separated list — from being read as one and matched on its tail.
func baseType(header string) (string, bool) {
	base, _, _ := strings.Cut(header, ";")
	base = strings.Trim(base, " \t")

	// RFC 6838 section 4.2 limits a type or subtype name to 127 characters, so
	// nothing longer than the pair plus its slash can name a media type. Checked
	// before folding case so that a megabyte-long header — Content-Type is
	// client-supplied, and net/http allows 1 MB of headers by default — cannot
	// force a megabyte-long allocation on every request.
	if len(base) > 255 {
		return "", false
	}
	base = strings.ToLower(base)

	typ, subtype, ok := strings.Cut(base, "/")
	if !ok || !isToken(typ) || !isToken(subtype) {
		return "", false
	}

	// A wildcard names a range of media types rather than one, and belongs in
	// Accept rather than Content-Type. Rejecting it here is what keeps it out of
	// the suffix rule, where "application/*+json" would otherwise name
	// "application/json".
	if strings.IndexByte(base, '*') != -1 {
		return "", false
	}
	return base, true
}

// isToken reports whether s is a non-empty token, per RFC 9110 section 5.6.2.
func isToken(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !isTChar(s[i]) {
			return false
		}
	}
	return true
}

// isTChar reports whether c is a tchar, per RFC 9110 section 5.6.2.
func isTChar(c byte) bool {
	switch c {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
