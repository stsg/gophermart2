package models

import (
	"encoding/json"
	"time"
)

type AccrualStatus string

const (
	AccrualStatusNew        AccrualStatus = "NEW"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
)

type Order struct {
	ID            string        `json:"number" db:"id"`
	UID           string        `json:"-" db:"uid"`
	Accrual       *float64      `json:"accrual,omitempty" db:"accrual"`
	AccrualStatus AccrualStatus `json:"status" db:"accrual_status"`
	UploadedAt    time.Time     `json:"uploaded_at" db:"uploaded_at"`
}

func (o *Order) Marshal() (res []byte, err error) {
	res, err = json.Marshal(o)
	return
}
