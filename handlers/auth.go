package handlers

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"

	"fmt"

	"cyberix.fr/frcc/models"
	"cyberix.fr/frcc/storage"
	"github.com/go-chi/chi/v5"
)

type iRegister interface {
	GetUserByEmailOrPhone(ctx context.Context, arg storage.GetUserByEmailOrPhoneParams) (*models.User, error)
	CreateUser(ctx context.Context, arg storage.CreateUserParams) (*models.User, error)
}

type iQueue interface {
	Send(ctx context.Context, arg models.Message) error
}

type RegisterRequest struct {
	FirstName    string `json:"first_name,omitempty"`
	LastName     string `json:"last_name,omitempty"`
	Email        string `json:"email,omitempty"`
	Quality      string `json:"quality,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Organization string `json:"organization,omitempty"`
}

type RegisterResponse struct {
}

func (appHandler *AppHandler) Register(mux chi.Router, db iRegister, q iQueue) {
	mux.Post("/register", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var input RegisterRequest
		httpStatus, err := appHandler.ParsingRequestBody(w, r, &input)
		if err != nil {
			http.Error(w, err.Error(), httpStatus)
			return
		}

		// check if user already exists
		user, err := db.GetUserByEmailOrPhone(ctx, storage.GetUserByEmailOrPhoneParams{
			Email: input.Email,
			Phone: input.Phone,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error when checking if user already exists: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// if user exists, stop and return error
		if user != nil {
			http.Error(w, fmt.Errorf("error user with email/phone already exists").Error(), http.StatusBadRequest)
			return
		}

		token, err := createSecret()
		if err != nil {
			http.Error(w, fmt.Errorf("error when creating token: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// continue the registration
		_, err = db.CreateUser(ctx, storage.CreateUserParams{
			FirstName:    input.FirstName,
			LastName:     input.LastName,
			Email:        input.Email,
			Quality:      input.Quality,
			Phone:        input.Phone,
			Organization: input.Organization,

			Token: token,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error when creating the new users: %w", err).Error(), http.StatusBadRequest)
			return
		}

		// send email
		// err = q.Send(ctx, models.Message{
		// 	"job":   "verification_email",
		// 	"email": input.Email,
		// 	"token": token,
		// })
		// if err != nil {
		// 	http.Error(w, fmt.Errorf("error when adding mail into queue: %v", err).Error(), http.StatusBadRequest)
		// 	return
		// }

		// return ok
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(true); err != nil {
			http.Error(w, "error encoding the result", http.StatusBadRequest)
			return
		}
	})
}

func createSecret() (string, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", secret), nil
}

type iLogin interface {
}

func (appHandler *AppHandler) Login(mux chi.Router) {
	mux.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		// ctx := r.Context()
	})
}
