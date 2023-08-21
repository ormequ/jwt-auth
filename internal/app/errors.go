package app

import "errors"

var (
	ErrNotFound         = errors.New("id not found")
	ErrPermissionDenied = errors.New("permission denied")
)