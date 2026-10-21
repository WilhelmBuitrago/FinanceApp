package repositories

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrConflict    = errors.New("conflict")
	ErrConstraint  = errors.New("constraint violation")
	ErrUnavailable = errors.New("service unavailable")
	ErrInternal    = errors.New("internal error")
)
