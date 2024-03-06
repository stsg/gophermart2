package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/middlewares"
	gophermart "github.com/stsg/gophermart2/internal/services/gophermart"
)

func GetOrders(g *gophermart.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middlewares.GetUserFromCtx(r.Context())
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		orders, err := g.Storage.GetOrdersByUID(r.Context(), user.ID)
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		res, err := json.Marshal(orders)
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(res)
	}
}
