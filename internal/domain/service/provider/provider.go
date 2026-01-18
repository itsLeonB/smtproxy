package provider

import (
	"context"

	"github.com/itsLeonB/smtproxy/internal/domain/entity"
)

// Provider defines the contract for transactional mail API providers
type Provider interface {
	// Name returns the provider identifier
	Name() string
	
	// Send sends an email via the provider's API
	Send(ctx context.Context, email *entity.Email) error
	
	// IsHealthy checks if the provider is available
	IsHealthy(ctx context.Context) error
}

// SendResult contains the result of a send operation
type SendResult struct {
	ProviderName string
	MessageID    string
	Error        error
}
