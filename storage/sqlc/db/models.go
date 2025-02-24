// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"database/sql"
)

type User struct {
	ID                     int32          `db:"id" json:"id"`
	FirstName              string         `db:"first_name" json:"first_name"`
	LastName               string         `db:"last_name" json:"last_name"`
	Email                  string         `db:"email" json:"email"`
	Quality                string         `db:"quality" json:"quality"`
	Phone                  string         `db:"phone" json:"phone"`
	Organization           string         `db:"organization" json:"organization"`
	CreatedAt              interface{}    `db:"created_at" json:"created_at"`
	UpdatedAt              interface{}    `db:"updated_at" json:"updated_at"`
	ConfirmationToken      string         `db:"confirmation_token" json:"confirmation_token"`
	CurrentOtp             sql.NullString `db:"current_otp" json:"current_otp"`
	CurrentOtpValidityTime interface{}    `db:"current_otp_validity_time" json:"current_otp_validity_time"`
	ConfirmedAccount       sql.NullBool   `db:"confirmed_account" json:"confirmed_account"`
}
