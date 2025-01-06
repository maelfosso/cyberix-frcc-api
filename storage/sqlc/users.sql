-- name: CreateUser :one
INSERT INTO users(first_name, last_name, email, quality, phone, organization, token)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByEmailOrPhone :one
SELECT *
FROM users
WHERE email = $1 OR phone = $2;

-- name: SetCurrentOtp :exec
UPDATE users
SET 
  current_otp = $1,
  current_otp_validity_time = $2
WHERE
  email = $3
;
