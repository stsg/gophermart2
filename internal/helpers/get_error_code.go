package helpers

import (
	"errors"
	"net/http"

	"github.com/stsg/gophermart2/internal/models"
)

func HTTPError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), GetStatusByError(err))
}

func GetStatusByError(err error) int {
	switch {
	case errorsAre(err, models.ErrInsufficientFunds):
		return http.StatusPaymentRequired
	case errorsAre(err, models.ErrUserAlreadyExists, models.ErrOrderBelongsAnotherUser):
		return http.StatusConflict
	case errorsAre(err, models.ErrUserUnauthorized, models.ErrInvalidLoginAttempt, models.ErrInvalidBearerTokenFormat):
		return http.StatusUnauthorized
	case errorsAre(err, models.ErrInvalidOrderNumber):
		return http.StatusUnprocessableEntity
	case errorsAre(err, models.ErrNoOrders, models.ErrNoWithdrawals):
		return http.StatusNoContent
	default:
		return http.StatusBadRequest
	}
}

func errorsAre(err error, targets ...error) bool {
	for _, target := range targets {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}
