package flags

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errEmptyRequestHeader    = errors.New("empty request header value")
	errInvalidRequestHeader  = errors.New("invalid request header value")
	errEmptyResponseHeader   = errors.New("empty response header value")
	errInvalidResponseHeader = errors.New("invalid response header value")
)

// RequestHeadersFlag holds header information for http request.
type RequestHeadersFlag []string

func (f *RequestHeadersFlag) String() string {
	return fmt.Sprintf("%s", *f)
}

// Set appends valid header values to RequestHeadersFlag.
func (f *RequestHeadersFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errEmptyRequestHeader
	}
	if strings.Count(value, ":") != 1 {
		return fmt.Errorf("%w: %s", errInvalidRequestHeader, value)
	}
	if len(strings.FieldsFunc(value, func(c rune) bool { return c == ':' })) != 2 {
		return fmt.Errorf("%w: %s", errInvalidRequestHeader, value)
	}
	*f = append(*f, value)
	return nil
}

// ResponseHeadersFlag ...
type ResponseHeadersFlag []string

func (f *ResponseHeadersFlag) String() string {
	return fmt.Sprintf("%s", *f)
}

// Set appends valid response headers for lookup.
func (f *ResponseHeadersFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errEmptyResponseHeader
	}
	if strings.Count(value, ":") != 1 {
		return fmt.Errorf("%w: %s", errInvalidResponseHeader, value)
	}
	if len(strings.FieldsFunc(value, func(c rune) bool { return c == ':' })) != 2 {
		return fmt.Errorf("%w: %s", errInvalidResponseHeader, value)
	}
	*f = append(*f, value)
	return nil
}
