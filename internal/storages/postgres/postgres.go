package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/stsg/gophermart2/internal/config"
	"github.com/stsg/gophermart2/internal/models"
	"github.com/stsg/gophermart2/internal/services/shutdowner"
	"github.com/stsg/gophermart2/internal/storages"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

const queryTimeout = time.Second

var (
	//go:embed sql/migrations/*.sql
	migrationsFs     embed.FS
	migrationsFsName = "sql/migrations"

	//go:embed sql/queries
	queriesFs     embed.FS
	queriesFsName = "sql/queries"
)

func getQueryFromFile(filename string) (string, error) {
	query, err := queriesFs.ReadFile(fmt.Sprintf("%s/%s", queriesFsName, filename))
	if err != nil {
		return "", err
	}
	return string(query), nil
}

type Storage struct {
	db      *sqlx.DB
	queries struct {
		insertOrUpdateBalancesByUID string
		updateBalanceWithdrawnByUID string
		selectBalanceByUID          string

		insertOrder            string
		updateOrders           string
		selectOrderByID        string
		selectOrdersByUID      string
		selectOrdersByStatuses string

		insertUser        string
		selectUserByLogin string

		insertWithdrawals      string
		selectWithdrawalsByUID string
	}
}

var _ storages.Storager = (*Storage)(nil)

func New(ctx context.Context) storages.Storager {
	db, err := sqlx.Connect("pgx", config.Get().DatabaseURI)
	if err != nil {
		zap.L().Fatal("storage constructor: connect to postgres db", zap.Error(err))
	}

	s := &Storage{db: db}
	if err = s.migrate(ctx); err != nil {
		zap.L().Fatal("storage constructor: applies migrations", zap.Error(err))
	}
	if err = s.setQueries(ctx); err != nil {
		zap.L().Fatal("storage constructor: set queries", zap.Error(err))
	}
	s.addToShutdowner(ctx)
	return s
}

func (s *Storage) AddUser(ctx context.Context, user models.User) (err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	res, err := s.db.NamedExecContext(ctx, s.queries.insertUser, user)
	if err != nil {
		return
	}

	numRowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}

	if numRowsAffected == 0 {
		return models.ErrUserAlreadyExists
	}
	return
}

func (s *Storage) GetUserByLogin(ctx context.Context, user models.User) (res models.User, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err = s.db.GetContext(ctx, &res, s.queries.selectUserByLogin, user.Login); err != nil {
		return
	}
	return
}

func (s *Storage) AddOrder(ctx context.Context, newOrder models.Order, tx *sqlx.Tx) (err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	var order models.Order
	if err = tx.GetContext(ctx, &order, s.queries.selectOrderByID, newOrder.ID); err == nil {
		if order.UID == newOrder.UID {
			return models.ErrOrderAlreadyExists
		}
		return models.ErrOrderBelongsAnotherUser
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return
	}

	ctx, cancel = context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	_, err = s.db.NamedExecContext(ctx, s.queries.insertOrder, &newOrder)
	return
}

func (s *Storage) UpdateOrder(ctx context.Context, order models.Order, tx *sqlx.Tx) (err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	_, err = tx.NamedExecContext(ctx, s.queries.updateOrders, &order)
	return
}

func (s *Storage) GetOrdersByUID(ctx context.Context, UID string) (orders []models.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err = s.db.SelectContext(ctx, &orders, s.queries.selectOrdersByUID, UID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoOrders
		}
		return
	}

	if len(orders) == 0 {
		return nil, models.ErrNoOrders
	}
	return
}

func (s *Storage) GetBalanceByUID(ctx context.Context, UID string) (balance models.Balance, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	err = s.db.GetContext(ctx, &balance, s.queries.selectBalanceByUID, UID)
	return
}

func (s *Storage) GetCurrentBalanceByUID(ctx context.Context, UID string, tx *sqlx.Tx) (float64, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	var balance models.Balance
	if err := tx.GetContext(ctx, &balance, s.queries.selectBalanceByUID, UID); err != nil {
		return -1, err
	}
	return balance.Current, nil
}

