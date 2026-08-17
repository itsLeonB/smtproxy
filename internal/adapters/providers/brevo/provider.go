package brevo

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/itsLeonB/smtproxy/internal/core/logger"
	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

const (
	// maxResponseBodySize caps how much of Brevo's response we'll buffer
	// into memory; Brevo responses are small JSON bodies, so anything
	// larger indicates a misbehaving/malicious server.
	maxResponseBodySize = 1 << 20 // 1MB
	// responseLogPreviewSize bounds how much of the response body is
	// included in debug logs.
	responseLogPreviewSize = 500
)

// previewBody returns a bounded preview of a response body suitable for logging.
func previewBody(body []byte) string {
	if len(body) <= responseLogPreviewSize {
		return string(body)
	}
	return string(body[:responseLogPreviewSize]) + "...(truncated)"
}

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

	logger.Debugf("brevo request built: to=%d cc=%d bcc=%d attachments=%d payload_bytes=%d",
		len(request.To), len(request.CC), len(request.BCC), len(request.Attachments), len(payload))

	// Create HTTP request
	url := p.config.BaseURL + "/smtp/email"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
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
	defer func() {
		if e := resp.Body.Close(); e != nil {
			logger.Error(e)
		}
	}()

	// Read response body once so it can be parsed regardless of status code.
	// Cap at maxResponseBodySize+1 to detect and reject oversized responses
	// without buffering unbounded data into memory.
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodySize+1))
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if len(respBody) > maxResponseBodySize {
		return fmt.Errorf("brevo response body exceeds maximum allowed size of %d bytes", maxResponseBodySize)
	}
	logger.Debugf("brevo response status=%d body_bytes=%d preview=%s", resp.StatusCode, len(respBody), previewBody(respBody))

	// Handle response
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// Parse error response
	var errorResp ErrorResponse
	if err := json.Unmarshal(respBody, &errorResp); err != nil {
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
	defer func() {
		if e := resp.Body.Close(); e != nil {
			logger.Error(e)
		}
	}()

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
	if email.Headers.From != nil && email.Headers.From.Address != "" {
		request.Sender = Contact{
			Email: email.Headers.From.Address,
			Name:  email.Headers.From.Name,
		}
	}

	// Set recipients
	for _, to := range email.Headers.To {
		if to != nil && to.Address != "" {
			request.To = append(request.To, Contact{
				Email: to.Address,
				Name:  to.Name,
			})
		}
	}

	for _, cc := range email.Headers.CC {
		if cc != nil && cc.Address != "" {
			request.CC = append(request.CC, Contact{
				Email: cc.Address,
				Name:  cc.Name,
			})
		}
	}

	for _, bcc := range email.Headers.BCC {
		if bcc != nil && bcc.Address != "" {
			request.BCC = append(request.BCC, Contact{
				Email: bcc.Address,
				Name:  bcc.Name,
			})
		}
	}

	// Set content
	if email.HTMLBody != "" {
		request.HTMLContent = email.HTMLBody
	}
	if email.TextBody != "" {
		request.TextContent = email.TextBody
	}

	// Set attachments
	for _, attachment := range email.Attachments {
		content, err := io.ReadAll(attachment.Content)
		if err != nil {
			continue // Skip invalid attachments
		}
		
		request.Attachments = append(request.Attachments, Attachment{
			Name:    attachment.Filename,
			Content: base64.StdEncoding.EncodeToString(content),
		})
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
