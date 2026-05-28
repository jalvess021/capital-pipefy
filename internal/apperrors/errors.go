package apperrors

import "errors"

var (
	ErrConflict = errors.New("conflict")
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("internal")
)
