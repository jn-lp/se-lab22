package datastore

import (
	"context"
)

type DSOptions struct {
	Context context.Context
}

type DSOption func(so *DSOptions)

// WithContext sets context of the datastore.
func WithContext(c context.Context) DSOption {
	return func(so *DSOptions) {
		so.Context = c
	}
}
