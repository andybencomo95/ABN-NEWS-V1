package models

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrDuplicateHash = errors.New("duplicate article hash")
	ErrSourceFailing = errors.New("source is in failing state")
)
