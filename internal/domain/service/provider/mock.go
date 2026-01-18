package provider

import (
	"context"
	"errors"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

// MockProvider is a test implementation of Provider
type MockProvider struct {
	name      string
	sendError error
	healthy   bool
}

// NewMockProvider creates a new mock provider
func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:    name,
		healthy: true,
	}
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	return m.name
}

// Send simulates sending an email
func (m *MockProvider) Send(ctx context.Context, email *entity.Email) error {
	return m.sendError
}

// IsHealthy returns the health status
func (m *MockProvider) IsHealthy(ctx context.Context) error {
	if !m.healthy {
		return errors.New("provider unhealthy")
	}
	return nil
}

// SetSendError sets the error to return on Send
func (m *MockProvider) SetSendError(err error) {
	m.sendError = err
}

// SetHealthy sets the health status
func (m *MockProvider) SetHealthy(healthy bool) {
	m.healthy = healthy
}
