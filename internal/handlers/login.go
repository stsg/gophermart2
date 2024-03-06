package handlers

import (
	"fmt"
	"net/http"

	"github.com/stsg/gophermart2/internal/auth"
	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/middlewares"
	"github.com/stsg/gophermart2/internal/models"
	"github.com/stsg/gophermart2/internal/services/gophermart2"
)

func Login(g *gophermart2.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middlewares.GetUserFromCtx(r.Context())
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		dbUser, err := g.Storage.GetUserByLogin(r.Context(), user)
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}
		if !auth.Authenticate(dbUser, user) {
			helpers.HTTPError(w, models.ErrInvalidLoginAttempt)
			return
		}

		token, err := auth.GenerateToken(user)
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}
		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w.WriteHeader(http.StatusOK)
	}
}
