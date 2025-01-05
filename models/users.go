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

	Token string `db:"token" json:"token"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
