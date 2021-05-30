package datastore

import (
	"errors"
)

var (
	// ErrNotFound for records that are not exist.
	ErrNotFound = errors.New("record does not exist")

	// ErrSegmentClosed for segments that are closed.
	ErrSegmentClosed = errors.New("segment closed")
)
