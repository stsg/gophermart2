package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/luhn"
	"github.com/stsg/gophermart2/internal/middlewares"
	"github.com/stsg/gophermart2/internal/models"
	"github.com/stsg/gophermart2/internal/services/gophermart2"
	"github.com/jmoiron/sqlx"
)

func Withdraw(g *gophermart2.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middlewares.GetUserFromCtx(r.Context())
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		withdrawal := models.Withdrawal{UID: user.ID}
		if err = json.NewDecoder(r.Body).Decode(&withdrawal); err != nil {
			helpers.HTTPError(w, err)
			return
		}

		if !luhn.Valid(withdrawal.OrderID) {
			helpers.HTTPError(w, models.ErrInvalidOrderNumber)
			return
		}

		err = g.Storage.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) (err error) {
			balance, err := g.Storage.GetCurrentBalanceByUID(ctx, user.ID, tx)
			if err != nil {
				return
			}

			if balance < withdrawal.Amount {
				return models.ErrInsufficientFunds
			}

			if err = g.Storage.AddWithdrawal(ctx, withdrawal, tx); err != nil {
				return
			}

			return g.Storage.IncrBalanceWithdrawnByUID(ctx, user.ID, withdrawal.Amount, tx)
		})
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
