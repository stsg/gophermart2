package gophermart2

import (
	"context"
	"time"

	"github.com/stsg/gophermart2/internal/accrual"
	"github.com/stsg/gophermart2/internal/storages"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const jobInterval = 10 * time.Second

type Gophermart struct {
	AccrualClient accrual.Client
	Storage       storages.Storager
}

func New(ctx context.Context, opts ...Option) *Gophermart {
	g := &Gophermart{
		AccrualClient: accrual.New(),
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.Storage == nil {
		WithDefaultStorage(ctx)(g)
	}

	g.updateOrdersBackground(ctx)
	return g
}

func (g *Gophermart) updateOrdersBackground(ctx context.Context) {
	ticker := time.NewTicker(jobInterval)

	go func() {
		defer func() {
			if p := recover(); p != nil {
				zap.L().Warn("recovered from panic", zap.Any("panic", p))
			}
		}()

		for {
			select {
			case <-ticker.C:
				g.updateOrders(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (g *Gophermart) updateOrders(ctx context.Context) {
	orders, err := g.Storage.GetUnfinishedOrders(ctx)
	if err != nil {
		zap.L().Warn("update orders: get unfinished orders", zap.Error(err))
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		order, err = g.AccrualClient.GetOrderInfo(order)
		if err != nil {
			zap.L().Warn("update orders: get order info", zap.Error(err))
			return
		}

		if err = g.Storage.Transaction(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
			if err := g.Storage.UpdateOrder(ctx, order, tx); err != nil {
				return err
			}
			if order.Accrual == nil {
				return nil
			}
			if err := g.Storage.AddOrIncrBalance(ctx, order.UID, order.Accrual, tx); err != nil {
				return err
			}
			return nil
		}); err != nil {
			zap.L().Warn("update orders: exec transaction error", zap.Error(err))
		}
	}
}
