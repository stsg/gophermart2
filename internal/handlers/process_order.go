package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/luhn"
	"github.com/stsg/gophermart2/internal/middlewares"
	"github.com/stsg/gophermart2/internal/models"
	gophermart "github.com/stsg/gophermart2/internal/services/gophermart"
)

func ProcessOrder(g *gophermart.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middlewares.GetUserFromCtx(r.Context())
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		var number int
		if err = json.NewDecoder(r.Body).Decode(&number); err != nil {
			helpers.HTTPError(w, err)
			return
		}

		order := models.Order{
			ID:            strconv.Itoa(number),
			UID:           user.ID,
			AccrualStatus: models.AccrualStatusNew,
		}
		if !luhn.Valid(order.ID) {
			helpers.HTTPError(w, models.ErrInvalidOrderNumber)
			return
		}

		err = g.Storage.Transaction(r.Context(), func(ctx context.Context, tx *sqlx.Tx) (err error) {
			if err = g.Storage.AddOrder(r.Context(), order, tx); err != nil {
				return
			}
			return g.Storage.AddOrIncrBalance(ctx, user.ID, order.Accrual, tx)
		})
		if err != nil {
			if errors.Is(err, models.ErrOrderAlreadyExists) {
				w.WriteHeader(http.StatusOK)
				return
			}
			helpers.HTTPError(w, err)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}
