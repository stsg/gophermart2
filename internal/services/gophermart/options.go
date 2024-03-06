package gophermart2

import (
	"context"

	"github.com/stsg/gophermart2/internal/storages/postgres"
)

type Option func(g *Gophermart)

func WithPostgreStorage(ctx context.Context) Option {
	return func(g *Gophermart) {
		g.Storage = postgres.New(ctx)
	}
}

func WithDefaultStorage(ctx context.Context) Option {
	return WithPostgreStorage(ctx)
}
