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

type iOtpEmailSender interface {
	SendOtpEmail(ctx context.Context, to models.Email, otp string) error
}

func SendOtpEmail(r registry, es iOtpEmailSender) {
	r.Register("otp_email", func(ctx context.Context, m models.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		otp, ok := m["otp"]
		if !ok {
			return errors.New("no otp in message")
		}

		if err := es.SendOtpEmail(ctx, models.Email(to), otp); err != nil {
			return fmt.Errorf("error sending verification email: %w", err)
		}

		return nil
	})
}

type iWelcomeEmailSender interface {
	SendWelcomeEmail(ctx context.Context, to models.Email) error
}

func SendWelcomeEmail(r registry, es iWelcomeEmailSender) {
	r.Register("welcome_email", func(ctx context.Context, m models.Message) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		to, ok := m["email"]
		if !ok {
			return errors.New("no email address in message")
		}

		if err := es.SendWelcomeEmail(ctx, models.Email(to)); err != nil {
			return fmt.Errorf("error sending verification email: %w", err)
		}

		return nil
	})
}
