package handlers

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/stsg/gophermart2/internal/auth"
	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/middlewares"
	"github.com/stsg/gophermart2/internal/services/gophermart"
	"golang.org/x/crypto/bcrypt"
)

func Register(g *gophermart2.Gophermart) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := middlewares.GetUserFromCtx(r.Context())
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			helpers.HTTPError(w, err)
			return
		}

		user.ID = uuid.NewString()
		user.Password = string(hashedPass)

		if err = g.Storage.AddUser(r.Context(), user); err != nil {
			helpers.HTTPError(w, err)
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
