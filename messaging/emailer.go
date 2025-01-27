package messaging

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"cyberix.fr/frcc/models"
	"go.uber.org/zap"
)

const (
	marketingMessageStream     = "broadcast"
	transactionalMessageStream = "outbound"
)

type nameAndEmail = string

//go:embed emails
var emails embed.FS

type Emailer struct {
	baseURL           string
	client            *http.Client
	log               *zap.Logger
	marketingFrom     nameAndEmail
	token             string
	transactionalFrom nameAndEmail
}

type NewEmailerOptions struct {
	BaseURL                   string
	Log                       *zap.Logger
	MarketingEmailAddress     string
	MarketingEmailName        string
	Token                     string
	TransactionalEmailAddress string
	TransactionalEmailName    string
}

func NewEmailer(opts NewEmailerOptions) *Emailer {
	return &Emailer{
		baseURL: opts.BaseURL,
		client:  &http.Client{Timeout: 3 * time.Second},
		log:     opts.Log,
		marketingFrom: createNameAndEmail(
			opts.MarketingEmailName,
			opts.MarketingEmailAddress,
		),
		token:             opts.Token,
		transactionalFrom: createNameAndEmail(opts.TransactionalEmailName, opts.TransactionalEmailAddress),
	}
}

func (e *Emailer) SendVerificationEmail(ctx context.Context, to models.Email, token string) error {
	keywords := map[string]string{
		"base_url":   e.baseURL,
		"action_url": e.baseURL + "/register/confirm/" + token,
	}

	return e.send(ctx, requestBody{
		MessageStream: transactionalMessageStream,
		From:          e.transactionalFrom,
		To:            to.String(),
		Subject:       "Verify your registration to FRCC",
		HtmlBody:      getEmail("verification_email.html", keywords),
		TextBody:      getEmail("verification_email.txt", keywords),
	})
}

func (e *Emailer) SendOtpEmail(ctx context.Context, to models.Email, name, otp string) error {
	keywords := map[string]string{
		"otp":     otp,
		"email":   to.String(),
		"name":    name,
		"website": os.Getenv("WEBSITE"),
	}

	return e.send(ctx, requestBody{
		MessageStream: transactionalMessageStream,
		From:          e.transactionalFrom,
		To:            to.String(),
		Subject:       "Votre code OTP pour l'enregistrement au Forum Régional sur la Sécurité",
		HtmlBody:      getEmail("otp_email.html", keywords),
		TextBody:      getEmail("otp_email.txt", keywords),
	})
}

type requestBody struct {
	MessageStream string
	From          nameAndEmail
	To            nameAndEmail
	Subject       string
	HtmlBody      string
	TextBody      string
}

func (e *Emailer) SendWelcomeEmail(ctx context.Context, to models.Email, name string) error {
	keywords := map[string]string{
		"email":   to.String(),
		"name":    name,
		"website": os.Getenv("WEBSITE"),
	}

	return e.send(ctx, requestBody{
		MessageStream: transactionalMessageStream,
		From:          e.transactionalFrom,
		To:            to.String(),
		Subject:       "Merci pour votre enregistrement au Forum Régional sur la Sécurité",
		HtmlBody:      getEmail("confirmation_email.html", keywords),
		TextBody:      getEmail("confirmation_email.txt", keywords),
	})
}

func (e *Emailer) send(ctx context.Context, body requestBody) error {
	bodyAsBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling request body to json: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.postmarkapp.com/email",
		bytes.NewReader(bodyAsBytes),
	)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Postmark-Server-Token", e.token)

	response, err := e.client.Do(request)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	bodyAsBytes, err = io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	if response.StatusCode > 299 {
		e.log.Info(
			"Error sending email",
			zap.Int("status", response.StatusCode),
			zap.String("response", string(bodyAsBytes)),
		)
		return fmt.Errorf("error sending email, got status %v", response.StatusCode)
	}

	return nil
}

func createNameAndEmail(name, email string) nameAndEmail {
	return fmt.Sprintf("%v <%v>", name, email)
}

func getEmail(path string, keywords map[string]string) string {
	email, err := emails.ReadFile("emails/" + path)
	if err != nil {
		panic(err)
	}

	emailString := string(email)
	for keyword, replacement := range keywords {
		emailString = strings.ReplaceAll(
			emailString,
			"{{"+keyword+"}}",
			replacement,
		)
	}

	return emailString
}