func (s *Storage) AddWithdrawal(ctx context.Context, withdrawal models.Withdrawal, tx *sqlx.Tx) (err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	_, err = tx.NamedExecContext(ctx, s.queries.insertWithdrawals, &withdrawal)
	return
}

func (s *Storage) GetWithdrawalsByUID(ctx context.Context, UID string) (withdrawals []models.Withdrawal, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err = s.db.SelectContext(ctx, &withdrawals, s.queries.selectWithdrawalsByUID, UID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoWithdrawals
		}
		return
	}

	if len(withdrawals) == 0 {
		return nil, models.ErrNoWithdrawals
	}
	return
}

func (s *Storage) AddOrIncrBalance(ctx context.Context, UID string, ptrBalance *float64, tx *sqlx.Tx) (err error) {
	var balance float64
	if ptrBalance != nil {
		balance = *ptrBalance
	}
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	_, err = tx.ExecContext(ctx, s.queries.insertOrUpdateBalancesByUID, balance, UID)
	return
}
func (s *Storage) IncrBalanceWithdrawnByUID(ctx context.Context, UID string, value float64, tx *sqlx.Tx) (err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	_, err = tx.ExecContext(ctx, s.queries.updateBalanceWithdrawnByUID, value, UID)
	return
}

func (s *Storage) GetUnfinishedOrders(ctx context.Context) (orders []models.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	err = s.db.SelectContext(ctx, &orders, s.queries.selectOrdersByStatuses, models.AccrualStatusNew, models.AccrualStatusProcessing)
	return
}

func (s *Storage) Transaction(ctx context.Context, f func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*queryTimeout)
	defer cancel()
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		zap.L().Warn("transaction: begin", zap.Error(err))
		return
	}

	defer func() {
		p := recover()
		switch {
		case p != nil:
			_ = tx.Rollback()
			zap.L().Fatal("transaction: panic", zap.Any("panic", p))
		case err != nil:
			_ = tx.Rollback()
			zap.L().Warn("transaction: error", zap.Error(err))
		default:
			if err = tx.Commit(); err != nil {
				zap.L().Fatal("transaction: commit", zap.Error(err))
			}
		}
	}()
	return f(ctx, tx)
}

func (s *Storage) setQueries(_ context.Context) error {
	files, err := queriesFs.ReadDir(queriesFsName)
	if err != nil {
		return err
	}

	for _, file := range files {
		query, err := getQueryFromFile(file.Name())
		if err != nil {
			return err
		}

		switch file.Name() {
		case "insert_or_update_balances_by_uid.sql":
			s.queries.insertOrUpdateBalancesByUID = query
		case "update_balance_withdrawn_by_uid.sql":
			s.queries.updateBalanceWithdrawnByUID = query
		case "select_balance_by_uid.sql":
			s.queries.selectBalanceByUID = query

		case "insert_order.sql":
			s.queries.insertOrder = query
		case "update_orders.sql":
			s.queries.updateOrders = query
		case "select_order_by_id.sql":
			s.queries.selectOrderByID = query
		case "select_orders_by_uid.sql":
			s.queries.selectOrdersByUID = query
		case "select_orders_by_statuses.sql":
			s.queries.selectOrdersByStatuses = query

		case "insert_user.sql":
			s.queries.insertUser = query
		case "select_user_by_login.sql":
			s.queries.selectUserByLogin = query

		case "insert_withdrawals.sql":
			s.queries.insertWithdrawals = query
		case "select_withdrawals_by_uid.sql":
			s.queries.selectWithdrawalsByUID = query
		}
	}
	return err
}

func (s *Storage) migrate(_ context.Context) (err error) {
	goose.SetBaseFS(migrationsFs)

	if err = goose.SetDialect("postgres"); err != nil {
		return
	}

	err = goose.Up(s.db.DB, migrationsFsName)
	return
}

func (s *Storage) addToShutdowner(_ context.Context) {
	shutdowner.Get().AddCloser(func(_ context.Context) error {
		return s.db.Close()
	})
}
