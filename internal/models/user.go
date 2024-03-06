package models

import (
	"time"

	"github.com/go-ozzo/ozzo-validation/v4"
)

type User struct {
	ID        string    `json:"id" db:"id"`
	Login     string    `json:"login" db:"login"`
	Password  string    `json:"password" db:"password"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

func (u *User) Validate() error {
	return validation.ValidateStruct(u, validation.Field(&u.Login), validation.Field(&u.Password, validation.Required))
}
