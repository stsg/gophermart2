package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/stsg/gophermart2/internal/auth"
	"github.com/stsg/gophermart2/internal/helpers"
	"github.com/stsg/gophermart2/internal/models"
)

type CtxType string

const UserCtxName CtxType = "user"

func UserValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var user models.User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			helpers.HTTPError(w, err)
			return
		}

		if err := user.Validate(); err != nil {
			helpers.HTTPError(w, err)
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserCtxName, user)))
	})
}

func TokenValidation() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := strings.Split(r.Header.Get("Authorization"), "Bearer ")
			if len(token) != 2 {
				helpers.HTTPError(w, models.ErrInvalidBearerTokenFormat)
				return
			}

			uid, err := auth.GetUIDFromValidToken(token[1])
			if err != nil {
				helpers.HTTPError(w, models.ErrUserUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), UserCtxName, models.User{ID: uid})))
		})
	}
}

func GetUserFromCtx(ctx context.Context) (models.User, error) {
	if user, ok := ctx.Value(UserCtxName).(models.User); ok {
		return user, nil
	}
	return models.User{}, models.ErrUserUnauthorized
}
