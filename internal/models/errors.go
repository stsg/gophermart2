package models

import (
	"errors"
)

var (
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrNoOrders                = errors.New("you have no orders")
	ErrNoWithdrawals           = errors.New("you have no withdrawals")
	ErrOrderAlreadyExists      = errors.New("this order already exists")
	ErrOrderBelongsAnotherUser = errors.New("this order belongs to another user")
	ErrUserAlreadyExists       = errors.New("this user already exists")

	ErrUserUnauthorized    = errors.New("user unauthorized")
	ErrInvalidLoginAttempt = errors.New("invalid username or password")

	ErrInvalidOrderNumber = errors.New("invalid order number")

	ErrInvalidBearerToken       = errors.New("invalid bearer token")
	ErrInvalidBearerTokenFormat = errors.New("bearer token not in proper format")
)
