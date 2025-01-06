package storage

import (
	"context"
	"database/sql"
	"time"

	"cyberix.fr/frcc/models"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(first_name, last_name, email, quality, phone, organization, token)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, first_name, last_name, email, quality, phone, organization, created_at, updated_at, token, current_otp, current_otp_validity_time
`

type CreateUserParams struct {
	FirstName    string `db:"first_name" json:"first_name"`
	LastName     string `db:"last_name" json:"last_name"`
	Email        string `db:"email" json:"email"`
	Quality      string `db:"quality" json:"quality"`
	Phone        string `db:"phone" json:"phone"`
	Organization string `db:"organization" json:"organization"`
	Token        string `db:"token" json:"token"`
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (*models.User, error) {
	row := q.db.QueryRowContext(ctx, createUser,
		arg.FirstName,
		arg.LastName,
		arg.Email,
		arg.Quality,
		arg.Phone,
		arg.Organization,
		arg.Token,
	)
	var i models.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Quality,
		&i.Phone,
		&i.Organization,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Token,
		&i.CurrentOtp,
		&i.CurrentOtpValidityTime,
	)
	return &i, err
}

const getUserByEmailOrPhone = `-- name: GetUserByEmailOrPhone :one
SELECT id, first_name, last_name, email, quality, phone, organization, created_at, updated_at, token, current_otp, current_otp_validity_time
FROM users
WHERE email = $1 OR phone = $2
`

type GetUserByEmailOrPhoneParams struct {
	Email string `db:"email" json:"email"`
	Phone string `db:"phone" json:"phone"`
}

func (q *Queries) GetUserByEmailOrPhone(ctx context.Context, arg GetUserByEmailOrPhoneParams) (*models.User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmailOrPhone, arg.Email, arg.Phone)
	var i models.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.Email,
		&i.Quality,
		&i.Phone,
		&i.Organization,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Token,
		&i.CurrentOtp,
		&i.CurrentOtpValidityTime,
	)

	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

const setCurrentOtp = `-- name: SetCurrentOtp :exec
UPDATE users
SET 
  current_otp = $1,
  current_otp_validity_time = $2
WHERE
  email = $3
`

type SetCurrentOtpParams struct {
	CurrentOtp             string    `db:"current_otp" json:"current_otp"`
	CurrentOtpValidityTime time.Time `db:"current_otp_validity_time" json:"current_otp_validity_time"`
	Email                  string    `db:"email" json:"email"`
}

func (q *Queries) SetCurrentOtp(ctx context.Context, arg SetCurrentOtpParams) error {
	_, err := q.db.ExecContext(ctx, setCurrentOtp, arg.CurrentOtp, arg.CurrentOtpValidityTime, arg.Email)
	return err
}
