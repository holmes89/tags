package internal

import "errors"

var(
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
	ErrInvalid = errors.New("invalid entity")
)
