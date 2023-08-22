package app

import "errors"

var (
	ErrNotFound         = errors.New("id not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrExpired          = errors.New("token has been expired")
	ErrInvalidUserID    = errors.New("invalid user ID")
)
