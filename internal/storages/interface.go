package storages

import (
	"context"

	"github.com/stsg/gophermart2/internal/models"
	"github.com/jmoiron/sqlx"
)

// Storager is the common interface implemented by all storages.
type Storager interface {
	StorageReader
	StorageWriter
	Transaction(ctx context.Context, f func(ctx context.Context, tx *sqlx.Tx) error) (err error)
}

type StorageReader interface {
	GetUserByLogin(ctx context.Context, user models.User) (models.User, error)

	GetOrdersByUID(ctx context.Context, UID string) ([]models.Order, error)
	GetUnfinishedOrders(ctx context.Context) ([]models.Order, error)

	GetBalanceByUID(ctx context.Context, UID string) (models.Balance, error)
	GetCurrentBalanceByUID(ctx context.Context, UID string, tx *sqlx.Tx) (float64, error)

	GetWithdrawalsByUID(ctx context.Context, UID string) ([]models.Withdrawal, error)
}

type StorageWriter interface {
	AddUser(ctx context.Context, user models.User) error

	AddOrder(ctx context.Context, OrderID models.Order, tx *sqlx.Tx) error
	UpdateOrder(ctx context.Context, order models.Order, tx *sqlx.Tx) error

	AddOrIncrBalance(ctx context.Context, UID string, ptrBalance *float64, tx *sqlx.Tx) error
	IncrBalanceWithdrawnByUID(ctx context.Context, UID string, value float64, tx *sqlx.Tx) error

	AddWithdrawal(ctx context.Context, wth models.Withdrawal, tx *sqlx.Tx) error
}
