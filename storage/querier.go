package storage

import (
	"context"

	"cyberix.fr/frcc/models"
)

type Querier interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (*models.User, error)
	ConfirmRegister(ctx context.Context, confirmationToken string) (*models.User, error)
	GetUserByEmailOrPhone(ctx context.Context, arg GetUserByEmailOrPhoneParams) (*models.User, error)
	SetCurrentOtp(ctx context.Context, arg SetCurrentOtpParams) error
}

type QuerierTx interface {
}

var _ Querier = (*Queries)(nil)
