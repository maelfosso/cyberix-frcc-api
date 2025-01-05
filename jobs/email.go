package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cyberix.fr/frcc/models"
)

type iVerificationEmailSender interface {
	SendVerificationEmail(ctx context.Context, to models.Email, token string) error
}

func SendVerificationEmail(r registry, es iVerificationEmailSender) {
	r.Register("verification_email", func(ctx context.Context, m models.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		token, ok := m["token"]
		if !ok {
			return errors.New("no token in message")
		}

		if err := es.SendVerificationEmail(ctx, models.Email(to), token); err != nil {
			return fmt.Errorf("error sending verification email: %w", err)
		}

		return nil
	})
}
