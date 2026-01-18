package brevo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

// Provider implements the Brevo email provider
type Provider struct {
	config *Config
	client *http.Client
}

// NewProvider creates a new Brevo provider
func NewProvider(config *Config) *Provider {
	return &Provider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "brevo"
}

// Send sends an email via Brevo API
func (p *Provider) Send(ctx context.Context, email *entity.Email) error {
	// Convert entity.Email to Brevo request
	request := p.buildRequest(email)
	
	// Marshal request
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	url := p.config.BaseURL + "/smtp/email"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", p.config.APIKey)
	
	// Send request
	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Handle response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	// Parse error response
	var errorResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		return fmt.Errorf("HTTP %d: failed to parse error response", resp.StatusCode)
	}
	
	return p.mapError(resp.StatusCode, &errorResp)
}

// IsHealthy checks if the provider is available
func (p *Provider) IsHealthy(ctx context.Context) error {
	url := p.config.BaseURL + "/account"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("api-key", p.config.APIKey)
	
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	
	return fmt.Errorf("health check failed: HTTP %d", resp.StatusCode)
}

// buildRequest converts entity.Email to Brevo SendRequest
func (p *Provider) buildRequest(email *entity.Email) *SendRequest {
	request := &SendRequest{
		Subject: email.Headers.Subject,
	}
	
	// Set sender
	if email.Headers.From != "" {
		request.Sender = Contact{Email: email.Headers.From}
	}
	
	// Set recipients
	for _, to := range email.Headers.To {
		request.To = append(request.To, Contact{Email: to})
	}
	
	for _, cc := range email.Headers.CC {
		request.CC = append(request.CC, Contact{Email: cc})
	}
	
	for _, bcc := range email.Headers.BCC {
		request.BCC = append(request.BCC, Contact{Email: bcc})
	}
	
	// Set content
	if email.HTMLBody != "" {
		request.HTMLContent = email.HTMLBody
	}
	if email.TextBody != "" {
		request.TextContent = email.TextBody
	}
	
	return request
}

// mapError maps Brevo API errors to standard errors
func (p *Provider) mapError(statusCode int, errorResp *ErrorResponse) error {
	message := errorResp.Message
	if message == "" {
		message = "unknown error"
	}
	
	switch statusCode {
	case 400:
		if strings.Contains(strings.ToLower(message), "invalid email") {
			return fmt.Errorf("invalid email address: %s", message)
		}
		return fmt.Errorf("bad request: %s", message)
	case 401:
		return fmt.Errorf("authentication failed: %s", message)
	case 402:
		return fmt.Errorf("insufficient credits: %s", message)
	case 403:
		return fmt.Errorf("forbidden: %s", message)
	case 429:
		return fmt.Errorf("rate limit exceeded: %s", message)
	case 500, 502, 503, 504:
		return fmt.Errorf("service unavailable: %s", message)
	default:
		return fmt.Errorf("API error %d: %s", statusCode, message)
	}
}
