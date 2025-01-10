package models

import (
	"time"
)

type User struct {
	ID           int32  `db:"id" json:"id"`
	FirstName    string `db:"first_name" json:"first_name"`
	LastName     string `db:"last_name" json:"last_name"`
	Email        string `db:"email" json:"email"`
	Quality      string `db:"quality" json:"quality"`
	Phone        string `db:"phone" json:"phone"`
	Organization string `db:"organization" json:"organization"`

	ConfirmationToken string `db:"confirmation_token" json:"confirmation_token"`
	ConfirmedAccount  bool   `db:"confirmed_account" json:"confirmed_account"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	CurrentOtp             *string    `db:"current_otp" json:"current_otp"`
	CurrentOtpValidityTime *time.Time `db:"current_otp_validity_time" json:"current_otp_validity_time"`
}
