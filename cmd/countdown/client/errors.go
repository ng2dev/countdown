package client

import (
	"github.com/iov-one/weave/errors"
)

var (
	// ErrNoMatch is returned when two compared values does not match
	ErrNoMatch = errors.Register(121, "no match")
	// ErrInvalid is returned when a value is invalid
	ErrInvalid = errors.Register(122, "invalid")
	// ErrPermission is returned when an action is not permitted
	ErrPermission = errors.Register(123, "not permitted")
)
