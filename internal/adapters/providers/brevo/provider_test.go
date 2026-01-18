package brevo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
	"github.com/stretchr/testify/assert"
)

func TestProvider_Name(t *testing.T) {
	config := &Config{APIKey: "test-key"}
	provider := NewProvider(config)

	assert.Equal(t, "brevo", provider.Name())
}

func TestProvider_Send_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/smtp/email", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("api-key"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"messageId": "test-message-id"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From:    &mail.Address{Address: "sender@example.com"},
			To:      []*mail.Address{{Address: "recipient@example.com"}},
			Subject: "Test Subject",
		},
		TextBody: "Test body",
		HTMLBody: "<p>Test body</p>",
	}

	err := provider.Send(context.Background(), email)
	assert.NoError(t, err)
}

func TestProvider_Send_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Invalid email address", "code": "invalid_parameter"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From:    &mail.Address{Address: "invalid-email"},
			To:      []*mail.Address{{Address: "recipient@example.com"}},
			Subject: "Test Subject",
		},
	}

	err := provider.Send(context.Background(), email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email address")
}

func TestProvider_Send_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message": "Invalid API key", "code": "unauthorized"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From:    &mail.Address{Address: "sender@example.com"},
			To:      []*mail.Address{{Address: "recipient@example.com"}},
			Subject: "Test Subject",
		},
	}

	err := provider.Send(context.Background(), email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")
}

func TestProvider_Send_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message": "Rate limit exceeded", "code": "rate_limit"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From:    &mail.Address{Address: "sender@example.com"},
			To:      []*mail.Address{{Address: "recipient@example.com"}},
			Subject: "Test Subject",
		},
	}

	err := provider.Send(context.Background(), email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "rate limit exceeded")
}

func TestProvider_Send_ServiceUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"message": "Service temporarily unavailable", "code": "service_unavailable"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From:    &mail.Address{Address: "sender@example.com"},
			To:      []*mail.Address{{Address: "recipient@example.com"}},
			Subject: "Test Subject",
		},
	}

	err := provider.Send(context.Background(), email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestProvider_IsHealthy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/account", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("api-key"))

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"email": "test@example.com"}`))
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	err := provider.IsHealthy(context.Background())
	assert.NoError(t, err)
}

func TestProvider_IsHealthy_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	config := &Config{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Timeout: 30 * time.Second,
	}
	provider := NewProvider(config)

	err := provider.IsHealthy(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "health check failed")
}

func TestProvider_BuildRequest(t *testing.T) {
	config := &Config{APIKey: "test-key"}
	provider := NewProvider(config)

	email := &entity.Email{
		Headers: entity.Headers{
			From: &mail.Address{Address: "ellionblessan@gmail.com", Name: "FOSS Sure"},
			To: []*mail.Address{
				{Address: "recipient1@example.com", Name: ""},
				{Address: "recipient2@example.com", Name: "John Doe"},
			},
			CC:      []*mail.Address{{Address: "cc@example.com", Name: ""}},
			BCC:     []*mail.Address{{Address: "bcc@example.com", Name: "Jane Smith"}},
			Subject: "Test Subject",
		},
		TextBody: "Plain text content",
		HTMLBody: "<p>HTML content</p>",
	}

	request := provider.buildRequest(email)

	assert.Equal(t, "Test Subject", request.Subject)

	// Sender with name and email
	assert.Equal(t, "ellionblessan@gmail.com", request.Sender.Email)
	assert.Equal(t, "FOSS Sure", request.Sender.Name)

	// Recipients
	assert.Len(t, request.To, 2)
	assert.Equal(t, "recipient1@example.com", request.To[0].Email)
	assert.Equal(t, "", request.To[0].Name)
	assert.Equal(t, "recipient2@example.com", request.To[1].Email)
	assert.Equal(t, "John Doe", request.To[1].Name)

	// CC
	assert.Len(t, request.CC, 1)
	assert.Equal(t, "cc@example.com", request.CC[0].Email)
	assert.Equal(t, "", request.CC[0].Name)

	// BCC with name
	assert.Len(t, request.BCC, 1)
	assert.Equal(t, "bcc@example.com", request.BCC[0].Email)
	assert.Equal(t, "Jane Smith", request.BCC[0].Name)

	assert.Equal(t, "Plain text content", request.TextContent)
	assert.Equal(t, "<p>HTML content</p>", request.HTMLContent)
}

func TestProvider_MapError(t *testing.T) {
	config := &Config{APIKey: "test-key"}
	provider := NewProvider(config)

	tests := []struct {
		statusCode int
		message    string
		expected   string
	}{
		{400, "Invalid email format", "invalid email address"},
		{401, "Invalid API key", "authentication failed"},
		{402, "Insufficient credits", "insufficient credits"},
		{403, "Access denied", "forbidden"},
		{429, "Too many requests", "rate limit exceeded"},
		{500, "Internal server error", "service unavailable"},
		{503, "Service unavailable", "service unavailable"},
	}

	for _, tt := range tests {
		errorResp := &ErrorResponse{Message: tt.message}
		err := provider.mapError(tt.statusCode, errorResp)

		assert.Error(t, err)
		assert.True(t, strings.Contains(strings.ToLower(err.Error()), tt.expected))
	}
}
