package brevo

import (
	"time"
)

// Config holds Brevo provider configuration
type Config struct {
	APIKey  string        `envconfig:"BREVO_API_KEY"`
	BaseURL string        `envconfig:"BREVO_BASE_URL" default:"https://api.brevo.com/v3"`
	Timeout time.Duration `envconfig:"BREVO_TIMEOUT" default:"30s"`
}

// SendRequest represents the Brevo send email request
type SendRequest struct {
	Sender      Contact   `json:"sender"`
	To          []Contact `json:"to"`
	CC          []Contact `json:"cc,omitempty"`
	BCC         []Contact `json:"bcc,omitempty"`
	Subject     string    `json:"subject"`
	HTMLContent string    `json:"htmlContent,omitempty"`
	TextContent string    `json:"textContent,omitempty"`
}

// Contact represents an email contact
type Contact struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// SendResponse represents the Brevo send email response
type SendResponse struct {
	MessageID string `json:"messageId"`
}

// ErrorResponse represents Brevo error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}
