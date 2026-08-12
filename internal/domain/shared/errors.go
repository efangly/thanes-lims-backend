package shared

import "errors"

// Sentinel domain errors returned by application-layer use cases. The HTTP
// adapter's central error mapper translates these into response-envelope
// status codes/error codes so individual handlers never need to know about
// HTTP status codes.
var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource conflict")
	ErrValidation   = errors.New("validation failed")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)
