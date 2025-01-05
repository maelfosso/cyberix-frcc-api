package storage

import (
	"context"
	"database/sql"

	"cyberix.fr/frcc/models"
)

const createUser = `-- name: CreateUser :one
INSERT INTO users(first_name, last_name, email, quality, phone, organization, token)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, first_name, last_name, email, quality, phone, organization, created_at, updated_at, token
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
	)
	return &i, err
}

const getUserByEmailOrPhone = `-- name: GetUserByEmailOrPhone :one
SELECT id, first_name, last_name, email, quality, phone, organization, created_at, updated_at, token
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
	)

	if err != nil && err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}
