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
	// ErrAccountSuspended is returned when a User with valid credentials is
	// blocked from authenticating because their Status is suspended (ADR
	// 0010). Distinct from ErrUnauthorized so the client can tell "wrong
	// password" from "account on hold".
	ErrAccountSuspended = errors.New("account suspended")
)
