package handlers

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"fmt"

	"cyberix.fr/frcc/models"
	"cyberix.fr/frcc/storage"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
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
			http.Error(w, fmt.Errorf("error checking if user already exists: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// if user exists, stop and return error
		if user != nil {
			http.Error(w, fmt.Errorf("error user with email/phone already exists").Error(), http.StatusBadRequest)
			return
		}

		token, err := createSecret()
		if err != nil {
			http.Error(w, fmt.Errorf("error creating token: %v", err).Error(), http.StatusBadRequest)
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

			ConfirmationToken: token,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error creating the new users: %w", err).Error(), http.StatusBadRequest)
			return
		}

		// send email
		err = q.Send(ctx, models.Message{
			"job":   "verification_email",
			"email": input.Email,
			"token": token,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error adding mail into queue: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// return ok
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(true); err != nil {
			http.Error(w, "error encoding the result", http.StatusBadRequest)
			return
		}
	})
}

type iRegisterConfirm interface {
	ConfirmRegister(ctx context.Context, token string) (*models.User, error)
}

func (appHandler *AppHandler) RegisterConfirm(mux chi.Router, db iRegisterConfirm, q iQueue) {
	mux.Post("/register/confirm", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		token := r.FormValue("token")

		user, err := db.ConfirmRegister(ctx, token)
		if err != nil {
			http.Error(w, "error saving email address confirmation", http.StatusBadRequest)
			return
		}
		if user == nil {
			http.Error(w, "error not user associated to this token", http.StatusBadRequest)
			return
		}

		err = q.Send(
			ctx,
			models.Message{
				"job":   "welcome_email",
				"email": user.Email,
			},
		)
		if err != nil {
			http.Error(w, "error saving email address confirmation", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(true); err != nil {
			http.Error(w, "error encoding the result", http.StatusBadRequest)
			return
		}
	})
}

type iLoginer interface {
	GetUserByEmailOrPhone(ctx context.Context, arg storage.GetUserByEmailOrPhoneParams) (*models.User, error)
	SetCurrentOtp(ctx context.Context, arg storage.SetCurrentOtpParams) error
}

type LoginRequest struct {
	Email string `json:"email,omitempty"`
}

func (appHandler *AppHandler) Login(mux chi.Router, db iLoginer, q iQueue) {
	mux.Post("/login", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var input LoginRequest
		httpStatus, err := appHandler.ParsingRequestBody(w, r, &input)
		if err != nil {
			http.Error(w, err.Error(), httpStatus)
			return
		}

		// check if user already exists
		user, err := db.GetUserByEmailOrPhone(ctx, storage.GetUserByEmailOrPhoneParams{
			Email: input.Email,
			Phone: input.Email,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error checking if user already exists: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// if user exists, stop and return error
		if user == nil {
			http.Error(w, fmt.Errorf("error user with email/phone does not exists").Error(), http.StatusBadRequest)
			return
		}

		if !user.ConfirmedAccount {
			http.Error(w, fmt.Errorf("sorry kindly confirmed your account").Error(), http.StatusBadRequest)
			return
		}

		otp := createOtp()

		duration := 2*time.Minute + 30*time.Second
		otpValidity := time.Now().UTC().Add(duration)

		err = db.SetCurrentOtp(ctx, storage.SetCurrentOtpParams{
			CurrentOtp:             otp,
			CurrentOtpValidityTime: otpValidity,
			Email:                  input.Email,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error updating current otp: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// send email
		err = q.Send(ctx, models.Message{
			"job":   "otp_email",
			"email": input.Email,
			"otp":   otp,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error adding mail into queue: %v", err).Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(true); err != nil {
			http.Error(w, "error encoding the result", http.StatusBadRequest)
			return
		}
	})
}

type iOtper interface {
	GetUserByEmailOrPhone(ctx context.Context, arg storage.GetUserByEmailOrPhoneParams) (*models.User, error)
}

type OtpRequest struct {
	Email string `json:"email,omitempty"`
	Otp   string `json:"otp,omitempty"`
}

func (appHandler *AppHandler) Otp(mux chi.Router, db iOtper) {
	mux.Post("/otp", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var input OtpRequest
		httpStatus, err := appHandler.ParsingRequestBody(w, r, &input)
		if err != nil {
			http.Error(w, err.Error(), httpStatus)
			return
		}

		// check if user already exists
		user, err := db.GetUserByEmailOrPhone(ctx, storage.GetUserByEmailOrPhoneParams{
			Email: input.Email,
			Phone: input.Email,
		})
		if err != nil {
			http.Error(w, fmt.Errorf("error checking if user already exists: %v", err).Error(), http.StatusBadRequest)
			return
		}

		// if user exists, stop and return error
		if user == nil {
			http.Error(w, fmt.Errorf("error user with email/phone does not exists").Error(), http.StatusBadRequest)
			return
		}

		log.Println("Now vs Validity: ", time.Now().Before(*user.CurrentOtpValidityTime), time.Now().UTC(), user.CurrentOtpValidityTime)
		if !time.Now().UTC().Before(*user.CurrentOtpValidityTime) {
			http.Error(w, fmt.Errorf("error otp has expired").Error(), http.StatusBadRequest)
			return
		}

		if input.Otp != *user.CurrentOtp {
			http.Error(w, fmt.Errorf("error wrong otp").Error(), http.StatusBadRequest)
			return
		}

		jwtToken, err := generateJWT(input.Email, fmt.Sprintf("%s %s", user.FirstName, user.LastName)) // Replace with actual user data
		if err != nil {
			http.Error(w, fmt.Errorf("error generating token").Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(
			w,
			&http.Cookie{
				Name:     "jwt",
				Value:    jwtToken,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				MaxAge:   3600 * 24,
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			},
		)

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

func createOtp() string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	randomNumber := r.Intn(900000) + 100000

	return fmt.Sprintf("%06d", randomNumber)
}

// Define a struct for the JWT claims (payload).
type Claims struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	jwt.RegisteredClaims
}

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func generateJWT(email, name string) (string, error) {
	claims := &Claims{
		Name:  name,
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cyberix-frcc-api", // Replace with your service name
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a new token object, specifying signing method and the claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign and get the complete encoded token as a string using the secret.
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
